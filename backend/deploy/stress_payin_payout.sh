#!/usr/bin/env bash
# 代收 + 代付 OpenAPI 压力测试，结束后可选跑 DB 不变量检查并生成 Markdown 报告。
#
# 依赖: curl, python3, go（用于 check-pay-invariants）
#
# 主要环境变量:
#   GATEWAY_BASE_URL          默认 http://127.0.0.1:8080
#   MERCHANT_ID MERCHANT_SECRET
#   PAYOUT_PRODUCT_CODE       默认 bank_card
#   PAYIN_TYPE                默认 mock
#   PAYIN_AMOUNT PAYOUT_AMOUNT 默认 100（分，代付小金额减轻余额压力）
#   STRESS_ITERATIONS         总轮数（每轮随机代收或代付；默认 40000）。若 STRESS_DURATION_SEC>0 则忽略此项
#   STRESS_DURATION_SEC       >0 时按**墙钟时间**持续加压（秒），与 STRESS_CONCURRENCY 配合更像 soak
#   STRESS_CONCURRENCY        最大并行 worker 数（默认 20）
#   PAYIN_WEIGHT_PERCENT      代收占比 0-100（默认 55）
#   CHANNEL_SIGN_SECRETS      回调验签密钥候选，逗号分隔（与 test_collect_flow 一致）
#   STRESS_SKIP_INVARIANTS    设为 1 则跳过不变量与 go 检查
#   STRESS_SKIP_PAYIN_NOTIFY  设为 1 则代收只测创建+查单（不调上游回调，不验入账）
#   STRESS_REPORT             报告路径；默认 backend/deploy/stress-reports/report-时间戳.md
#
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
MERCHANT_ID="${MERCHANT_ID:-m_demo}"
MERCHANT_SECRET="${MERCHANT_SECRET:-demo_secret}"
PAYOUT_PRODUCT_CODE="${PAYOUT_PRODUCT_CODE:-bank_card}"
PAYIN_TYPE="${PAYIN_TYPE:-mock}"
PAYIN_AMOUNT="${PAYIN_AMOUNT:-100}"
PAYOUT_AMOUNT="${PAYOUT_AMOUNT:-100}"
STRESS_ITERATIONS="${STRESS_ITERATIONS:-40000}"
STRESS_DURATION_SEC="${STRESS_DURATION_SEC:-0}"
STRESS_CONCURRENCY="${STRESS_CONCURRENCY:-20}"
PAYIN_WEIGHT_PERCENT="${PAYIN_WEIGHT_PERCENT:-55}"
CHANNEL_SIGN_SECRET="${CHANNEL_SIGN_SECRET:-channel_secret}"
CHANNEL_SIGN_SECRETS="${CHANNEL_SIGN_SECRETS:-channel_secret,channel_secret_b,channel_secret_wechat,channel_secret_alipay}"
STRESS_SKIP_INVARIANTS="${STRESS_SKIP_INVARIANTS:-0}"
STRESS_SKIP_PAYIN_NOTIFY="${STRESS_SKIP_PAYIN_NOTIFY:-0}"

mkdir -p "${SCRIPT_DIR}/stress-reports"
TS="$(date +%Y%m%d-%H%M%S)"
STRESS_REPORT="${STRESS_REPORT:-${SCRIPT_DIR}/stress-reports/report-${TS}.md}"
RDIR="$(mktemp -d "${TMPDIR:-/tmp}/pay-stress.XXXXXX")"
RESULTS="${RDIR}/results.tsv"
export RESULTS RDIR

cleanup() { rm -rf "${RDIR}"; }
trap cleanup EXIT

if ! command -v curl >/dev/null 2>&1 || ! command -v python3 >/dev/null 2>&1; then
  echo "need curl and python3" >&2
  exit 1
fi

md5_sign() {
  local secret="$1"
  local kv_json="$2"
  python3 - "$secret" "$kv_json" <<'PY'
import hashlib, json, sys
secret = sys.argv[1]
params = json.loads(sys.argv[2])
items = []
for k in sorted(params.keys(), key=lambda x: x.lower()):
    if k.lower() == "sign":
        continue
    v = params[k]
    if v is None:
        continue
    s = str(v)
    if s == "":
        continue
    items.append((k.lower(), s))
base = "&".join(f"{k}={v}" for k, v in items)
if base:
    base += "&"
base += f"key={secret}"
print(hashlib.md5(base.encode("utf-8")).hexdigest())
PY
}

json_get() {
  python3 - "$1" "$2" <<'PY'
import json, sys
try:
    obj = json.loads(sys.argv[1])
except Exception:
    print(""); sys.exit(0)
path = sys.argv[2].split(".")
cur = obj
for p in path:
    if not isinstance(cur, dict) or p not in cur:
        print(""); sys.exit(0)
    cur = cur[p]
print(cur if cur is not None else "")
PY
}

signed_query_string() {
  local req_json="$1"
  local secret="$2"
  local with_nonce
  with_nonce="$(python3 - "${req_json}" <<'PY'
import json, sys, time, uuid
p = json.loads(sys.argv[1])
if "timestamp" not in p or not str(p["timestamp"]).strip():
    p["timestamp"] = str(int(time.time()))
if "nonce" not in p or not str(p["nonce"]).strip():
    p["nonce"] = uuid.uuid4().hex[:24]
print(json.dumps(p))
PY
)"
  local sign
  sign="$(md5_sign "${secret}" "${with_nonce}")"
  python3 - "${with_nonce}" "${sign}" <<'PY'
import json, sys, urllib.parse
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(urllib.parse.urlencode(p))
PY
}

notify_with_candidates() {
  local order_no="$1"
  local paid_amount="$2"
  local upstream_trade_no="$3"
  local channel_id="$4"
  local candidates_csv="$5"
  local old_ifs secret resp ok code last=""
  old_ifs="$IFS"
  IFS=','
  for secret in ${candidates_csv}; do
    secret="$(echo "${secret}" | tr -d ' ')"
    [[ -z "${secret}" ]] && continue
    params_json="$(python3 - <<PY
import json
print(json.dumps({
  "order_no": "${order_no}",
  "paid_amount": int("${paid_amount}"),
  "upstream_trade_no": "${upstream_trade_no}",
  "channel_id": int("${channel_id}")
}))
PY
)"
    sign="$(md5_sign "${secret}" "${params_json}")"
    body="$(python3 - "${params_json}" "${sign}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(json.dumps(p))
PY
)"
    tmpf="$(mktemp)"
    http_code="$(
      curl -sS -o "${tmpf}" -w "%{http_code}" -X POST "${GATEWAY_BASE_URL}/v1/callback/notify" \
        -H "Content-Type: application/json" -d "${body}" || echo "000"
    )"
    resp="$(cat "${tmpf}" 2>/dev/null || true)"
    rm -f "${tmpf}"
    last="${resp}"
    [[ "${http_code}" != "200" ]] && continue
    ok="$(json_get "${resp}" "ok" || true)"
    code="$(json_get "${resp}" "reason_code" || true)"
    if [[ "${ok}" == "True" || "${ok}" == "true" ]]; then
      IFS="$old_ifs"
      echo "${resp}"
      return 0
    fi
    if [[ "${code}" != "INVALID_SIGN" ]]; then
      IFS="$old_ifs"
      echo "${resp}"
      return 0
    fi
  done
  IFS="$old_ifs"
  echo "${last}"
  return 0
}

build_signed_payin_body() {
  local merchant_order_no="$1"
  local amount="$2"
  local payin_type="$3"
  local params
  params="$(python3 - <<PY
import json, time, uuid
print(json.dumps({
  "merchant_id": "${MERCHANT_ID}",
  "merchant_order_no": "${merchant_order_no}",
  "amount": int("${amount}"),
  "currency": "CNY",
  "payin_type": "${payin_type}",
  "notify_url": "",
  "timestamp": int(time.time()),
  "nonce": uuid.uuid4().hex[:24]
}))
PY
)"
  local sign
  sign="$(md5_sign "${MERCHANT_SECRET}" "${params}")"
  python3 - "${params}" "${sign}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(json.dumps(p))
PY
}

build_signed_payout_body() {
  local merchant_order_no="$1"
  local amount="$2"
  local product_code="$3"
  local params
  params="$(python3 - <<PY
import json, time, uuid
print(json.dumps({
  "merchant_id": "${MERCHANT_ID}",
  "merchant_order_no": "${merchant_order_no}",
  "amount": int("${amount}"),
  "currency": "CNY",
  "payout_product_code": "${product_code}",
  "timestamp": int(time.time()),
  "nonce": uuid.uuid4().hex[:24]
}))
PY
)"
  local sign
  sign="$(md5_sign "${MERCHANT_SECRET}" "${params}")"
  python3 - "${params}" "${sign}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(json.dumps(p))
PY
}

curl_json_metrics() {
  local tmpf="$1"
  shift
  curl -sS -o "${tmpf}" -w "%{http_code}|%{time_total}" "$@" || echo "000|0"
}

append_result() {
  # $1=op $2=stage $3=ok|FAIL $4=http $5=time $6=extra
  printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$4" "$5" "$6" >>"${RESULTS}"
}

worker_payin() {
  local id="$1"
  local mon="PI-S-${TS}-${id}-$RANDOM$RANDOM"
  local tmpf body metrics http t api_code

  tmpf="$(mktemp)"
  body="$(build_signed_payin_body "${mon}" "${PAYIN_AMOUNT}" "${PAYIN_TYPE}")"
  metrics="$(curl -sS -o "${tmpf}" -w "%{http_code}|%{time_total}" \
    -X POST "${GATEWAY_BASE_URL}/v1/payin/order" \
    -H "Content-Type: application/json" -d "${body}")" || metrics="000|0"
  http="${metrics%%|*}"
  t="${metrics##*|}"
  body="$(cat "${tmpf}")"
  rm -f "${tmpf}"

  api_code="$(json_get "${body}" "code")"
  if [[ "${http}" != "200" ]]; then
    append_result "PAYIN" "create" "FAIL" "${http}" "${t}" "${api_code:-HTTP}"
    return
  fi
  local order_no channel_id
  order_no="$(json_get "${body}" "order_no")"
  channel_id="$(json_get "${body}" "channel_id")"
  if [[ -z "${order_no}" || -z "${channel_id}" || "${channel_id}" == "0" ]]; then
    append_result "PAYIN" "create" "FAIL" "${http}" "${t}" "no_order"
    return
  fi
  append_result "PAYIN" "create" "OK" "${http}" "${t}" "-"

  if [[ "${STRESS_SKIP_PAYIN_NOTIFY}" == "1" ]]; then
    req_json="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","order_no":"${order_no}","timestamp":str(int(time.time()))}))
PY
)"
    qs="$(signed_query_string "${req_json}" "${MERCHANT_SECRET}")"
    tmpf="$(mktemp)"
    metrics="$(curl_json_metrics "${tmpf}" "${GATEWAY_BASE_URL}/v1/payin/query?${qs}")"
    http="${metrics%%|*}"
    t="${metrics##*|}"
    body="$(cat "${tmpf}")"
    rm -f "${tmpf}"
    api_code="$(json_get "${body}" "code")"
    [[ "${http}" == "200" && -z "${api_code}" ]] && append_result "PAYIN" "query" "OK" "${http}" "${t}" "-" || append_result "PAYIN" "query" "FAIL" "${http}" "${t}" "${api_code:-bad}"
    return
  fi

  local upstream="UP-S-${id}-$RANDOM"
  local nresp nok
  nresp="$(notify_with_candidates "${order_no}" "${PAYIN_AMOUNT}" "${upstream}" "${channel_id}" "${CHANNEL_SIGN_SECRET},${CHANNEL_SIGN_SECRETS}")"
  nok="$(json_get "${nresp}" "ok" || true)"
  if [[ "${nok}" != "True" && "${nok}" != "true" ]]; then
    append_result "PAYIN" "notify" "FAIL" "200" "0" "$(json_get "${nresp}" "reason_code")"
    return
  fi
  append_result "PAYIN" "notify" "OK" "200" "0" "-"

  req_json="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","order_no":"${order_no}","timestamp":str(int(time.time()))}))
PY
)"
  qs="$(signed_query_string "${req_json}" "${MERCHANT_SECRET}")"
  tmpf="$(mktemp)"
  metrics="$(curl_json_metrics "${tmpf}" "${GATEWAY_BASE_URL}/v1/payin/query?${qs}")"
  http="${metrics%%|*}"
  t="${metrics##*|}"
  body="$(cat "${tmpf}")"
  rm -f "${tmpf}"
  api_code="$(json_get "${body}" "code")"
  local st
  st="$(json_get "${body}" "order.status")"
  if [[ "${http}" == "200" && -z "${api_code}" && "${st}" == "1" ]]; then
    append_result "PAYIN" "query" "OK" "${http}" "${t}" "-"
  else
    append_result "PAYIN" "query" "FAIL" "${http}" "${t}" "status=${st}:${api_code}"
  fi
}

worker_payout() {
  local id="$1"
  local mon="PO-S-${TS}-${id}-$RANDOM$RANDOM"
  local tmpf body metrics http t api_code order_no

  tmpf="$(mktemp)"
  body="$(build_signed_payout_body "${mon}" "${PAYOUT_AMOUNT}" "${PAYOUT_PRODUCT_CODE}")"
  metrics="$(curl -sS -o "${tmpf}" -w "%{http_code}|%{time_total}" \
    -X POST "${GATEWAY_BASE_URL}/v1/payout/order" \
    -H "Content-Type: application/json" -d "${body}")" || metrics="000|0"
  http="${metrics%%|*}"
  t="${metrics##*|}"
  body="$(cat "${tmpf}")"
  rm -f "${tmpf}"

  api_code="$(json_get "${body}" "code")"
  if [[ "${http}" != "200" ]]; then
    append_result "PAYOUT" "create" "FAIL" "${http}" "${t}" "${api_code:-HTTP}"
    return
  fi
  if [[ -n "${api_code}" ]]; then
    append_result "PAYOUT" "create" "FAIL" "${http}" "${t}" "${api_code}"
    return
  fi
  order_no="$(json_get "${body}" "order_no")"
  if [[ -z "${order_no}" ]]; then
    append_result "PAYOUT" "create" "FAIL" "${http}" "${t}" "no_order"
    return
  fi
  append_result "PAYOUT" "create" "OK" "${http}" "${t}" "-"

  req_json="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","order_no":"${order_no}","timestamp":str(int(time.time()))}))
PY
)"
  qs="$(signed_query_string "${req_json}" "${MERCHANT_SECRET}")"
  tmpf="$(mktemp)"
  metrics="$(curl_json_metrics "${tmpf}" "${GATEWAY_BASE_URL}/v1/payout/query?${qs}")"
  http="${metrics%%|*}"
  t="${metrics##*|}"
  body="$(cat "${tmpf}")"
  rm -f "${tmpf}"
  api_code="$(json_get "${body}" "code")"
  if [[ "${http}" == "200" && -z "${api_code}" ]]; then
    append_result "PAYOUT" "query" "OK" "${http}" "${t}" "-"
  else
    append_result "PAYOUT" "query" "FAIL" "${http}" "${t}" "${api_code:-bad}"
  fi
}

worker_mixed() {
  local id="$1"
  local r=$((RANDOM % 100))
  if [[ "${r}" -lt "${PAYIN_WEIGHT_PERCENT}" ]]; then
    worker_payin "${id}"
  else
    worker_payout "${id}"
  fi
}

health_check() {
  local code
  code="$(curl -sS -o /dev/null -w "%{http_code}" "${GATEWAY_BASE_URL}/health" || echo "000")"
  [[ "${code}" == "200" ]]
}

echo "==> 健康检查 ${GATEWAY_BASE_URL}/health"
if ! health_check; then
  echo "网关 /health 非 200，请确认服务已启动" >&2
  exit 1
fi

started_at="$(date "+%Y-%m-%dT%H:%M:%S%z")"
t0="$(date +%s)"
: >"${RESULTS}"

ACTUAL_ROUNDS=0
if [[ "${STRESS_DURATION_SEC}" =~ ^[0-9]+$ ]] && [[ "${STRESS_DURATION_SEC}" -gt 0 ]]; then
  echo "==> 压力参数 mode=duration sec=${STRESS_DURATION_SEC} concurrency=${STRESS_CONCURRENCY} payin_weight=${PAYIN_WEIGHT_PERCENT}%"
  iter=0
  while (( $(date +%s) - t0 < STRESS_DURATION_SEC )); do
    while [[ "$(jobs -r | wc -l | tr -d ' ')" -ge "${STRESS_CONCURRENCY}" ]]; do
      sleep 0.05
    done
    iter=$((iter + 1))
    worker_mixed "${iter}" &
  done
  ACTUAL_ROUNDS="${iter}"
else
  echo "==> 压力参数 mode=rounds iterations=${STRESS_ITERATIONS} concurrency=${STRESS_CONCURRENCY} payin_weight=${PAYIN_WEIGHT_PERCENT}%"
  iter=1
  while [[ "${iter}" -le "${STRESS_ITERATIONS}" ]]; do
    while [[ "$(jobs -r | wc -l | tr -d ' ')" -ge "${STRESS_CONCURRENCY}" ]]; do
      sleep 0.05
    done
    worker_mixed "${iter}" &
    iter=$((iter + 1))
  done
  ACTUAL_ROUNDS="${STRESS_ITERATIONS}"
fi
echo "    结果 TSV -> ${RESULTS}"
wait

ended_at="$(date "+%Y-%m-%dT%H:%M:%S%z")"
t1="$(date +%s)"
elapsed=$((t1 - t0))

INV_STDOUT="${RDIR}/invariants.out"
INV_RC=0
if [[ "${STRESS_SKIP_INVARIANTS}" != "1" ]]; then
  echo "==> DB 不变量检查 (check-pay-invariants)"
  bash "${SCRIPT_DIR}/check_payin_payout_invariants.sh" >"${INV_STDOUT}" 2>&1
  INV_RC=$?
else
  echo "(跳过不变量 STRESS_SKIP_INVARIANTS=1)" >"${INV_STDOUT}"
fi

# ---------- 报告：聚合 ----------
python3 "${SCRIPT_DIR}/stress_report_gen.py" \
  "${RESULTS}" "${started_at}" "${ended_at}" "${elapsed}" "${INV_RC}" \
  "${GATEWAY_BASE_URL}" "${MERCHANT_ID}" "${STRESS_ITERATIONS}" "${STRESS_DURATION_SEC}" "${ACTUAL_ROUNDS}" \
  "${STRESS_CONCURRENCY}" "${PAYIN_WEIGHT_PERCENT}" "${PAYIN_AMOUNT}" "${PAYOUT_AMOUNT}" "${PAYIN_TYPE}" \
  "${STRESS_SKIP_PAYIN_NOTIFY}" "${STRESS_SKIP_INVARIANTS}" "${INV_STDOUT}" "${STRESS_REPORT}"

echo "==> 报告: ${STRESS_REPORT}"

stress_failures="$(awk -F'\t' '$3 == "FAIL" { c++ } END { print c+0 }' "${RESULTS}" 2>/dev/null || echo 0)"
exit_code=0
if [[ "${stress_failures}" -gt 0 ]]; then
  echo "==> 压力阶段失败行数: ${stress_failures}（见报告）" >&2
  exit_code=1
fi
if [[ "${STRESS_SKIP_INVARIANTS}" != "1" && "${INV_RC}" -ne 0 ]]; then
  echo "==> DB 不变量检查未通过，退出码 ${INV_RC}" >&2
  exit_code=1
fi
exit "${exit_code}"

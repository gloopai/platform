#!/usr/bin/env bash
set -euo pipefail

# 收款主链路端到端用例（MVP）：
# 1) 创建订单
# 2) 模拟上游回调置成功
# 3) 查单校验状态
# 4) 通过管理台接口校验订单侧已落账
# 5) 通过管理台接口校验资金流水已产生

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
MERCHANT_ID="${MERCHANT_ID:-m_demo}"
MERCHANT_SECRET="${MERCHANT_SECRET:-demo_secret}"
PAY_TYPE="${PAY_TYPE:-mock}"
AMOUNT="${AMOUNT:-1234}"

CHANNEL_SIGN_SECRET="${CHANNEL_SIGN_SECRET:-channel_secret}"
CHANNEL_SIGN_SECRETS="${CHANNEL_SIGN_SECRETS:-channel_secret,channel_secret_b,channel_secret_wechat,channel_secret_alipay}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"
DEFAULT_TEST_CASES_JSON='[
  {"name":"default-rate-mock","merchant_id":"m_demo","merchant_secret":"demo_secret","pay_type":"mock","amount":1234,"expected_fee_rate_bps":60},
  {"name":"product-rate-wechat","merchant_id":"m_rate_mix","merchant_secret":"demo_secret_mix","pay_type":"wechat","amount":2000,"expected_fee_rate_bps":120},
  {"name":"zero-rate-alipay","merchant_id":"m_zero_fee","merchant_secret":"demo_secret_zero","pay_type":"alipay","amount":1500,"expected_fee_rate_bps":0}
]'
TEST_CASES_JSON="${TEST_CASES_JSON:-$DEFAULT_TEST_CASES_JSON}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl not found"
  exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 not found"
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
  local json="$1"
  local path="$2"
  python3 - "$json" "$path" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
path = sys.argv[2].split(".")
cur = obj
for p in path:
    if not isinstance(cur, dict) or p not in cur:
        print("")
        sys.exit(0)
    cur = cur[p]
print(cur if cur is not None else "")
PY
}

notify_request() {
  local order_no="$1"
  local paid_amount="$2"
  local upstream_trade_no="$3"
  local channel_id="$4"
  local sign_secret="$5"
  local body
  local params_json
  local sign

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
  sign="$(md5_sign "${sign_secret}" "${params_json}")"
  body="$(python3 - "${params_json}" "${sign}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(json.dumps(p))
PY
)"
  curl -sS -X POST "${GATEWAY_BASE_URL}/v1/callback/notify" \
    -H "Content-Type: application/json" \
    -d "${body}"
}

assert_notify() {
  local name="$1"
  local resp="$2"
  local expected_ok="$3"
  local expected_code="$4"
  local actual_ok
  local actual_code

  actual_ok="$(json_get "${resp}" "ok" || true)"
  actual_code="$(json_get "${resp}" "reason_code" || true)"

  if [[ "${actual_ok}" != "${expected_ok}" ]]; then
    echo "${name} expected ok=${expected_ok}, got ok=${actual_ok}: ${resp}"
    exit 1
  fi
  if [[ "${expected_code}" != "" && "${actual_code}" != "${expected_code}" ]]; then
    echo "${name} expected reason_code=${expected_code}, got ${actual_code}: ${resp}"
    exit 1
  fi
  echo "  ${name} ok=${actual_ok} reason_code=${actual_code}"
}

notify_with_candidates() {
  local order_no="$1"
  local paid_amount="$2"
  local upstream_trade_no="$3"
  local channel_id="$4"
  local candidates_csv="$5"
  local old_ifs
  local secret
  local resp
  local ok
  local code

  old_ifs="$IFS"
  IFS=','
  for secret in ${candidates_csv}; do
    secret="$(echo "${secret}" | tr -d ' ')"
    if [[ -z "${secret}" ]]; then
      continue
    fi
    resp="$(notify_request "${order_no}" "${paid_amount}" "${upstream_trade_no}" "${channel_id}" "${secret}")"
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
  # 全部候选都签名失败，返回最后一次响应
  echo "${resp}"
  return 0
}

echo "[0/2] 管理台登录..."
ADMIN_LOGIN_BODY="$(python3 - <<PY
import json
print(json.dumps({
  "username": "${ADMIN_USERNAME}",
  "password": "${ADMIN_PASSWORD}"
}))
PY
)"
ADMIN_LOGIN_RESP="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/admin/login" \
  -H "Content-Type: application/json" \
  -d "${ADMIN_LOGIN_BODY}")"
ADMIN_TOKEN="$(json_get "${ADMIN_LOGIN_RESP}" "token" || true)"
if [[ -z "${ADMIN_TOKEN}" ]]; then
  echo "admin login failed: ${ADMIN_LOGIN_RESP}"
  exit 1
fi

list_cases() {
  python3 - "${TEST_CASES_JSON}" "${DEFAULT_TEST_CASES_JSON}" <<'PY'
import json, sys
raw = sys.argv[1]
fallback = sys.argv[2]
try:
    cases = json.loads(raw)
except Exception:
    try:
        cases = json.loads(fallback)
    except Exception as e:
        print(f"invalid TEST_CASES_JSON: {e}", file=sys.stderr)
        sys.exit(1)
for c in cases:
    print(
      f"{c.get('name','case')}\t{c['merchant_id']}\t{c['merchant_secret']}\t{c['pay_type']}\t{int(c['amount'])}\t{int(c.get('expected_fee_rate_bps', 0))}"
    )
PY
}

verify_admin_order() {
  local admin_orders_resp="$1"
  local order_no="$2"
  local amount="$3"
  local upstream_trade_no="$4"
  local channel_id="$5"
  local expect_fee_rate_bps="$6"
  python3 - "${admin_orders_resp}" "${order_no}" "${amount}" "${upstream_trade_no}" "${channel_id}" "${expect_fee_rate_bps}" <<'PY'
import json, sys
resp = json.loads(sys.argv[1])
order_no = sys.argv[2]
amount = int(sys.argv[3])
upstream_trade_no = sys.argv[4]
channel_id = int(sys.argv[5])
expect_fee_rate_bps = int(sys.argv[6])
orders = resp.get("orders") or []
target = next((o for o in orders if o.get("order_no") == order_no), None)
if target is None:
    print(f"admin orders missing target order: {resp}")
    sys.exit(1)
if int(target.get("status", -1)) != 1:
    print(f"admin orders status expected 1, got {target.get('status')}: {target}")
    sys.exit(1)
if int(target.get("paid_amount", -1)) != amount:
    print(f"admin orders paid_amount expected {amount}, got {target.get('paid_amount')}: {target}")
    sys.exit(1)
if str(target.get("upstream_trade_no", "")) != upstream_trade_no:
    print(f"admin orders upstream_trade_no mismatch: {target}")
    sys.exit(1)
if int(target.get("channel_id", 0)) != channel_id:
    print(f"admin orders channel_id mismatch: {target}")
    sys.exit(1)
fee_rate_bps = int(target.get("fee_rate_bps", -1))
fee_amount = int(target.get("fee_amount", -1))
net_amount = int(target.get("net_amount", -1))
expected_fee = amount * expect_fee_rate_bps // 10000
expected_net = amount - expected_fee
if fee_rate_bps != expect_fee_rate_bps:
    print(f"admin orders fee_rate_bps expected {expect_fee_rate_bps}, got {fee_rate_bps}: {target}")
    sys.exit(1)
if fee_amount != expected_fee:
    print(f"admin orders fee_amount expected {expected_fee}, got {fee_amount}: {target}")
    sys.exit(1)
if net_amount != expected_net:
    print(f"admin orders net_amount expected {expected_net}, got {net_amount}: {target}")
    sys.exit(1)
print(f"admin orders view ok (fee_rate_bps={fee_rate_bps}, fee_amount={fee_amount}, net_amount={net_amount})")
PY
}

verify_admin_settlement() {
  local admin_settlement_resp="$1"
  local order_no="$2"
  python3 - "${admin_settlement_resp}" "${order_no}" <<'PY'
import json, sys
resp = json.loads(sys.argv[1])
order_no = sys.argv[2]
logs = resp.get("logs") or []
ok = any((l.get("order_no") == order_no and l.get("change_type") == "ORDER_PAID") for l in logs)
if not ok:
    print(f"admin settlement logs missing ORDER_PAID for order: {resp}")
    sys.exit(1)
print("admin settlement logs ok")
PY
}

run_case() {
  local case_name="$1"
  local merchant_id="$2"
  local merchant_secret="$3"
  local pay_type="$4"
  local amount="$5"
  local expect_fee_rate_bps="$6"

  echo
  echo "========== CASE: ${case_name} =========="
  echo "[1/6] 创建订单..."
  local merchant_order_no="MO-E2E-${case_name}-$(date +%s)"
  local create_params_json
  create_params_json="$(python3 - <<PY
import json
print(json.dumps({
  "merchant_id": "${merchant_id}",
  "merchant_order_no": "${merchant_order_no}",
  "amount": int("${amount}"),
  "currency": "CNY",
  "pay_type": "${pay_type}",
  "notify_url": ""
}))
PY
)"
  local create_sign
  create_sign="$(md5_sign "${merchant_secret}" "${create_params_json}")"
  local create_body
  create_body="$(python3 - "${create_params_json}" "${create_sign}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(json.dumps(p))
PY
)"
  local create_resp
  create_resp="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/pay/order" -H "Content-Type: application/json" -d "${create_body}")"
  local order_no
  local channel_id
  order_no="$(json_get "${create_resp}" "order_no")"
  channel_id="$(json_get "${create_resp}" "channel_id")"
  if [[ -z "${order_no}" ]]; then
    echo "create order failed: ${create_resp}"
    exit 1
  fi
  if [[ -z "${channel_id}" || "${channel_id}" == "0" ]]; then
    echo "create order returned invalid channel_id: ${create_resp}"
    exit 1
  fi
  echo "  order_no=${order_no} channel_id=${channel_id}"

  echo "[2/6] 模拟上游回调..."
  local upstream_trade_no="UP-E2E-${case_name}-$(date +%s)"
  local notify_resp
  notify_resp="$(notify_with_candidates "${order_no}" "${amount}" "${upstream_trade_no}" "${channel_id}" "${CHANNEL_SIGN_SECRET},${CHANNEL_SIGN_SECRETS}")"
  assert_notify "notify_paid_first" "${notify_resp}" "True" ""

  echo "[3/6] 查单校验金额与手续费快照..."
  local query_params_json
  query_params_json="$(python3 - <<PY
import json, time
print(json.dumps({
  "merchant_id": "${merchant_id}",
  "order_no": "${order_no}",
  "timestamp": str(int(time.time()))
}))
PY
)"
  local query_sign
  query_sign="$(md5_sign "${merchant_secret}" "${query_params_json}")"
  local query_url
  query_url="$(python3 - "${query_params_json}" "${query_sign}" <<'PY'
import json, sys, urllib.parse
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(urllib.parse.urlencode(p))
PY
)"
  local query_resp
  query_resp="$(curl -sS "${GATEWAY_BASE_URL}/v1/pay/query?${query_url}")"
  local query_status
  local query_paid_amount
  local query_fee_rate_bps
  local query_fee_amount
  local query_net_amount
  query_status="$(json_get "${query_resp}" "order.status")"
  query_paid_amount="$(json_get "${query_resp}" "order.paid_amount")"
  query_fee_rate_bps="$(json_get "${query_resp}" "order.fee_rate_bps")"
  query_fee_amount="$(json_get "${query_resp}" "order.fee_amount")"
  query_net_amount="$(json_get "${query_resp}" "order.net_amount")"
  local expected_fee_amount=$(( amount * expect_fee_rate_bps / 10000 ))
  local expected_net_amount=$(( amount - expected_fee_amount ))
  if [[ "${query_status}" != "1" ]]; then
    echo "query order status expected 1, got ${query_status}: ${query_resp}"
    exit 1
  fi
  if [[ "${query_paid_amount}" != "${amount}" ]]; then
    echo "query order paid_amount expected ${amount}, got ${query_paid_amount}: ${query_resp}"
    exit 1
  fi
  if [[ "${query_fee_rate_bps}" != "${expect_fee_rate_bps}" ]]; then
    echo "query order fee_rate_bps expected ${expect_fee_rate_bps}, got ${query_fee_rate_bps}: ${query_resp}"
    exit 1
  fi
  if [[ "${query_fee_amount}" != "${expected_fee_amount}" ]]; then
    echo "query order fee_amount expected ${expected_fee_amount}, got ${query_fee_amount}: ${query_resp}"
    exit 1
  fi
  if [[ "${query_net_amount}" != "${expected_net_amount}" ]]; then
    echo "query order net_amount expected ${expected_net_amount}, got ${query_net_amount}: ${query_resp}"
    exit 1
  fi
  echo "  order status=1 fee_rate_bps=${query_fee_rate_bps} fee_amount=${query_fee_amount} net_amount=${query_net_amount}"

  echo "[4/6] 管理台接口校验订单视图..."
  local admin_orders_resp
  admin_orders_resp="$(curl -sS "${GATEWAY_BASE_URL}/v1/admin/orders?merchant_id=${merchant_id}&keyword=${order_no}&limit=20" -H "X-Admin-Token: ${ADMIN_TOKEN}")"
  verify_admin_order "${admin_orders_resp}" "${order_no}" "${amount}" "${upstream_trade_no}" "${channel_id}" "${expect_fee_rate_bps}"

  echo "[5/6] 管理台接口校验资金流水..."
  local admin_settlement_resp
  admin_settlement_resp="$(curl -sS "${GATEWAY_BASE_URL}/v1/admin/settlement/logs?merchant_id=${merchant_id}&limit=100" -H "X-Admin-Token: ${ADMIN_TOKEN}")"
  verify_admin_settlement "${admin_settlement_resp}" "${order_no}"

  echo "[6/6] 回调状态矩阵..."
  local replay_ok_resp
  replay_ok_resp="$(notify_with_candidates "${order_no}" "${amount}" "${upstream_trade_no}" "${channel_id}" "${CHANNEL_SIGN_SECRET},${CHANNEL_SIGN_SECRETS}")"
  assert_notify "notify_idempotent_replay" "${replay_ok_resp}" "True" "IDEMPOTENT_REPLAY_ACCEPTED"

  local replay_mismatch_resp
  replay_mismatch_resp="$(notify_with_candidates "${order_no}" "${amount}" "UP-E2E-MISMATCH-$(date +%s)" "${channel_id}" "${CHANNEL_SIGN_SECRET},${CHANNEL_SIGN_SECRETS}")"
  assert_notify "notify_replay_mismatch" "${replay_mismatch_resp}" "False" "REPLAY_PAYLOAD_MISMATCH"

  local invalid_sign_resp
  invalid_sign_resp="$(notify_request "${order_no}" "${amount}" "UP-E2E-BAD-SIGN-$(date +%s)" "${channel_id}" "wrong_secret")"
  assert_notify "notify_invalid_sign" "${invalid_sign_resp}" "False" "INVALID_SIGN"

  local not_found_resp
  not_found_resp="$(notify_with_candidates "P-NOT-FOUND-$(date +%s)" "${amount}" "UP-E2E-NOTFOUND-$(date +%s)" "${channel_id}" "${CHANNEL_SIGN_SECRET},${CHANNEL_SIGN_SECRETS}")"
  assert_notify "notify_order_not_found" "${not_found_resp}" "False" "ORDER_NOT_FOUND"

  local invalid_params_resp
  invalid_params_resp="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/callback/notify" \
    -H "Content-Type: application/json" \
    -d '{"order_no":"","paid_amount":0,"upstream_trade_no":"","channel_id":0,"sign":"x"}')"
  assert_notify "notify_invalid_params" "${invalid_params_resp}" "False" "INVALID_NOTIFY_PARAMS"

  echo "  case=${case_name} done"
}

echo
echo "[1/2] 执行多测试样例..."
case_lines="$(list_cases)"
if [[ -z "${case_lines}" ]]; then
  echo "no valid test cases parsed from TEST_CASES_JSON"
  exit 1
fi
executed_cases=0
while IFS=$'\t' read -r case_name case_merchant_id case_merchant_secret case_pay_type case_amount case_expected_fee_rate_bps; do
  if [[ -z "${case_name}" ]]; then
    continue
  fi
  executed_cases=$((executed_cases + 1))
  run_case "${case_name}" "${case_merchant_id}" "${case_merchant_secret}" "${case_pay_type}" "${case_amount}" "${case_expected_fee_rate_bps}"
done <<< "${case_lines}"
if [[ "${executed_cases}" -le 0 ]]; then
  echo "no test case executed"
  exit 1
fi

echo
echo "PASS: 多商户/多产品/多费率收款主链路与回调状态矩阵验证成功"

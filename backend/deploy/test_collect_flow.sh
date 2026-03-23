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
CHANNEL_SIGN_SECRETS="${CHANNEL_SIGN_SECRETS:-channel_secret,channel_secret_b}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"

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

echo "[1/5] 创建订单..."
MERCHANT_ORDER_NO="MO-E2E-$(date +%s)"
CREATE_PARAMS_JSON="$(python3 - <<PY
import json
print(json.dumps({
  "merchant_id": "${MERCHANT_ID}",
  "merchant_order_no": "${MERCHANT_ORDER_NO}",
  "amount": int("${AMOUNT}"),
  "currency": "CNY",
  "pay_type": "${PAY_TYPE}",
  "notify_url": ""
}))
PY
)"
CREATE_SIGN="$(md5_sign "${MERCHANT_SECRET}" "${CREATE_PARAMS_JSON}")"
CREATE_BODY="$(python3 - "${CREATE_PARAMS_JSON}" "${CREATE_SIGN}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(json.dumps(p))
PY
)"

CREATE_RESP="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/pay/order" \
  -H "Content-Type: application/json" \
  -d "${CREATE_BODY}")"

ORDER_NO="$(json_get "${CREATE_RESP}" "order_no")"
CHANNEL_ID="$(json_get "${CREATE_RESP}" "channel_id")"
if [[ -z "${ORDER_NO}" ]]; then
  echo "create order failed: ${CREATE_RESP}"
  exit 1
fi
if [[ -z "${CHANNEL_ID}" || "${CHANNEL_ID}" == "0" ]]; then
  echo "create order returned invalid channel_id: ${CREATE_RESP}"
  exit 1
fi
echo "  order_no=${ORDER_NO} channel_id=${CHANNEL_ID}"

echo "[2/5] 模拟上游回调..."
UPSTREAM_TRADE_NO="UP-E2E-$(date +%s)"
NOTIFY_RESP="$(notify_with_candidates "${ORDER_NO}" "${AMOUNT}" "${UPSTREAM_TRADE_NO}" "${CHANNEL_ID}" "${CHANNEL_SIGN_SECRET},${CHANNEL_SIGN_SECRETS}")"
assert_notify "notify_paid_first" "${NOTIFY_RESP}" "True" ""

echo "[3/5] 查单校验..."
QUERY_PARAMS_JSON="$(python3 - <<PY
import json, time
print(json.dumps({
  "merchant_id": "${MERCHANT_ID}",
  "order_no": "${ORDER_NO}",
  "timestamp": str(int(time.time()))
}))
PY
)"
QUERY_SIGN="$(md5_sign "${MERCHANT_SECRET}" "${QUERY_PARAMS_JSON}")"
QUERY_URL="$(python3 - "${QUERY_PARAMS_JSON}" "${QUERY_SIGN}" <<'PY'
import json, sys, urllib.parse
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(urllib.parse.urlencode(p))
PY
)"
QUERY_RESP="$(curl -sS "${GATEWAY_BASE_URL}/v1/pay/query?${QUERY_URL}")"
QUERY_STATUS="$(json_get "${QUERY_RESP}" "order.status")"
QUERY_PAID_AMOUNT="$(json_get "${QUERY_RESP}" "order.paid_amount")"
if [[ "${QUERY_STATUS}" != "1" ]]; then
  echo "query order status expected 1, got ${QUERY_STATUS}: ${QUERY_RESP}"
  exit 1
fi
if [[ "${QUERY_PAID_AMOUNT}" != "${AMOUNT}" ]]; then
  echo "query order paid_amount expected ${AMOUNT}, got ${QUERY_PAID_AMOUNT}: ${QUERY_RESP}"
  exit 1
fi
echo "  order status=1 paid_amount=${QUERY_PAID_AMOUNT}"

echo "[4/5] 管理台接口校验订单视图..."
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

ADMIN_ORDERS_RESP="$(curl -sS "${GATEWAY_BASE_URL}/v1/admin/orders?merchant_id=${MERCHANT_ID}&keyword=${ORDER_NO}&limit=20" \
  -H "X-Admin-Token: ${ADMIN_TOKEN}")"
python3 - "${ADMIN_ORDERS_RESP}" "${ORDER_NO}" "${AMOUNT}" "${UPSTREAM_TRADE_NO}" "${CHANNEL_ID}" <<'PY'
import json, sys
resp = json.loads(sys.argv[1])
order_no = sys.argv[2]
amount = int(sys.argv[3])
upstream_trade_no = sys.argv[4]
channel_id = int(sys.argv[5])
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
print("admin orders view ok")
PY
echo "  admin orders view ok"

echo "[5/5] 管理台接口校验资金流水..."
ADMIN_SETTLEMENT_RESP="$(curl -sS "${GATEWAY_BASE_URL}/v1/admin/settlement/logs?merchant_id=${MERCHANT_ID}&limit=100" \
  -H "X-Admin-Token: ${ADMIN_TOKEN}")"
python3 - "${ADMIN_SETTLEMENT_RESP}" "${ORDER_NO}" <<'PY'
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
echo "  admin settlement logs ok"

echo "[6/6] 回调状态矩阵..."
REPLAY_OK_RESP="$(notify_request "${ORDER_NO}" "${AMOUNT}" "${UPSTREAM_TRADE_NO}" "${CHANNEL_ID}" "${CHANNEL_SIGN_SECRET}")"
assert_notify "notify_idempotent_replay" "${REPLAY_OK_RESP}" "True" "IDEMPOTENT_REPLAY_ACCEPTED"

REPLAY_MISMATCH_RESP="$(notify_request "${ORDER_NO}" "${AMOUNT}" "UP-E2E-MISMATCH-$(date +%s)" "${CHANNEL_ID}" "${CHANNEL_SIGN_SECRET}")"
assert_notify "notify_replay_mismatch" "${REPLAY_MISMATCH_RESP}" "False" "REPLAY_PAYLOAD_MISMATCH"

INVALID_SIGN_RESP="$(notify_request "${ORDER_NO}" "${AMOUNT}" "UP-E2E-BAD-SIGN-$(date +%s)" "${CHANNEL_ID}" "wrong_secret")"
assert_notify "notify_invalid_sign" "${INVALID_SIGN_RESP}" "False" "INVALID_SIGN"

NOT_FOUND_RESP="$(notify_request "P-NOT-FOUND-$(date +%s)" "${AMOUNT}" "UP-E2E-NOTFOUND-$(date +%s)" "${CHANNEL_ID}" "${CHANNEL_SIGN_SECRET}")"
assert_notify "notify_order_not_found" "${NOT_FOUND_RESP}" "False" "ORDER_NOT_FOUND"

INVALID_PARAMS_RESP="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/callback/notify" \
  -H "Content-Type: application/json" \
  -d '{"order_no":"","paid_amount":0,"upstream_trade_no":"","channel_id":0,"sign":"x"}')"
assert_notify "notify_invalid_params" "${INVALID_PARAMS_RESP}" "False" "INVALID_NOTIFY_PARAMS"

echo
echo "PASS: 收款主链路与回调状态矩阵验证成功"
echo "order_no=${ORDER_NO}"

#!/usr/bin/env bash
set -euo pipefail

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
MERCHANT_ID="${MERCHANT_ID:-m_demo}"
MERCHANT_SECRET="${MERCHANT_SECRET:-demo_secret}"
PAYOUT_PRODUCT_CODE="${PAYOUT_PRODUCT_CODE:-bank_card}"
AMOUNT="${AMOUNT:-1234}"
SIMULATE_PAYOUT_SUCCESS="${SIMULATE_PAYOUT_SUCCESS:-1}"

md5_sign() {
  local secret="$1"
  local kv_json="$2"
  python3 - "$secret" "$kv_json" <<'PY'
import hashlib, json, sys
secret = sys.argv[1]
params = json.loads(sys.argv[2])
pairs = []
for k in sorted(params.keys(), key=lambda x: x.lower()):
    if k.lower() == "sign":
        continue
    v = str(params[k]) if params[k] is not None else ""
    if v:
        pairs.append(f"{k.lower()}={v}")
raw = "&".join(pairs)
if raw:
    raw += "&"
raw += f"key={secret}"
print(hashlib.md5(raw.encode("utf-8")).hexdigest())
PY
}

json_get() {
  local json="$1"
  local path="$2"
  python3 - "$json" "$path" <<'PY'
import json, sys
try:
    obj = json.loads(sys.argv[1])
except Exception:
    print("")
    sys.exit(0)
path = sys.argv[2].split(".")
cur = obj
for p in path:
    if not isinstance(cur, dict) or p not in cur:
        print(""); sys.exit(0)
    cur = cur[p]
print(cur if cur is not None else "")
PY
}

json_has_nonempty_array() {
  local json="$1"
  local path="$2"
  python3 - "$json" "$path" <<'PY'
import json, sys
try:
    obj = json.loads(sys.argv[1])
except Exception:
    print("false")
    sys.exit(0)
path = sys.argv[2].split(".")
cur = obj
for p in path:
    if not isinstance(cur, dict) or p not in cur:
        print("false")
        sys.exit(0)
    cur = cur[p]
print("true" if isinstance(cur, list) and len(cur) > 0 else "false")
PY
}

signed_query_string() {
  local req_json="$1"
  local secret="$2"
  local sign
  sign="$(md5_sign "${secret}" "${req_json}")"
  python3 - "${req_json}" "${sign}" <<'PY'
import json, sys, urllib.parse
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(urllib.parse.urlencode(p))
PY
}

query_merchant_balance() {
  local req_json
  req_json="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","timestamp":int(time.time())}))
PY
)"
  local qs
  qs="$(signed_query_string "${req_json}" "${MERCHANT_SECRET}")"
  curl -fsS "${GATEWAY_BASE_URL}/v1/merchant/balance/query?${qs}"
}

echo "[1/6] merchant login + capture payout balance before create"
merchant_login_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/merchant/login" -H "Content-Type: application/json" -d "{\"merchant_id\":\"${MERCHANT_ID}\",\"api_secret\":\"${MERCHANT_SECRET}\"}")"
merchant_token="$(json_get "${merchant_login_resp}" "token")"
if [[ -z "${merchant_token}" ]]; then
  echo "merchant login failed: ${merchant_login_resp}"
  exit 1
fi
merchant_balance_before="$(query_merchant_balance)"
payout_balance_before="$(json_get "${merchant_balance_before}" "payout_balance")"
if [[ -z "${payout_balance_before}" ]]; then
  echo "get merchant balance before failed: ${merchant_balance_before}"
  exit 1
fi
echo "  payout_balance_before=${payout_balance_before}"

echo "[2/6] create payout order"
merchant_order_no="PO-E2E-$(date +%s)-$RANDOM"
params="$(python3 - <<PY
import json
print(json.dumps({
  "merchant_id": "${MERCHANT_ID}",
  "merchant_order_no": "${merchant_order_no}",
  "amount": int("${AMOUNT}"),
  "currency": "CNY",
  "payout_product_code": "${PAYOUT_PRODUCT_CODE}"
}))
PY
)"
sign="$(md5_sign "${MERCHANT_SECRET}" "${params}")"
body="$(python3 - "${params}" "${sign}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(json.dumps(p))
PY
)"
create_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${body}")"
order_no="$(json_get "${create_resp}" "order_no")"
if [[ -z "${order_no}" ]]; then
  echo "create payout order failed: ${create_resp}"
  exit 1
fi
echo "  order_no=${order_no}"

echo "[3/6] query payout order + verify debit amount"
query_params="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","order_no":"${order_no}","timestamp":str(int(time.time()))}))
PY
)"
query_url="$(signed_query_string "${query_params}" "${MERCHANT_SECRET}")"
query_resp="$(curl -fsS "${GATEWAY_BASE_URL}/v1/payout/query?${query_url}")"
query_order_no="$(json_get "${query_resp}" "order.order_no")"
if [[ "${query_order_no}" != "${order_no}" ]]; then
  echo "query payout order mismatch: ${query_resp}"
  exit 1
fi
query_fee_amount="$(json_get "${query_resp}" "order.fee_amount")"
if [[ -z "${query_fee_amount}" ]]; then
  echo "query payout order missing fee_amount: ${query_resp}"
  exit 1
fi
expected_debit=$(( AMOUNT + query_fee_amount ))
merchant_balance_after_create="$(query_merchant_balance)"
payout_balance_after_create="$(json_get "${merchant_balance_after_create}" "payout_balance")"
if [[ -z "${payout_balance_after_create}" ]]; then
  echo "get merchant balance after create failed: ${merchant_balance_after_create}"
  exit 1
fi
actual_debit=$(( payout_balance_before - payout_balance_after_create ))
if [[ "${actual_debit}" -ne "${expected_debit}" ]]; then
  echo "payout debit mismatch: expected=${expected_debit}, actual=${actual_debit}, query=${query_resp}, balance_before=${merchant_balance_before}, balance_after=${merchant_balance_after_create}"
  exit 1
fi
echo "  debit_ok expected=${expected_debit} actual=${actual_debit}"

echo "[4/6] merchant payout_orders list visibility"
merchant_list_resp="$(curl -fsS "${GATEWAY_BASE_URL}/v1/merchant/payout_orders?order_no=${order_no}&limit=20" -H "X-Merchant-Token: ${merchant_token}")"
if [[ "$(json_has_nonempty_array "${merchant_list_resp}" "orders")" != "true" ]]; then
  echo "merchant payout orders list failed: ${merchant_list_resp}"
  exit 1
fi

echo "[5/6] admin payout_orders list visibility"
admin_login_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/admin/login" -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}')"
admin_token="$(json_get "${admin_login_resp}" "token")"
if [[ -z "${admin_token}" ]]; then
  echo "admin login failed: ${admin_login_resp}"
  exit 1
fi
admin_list_resp="$(curl -fsS "${GATEWAY_BASE_URL}/v1/admin/payout_orders?merchant_id=${MERCHANT_ID}&keyword=${order_no}&limit=20" -H "X-Admin-Token: ${admin_token}")"
if [[ "$(json_has_nonempty_array "${admin_list_resp}" "orders")" != "true" ]]; then
  echo "admin payout orders list failed: ${admin_list_resp}"
  exit 1
fi

echo "[6/6] optional simulate payout success + verify no second debit"
if [[ "${SIMULATE_PAYOUT_SUCCESS}" == "1" ]]; then
  mock_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/admin/payout_orders/${order_no}/mock_success" -H "X-Admin-Token: ${admin_token}" -H "Content-Type: application/json" -d '{}')"
  mock_ok="$(json_get "${mock_resp}" "ok")"
  if [[ "${mock_ok}" != "True" && "${mock_ok}" != "true" ]]; then
    echo "mock payout success failed: ${mock_resp}"
    exit 1
  fi
  query_resp_after_success="$(curl -fsS "${GATEWAY_BASE_URL}/v1/payout/query?${query_url}")"
  query_status_after_success="$(json_get "${query_resp_after_success}" "order.status")"
  if [[ "${query_status_after_success}" != "1" ]]; then
    echo "simulate payout success failed, expected status=1: ${query_resp_after_success}"
    exit 1
  fi
  merchant_balance_after_success="$(query_merchant_balance)"
  payout_balance_after_success="$(json_get "${merchant_balance_after_success}" "payout_balance")"
  if [[ "${payout_balance_after_success}" -ne "${payout_balance_after_create}" ]]; then
    echo "payout balance changed unexpectedly after success simulation: before=${payout_balance_after_create}, after=${payout_balance_after_success}"
    exit 1
  fi
  echo "  simulated_success_ok status=${query_status_after_success} payout_balance_unchanged=${payout_balance_after_success}"
else
  echo "  skip simulate payout success (set SIMULATE_PAYOUT_SUCCESS=1 to enable)"
fi

echo "PASS: payout flow verification completed (balance API + mock success)"

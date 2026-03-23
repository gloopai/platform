#!/usr/bin/env bash
set -euo pipefail

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
MERCHANT_ID="${MERCHANT_ID:-m_demo}"
MERCHANT_SECRET="${MERCHANT_SECRET:-demo_secret}"
PAYOUT_PRODUCT_CODE="${PAYOUT_PRODUCT_CODE:-bank_card}"
AMOUNT="${AMOUNT:-1234}"

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
obj = json.loads(sys.argv[1]); path = sys.argv[2].split(".")
cur = obj
for p in path:
    if not isinstance(cur, dict) or p not in cur:
        print(""); sys.exit(0)
    cur = cur[p]
print(cur if cur is not None else "")
PY
}

echo "[1/4] create payout order"
merchant_order_no="PO-E2E-$(date +%s)"
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
create_resp="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${body}")"
order_no="$(json_get "${create_resp}" "order_no")"
if [[ -z "${order_no}" ]]; then
  echo "create payout order failed: ${create_resp}"
  exit 1
fi
echo "  order_no=${order_no}"

echo "[2/4] query payout order"
query_params="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","order_no":"${order_no}","timestamp":str(int(time.time()))}))
PY
)"
query_sign="$(md5_sign "${MERCHANT_SECRET}" "${query_params}")"
query_url="$(python3 - "${query_params}" "${query_sign}" <<'PY'
import json, sys, urllib.parse
p = json.loads(sys.argv[1]); p["sign"] = sys.argv[2]
print(urllib.parse.urlencode(p))
PY
)"
query_resp="$(curl -sS "${GATEWAY_BASE_URL}/v1/payout/query?${query_url}")"
query_order_no="$(json_get "${query_resp}" "order.order_no")"
if [[ "${query_order_no}" != "${order_no}" ]]; then
  echo "query payout order mismatch: ${query_resp}"
  exit 1
fi

echo "[3/4] merchant payout_orders list visibility"
merchant_login_resp="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/merchant/login" -H "Content-Type: application/json" -d "{\"merchant_id\":\"${MERCHANT_ID}\",\"api_secret\":\"${MERCHANT_SECRET}\"}")"
merchant_token="$(json_get "${merchant_login_resp}" "token")"
if [[ -z "${merchant_token}" ]]; then
  echo "merchant login failed: ${merchant_login_resp}"
  exit 1
fi
merchant_list_resp="$(curl -sS "${GATEWAY_BASE_URL}/v1/merchant/payout_orders?order_no=${order_no}&limit=20" -H "X-Merchant-Token: ${merchant_token}")"
if [[ "$(json_get "${merchant_list_resp}" "orders")" == "" ]]; then
  echo "merchant payout orders list failed: ${merchant_list_resp}"
  exit 1
fi

echo "[4/4] admin payout_orders list visibility"
admin_login_resp="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/admin/login" -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}')"
admin_token="$(json_get "${admin_login_resp}" "token")"
if [[ -z "${admin_token}" ]]; then
  echo "admin login failed: ${admin_login_resp}"
  exit 1
fi
admin_list_resp="$(curl -sS "${GATEWAY_BASE_URL}/v1/admin/payout_orders?merchant_id=${MERCHANT_ID}&keyword=${order_no}&limit=20" -H "X-Admin-Token: ${admin_token}")"
if [[ "$(json_get "${admin_list_resp}" "orders")" == "" ]]; then
  echo "admin payout orders list failed: ${admin_list_resp}"
  exit 1
fi

echo "PASS: payout flow basic verification completed"

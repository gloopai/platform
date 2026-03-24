#!/usr/bin/env bash
set -euo pipefail

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
MERCHANT_ID="${MERCHANT_ID:-m_demo}"
MERCHANT_SECRET="${MERCHANT_SECRET:-demo_secret}"
PAYOUT_PRODUCT_CODE="${PAYOUT_PRODUCT_CODE:-bank_card}"
AMOUNT="${AMOUNT:-1234}"
SIMULATE_PAYOUT_SUCCESS="${SIMULATE_PAYOUT_SUCCESS:-1}"
CONCURRENT_DUP_ATTEMPTS="${CONCURRENT_DUP_ATTEMPTS:-8}"
CONCURRENT_DUP_ATTEMPTS_HIGH="${CONCURRENT_DUP_ATTEMPTS_HIGH:-30}"
SECOND_MERCHANT_ID="${SECOND_MERCHANT_ID:-m_e2e_guard}"
SECOND_MERCHANT_SECRET="${SECOND_MERCHANT_SECRET:-e2e_guard_secret}"

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

build_signed_payout_body() {
  local merchant_order_no="$1"
  local amount="$2"
  local product_code="$3"
  local params
  params="$(python3 - <<PY
import json
print(json.dumps({
  "merchant_id": "${MERCHANT_ID}",
  "merchant_order_no": "${merchant_order_no}",
  "amount": int("${amount}"),
  "currency": "CNY",
  "payout_product_code": "${product_code}",
  "timestamp": int(__import__("time").time()),
  "nonce": __import__("uuid").uuid4().hex[:24]
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

build_signed_payout_query() {
  local merchant_id="$1"
  local merchant_secret="$2"
  local order_no="$3"
  local merchant_order_no="$4"
  local req_json
  req_json="$(python3 - <<PY
import json, time
d = {"merchant_id": "${merchant_id}", "timestamp": str(int(time.time()))}
if "${order_no}":
    d["order_no"] = "${order_no}"
if "${merchant_order_no}":
    d["merchant_order_no"] = "${merchant_order_no}"
print(json.dumps(d))
PY
)"
  signed_query_string "${req_json}" "${merchant_secret}"
}

query_payout_order() {
  local merchant_id="$1"
  local merchant_secret="$2"
  local order_no="$3"
  local merchant_order_no="$4"
  local qs
  qs="$(build_signed_payout_query "${merchant_id}" "${merchant_secret}" "${order_no}" "${merchant_order_no}")"
  curl -fsS "${GATEWAY_BASE_URL}/v1/payout/query?${qs}"
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

assert_http_error() {
  local raw="$1"
  local expected_status="$2"
  local expected_code="$3"
  local label="$4"
  local status_code body code
  status_code="${raw##*$'\n'}"
  body="${raw%$'\n'*}"
  code="$(json_get "${body}" "code")"
  if [[ "${status_code}" != "${expected_status}" || "${code}" != "${expected_code}" ]]; then
    echo "${label} failed: expected status=${expected_status} code=${expected_code}, got status=${status_code}, body=${body}"
    exit 1
  fi
  echo "  ${label}_ok status=${status_code} code=${code}"
}

echo "[1/18] merchant login + capture available balance before create"
merchant_login_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/merchant/login" -H "Content-Type: application/json" -d "{\"merchant_id\":\"${MERCHANT_ID}\",\"api_secret\":\"${MERCHANT_SECRET}\"}")"
merchant_token="$(json_get "${merchant_login_resp}" "token")"
if [[ -z "${merchant_token}" ]]; then
  echo "merchant login failed: ${merchant_login_resp}"
  exit 1
fi
merchant_balance_before="$(query_merchant_balance)"
available_balance_before="$(json_get "${merchant_balance_before}" "available_balance")"
if [[ -z "${available_balance_before}" ]]; then
  echo "get merchant balance before failed: ${merchant_balance_before}"
  exit 1
fi
echo "  available_balance_before=${available_balance_before}"

echo "[2/18] create payout order"
merchant_order_no="PO-E2E-$(date +%s)-$RANDOM"
body="$(build_signed_payout_body "${merchant_order_no}" "${AMOUNT}" "${PAYOUT_PRODUCT_CODE}")"
create_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${body}")"
order_no="$(json_get "${create_resp}" "order_no")"
if [[ -z "${order_no}" ]]; then
  echo "create payout order failed: ${create_resp}"
  exit 1
fi
echo "  order_no=${order_no}"

echo "[3/18] query payout order + verify debit amount"
query_params="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","order_no":"${order_no}","timestamp":str(int(time.time()))}))
PY
)"
query_resp="$(query_payout_order "${MERCHANT_ID}" "${MERCHANT_SECRET}" "${order_no}" "")"
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
available_balance_after_create="$(json_get "${merchant_balance_after_create}" "available_balance")"
if [[ -z "${available_balance_after_create}" ]]; then
  echo "get merchant balance after create failed: ${merchant_balance_after_create}"
  exit 1
fi
actual_debit=$(( available_balance_before - available_balance_after_create ))
if [[ "${actual_debit}" -ne "${expected_debit}" ]]; then
    echo "available balance debit mismatch: expected=${expected_debit}, actual=${actual_debit}, query=${query_resp}, balance_before=${merchant_balance_before}, balance_after=${merchant_balance_after_create}"
  exit 1
fi
echo "  debit_ok expected=${expected_debit} actual=${actual_debit}"

echo "[4/18] duplicate merchant_order_no while pending should fail (no second debit)"
dup_body="$(build_signed_payout_body "${merchant_order_no}" "${AMOUNT}" "${PAYOUT_PRODUCT_CODE}")"
dup_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${dup_body}" -w $'\n%{http_code}')"
dup_status="${dup_raw##*$'\n'}"
dup_body="${dup_raw%$'\n'*}"
dup_code="$(json_get "${dup_body}" "code")"
if [[ "${dup_status}" != "422" || "${dup_code}" != "PAYOUT_ORDER_ALREADY_EXISTS_PENDING" ]]; then
  echo "duplicate pending idempotent check failed: status=${dup_status}, body=${dup_body}"
  exit 1
fi
merchant_balance_after_dup="$(query_merchant_balance)"
available_balance_after_dup="$(json_get "${merchant_balance_after_dup}" "available_balance")"
if [[ "${available_balance_after_dup}" -ne "${available_balance_after_create}" ]]; then
  echo "available balance changed unexpectedly after duplicate create: before=${available_balance_after_create}, after=${available_balance_after_dup}"
  exit 1
fi
echo "  duplicate_pending_ok status=${dup_status} code=${dup_code} available_balance_unchanged=${available_balance_after_dup}"

echo "[5/18] insufficient available balance should fail without debit"
big_amount=$(( available_balance_after_create + 999999 ))
insufficient_order_no="PO-E2E-INSUFFICIENT-$(date +%s)-$RANDOM"
insufficient_body="$(build_signed_payout_body "${insufficient_order_no}" "${big_amount}" "${PAYOUT_PRODUCT_CODE}")"
insufficient_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${insufficient_body}" -w $'\n%{http_code}')"
insufficient_status="${insufficient_raw##*$'\n'}"
insufficient_resp="${insufficient_raw%$'\n'*}"
insufficient_code="$(json_get "${insufficient_resp}" "code")"
if [[ "${insufficient_status}" != "422" || "${insufficient_code}" != "INSUFFICIENT_AVAILABLE_BALANCE" ]]; then
  echo "insufficient balance check failed: status=${insufficient_status}, body=${insufficient_resp}"
  exit 1
fi
merchant_balance_after_insufficient="$(query_merchant_balance)"
available_balance_after_insufficient="$(json_get "${merchant_balance_after_insufficient}" "available_balance")"
if [[ "${available_balance_after_insufficient}" -ne "${available_balance_after_create}" ]]; then
  echo "available balance changed unexpectedly after insufficient-balance reject: before=${available_balance_after_create}, after=${available_balance_after_insufficient}"
  exit 1
fi
echo "  insufficient_balance_ok status=${insufficient_status} code=${insufficient_code} available_balance_unchanged=${available_balance_after_insufficient}"

echo "[6/18] insufficient balance path should mark order failed"
insufficient_query_params="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","merchant_order_no":"${insufficient_order_no}","timestamp":str(int(time.time()))}))
PY
)"
insufficient_query_url="$(signed_query_string "${insufficient_query_params}" "${MERCHANT_SECRET}")"
insufficient_query_resp="$(curl -fsS "${GATEWAY_BASE_URL}/v1/payout/query?${insufficient_query_url}")"
insufficient_query_order_no="$(json_get "${insufficient_query_resp}" "order.order_no")"
insufficient_query_status="$(json_get "${insufficient_query_resp}" "order.status")"
if [[ -z "${insufficient_query_order_no}" || "${insufficient_query_status}" != "2" ]]; then
  echo "insufficient-order failed-status check failed: ${insufficient_query_resp}"
  exit 1
fi
echo "  insufficient_failed_order_ok order_no=${insufficient_query_order_no} status=${insufficient_query_status}"

echo "[7/18] retry same insufficient merchant_order_no should return existing failed order"
insufficient_retry_body="$(build_signed_payout_body "${insufficient_order_no}" "${big_amount}" "${PAYOUT_PRODUCT_CODE}")"
insufficient_retry_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${insufficient_retry_body}" -w $'\n%{http_code}')"
insufficient_retry_status="${insufficient_retry_raw##*$'\n'}"
insufficient_retry_body="${insufficient_retry_raw%$'\n'*}"
insufficient_retry_order_no="$(json_get "${insufficient_retry_body}" "order_no")"
insufficient_retry_order_status="$(json_get "${insufficient_retry_body}" "status")"
if [[ "${insufficient_retry_status}" != "200" || -z "${insufficient_retry_order_no}" || "${insufficient_retry_order_status}" != "2" ]]; then
  echo "insufficient retry existing failed-order check failed: status=${insufficient_retry_status}, body=${insufficient_retry_body}"
  exit 1
fi
if [[ "${insufficient_retry_order_no}" != "${insufficient_query_order_no}" ]]; then
  echo "insufficient retry returned different order_no: expect=${insufficient_query_order_no}, got=${insufficient_retry_order_no}"
  exit 1
fi
merchant_balance_after_insufficient_retry="$(query_merchant_balance)"
available_balance_after_insufficient_retry="$(json_get "${merchant_balance_after_insufficient_retry}" "available_balance")"
if [[ "${available_balance_after_insufficient_retry}" -ne "${available_balance_after_insufficient}" ]]; then
  echo "available balance changed unexpectedly after insufficient retry: before=${available_balance_after_insufficient}, after=${available_balance_after_insufficient_retry}"
  exit 1
fi
echo "  insufficient_retry_failed_order_ok order_no=${insufficient_retry_order_no} status=${insufficient_retry_order_status} available_balance_unchanged=${available_balance_after_insufficient_retry}"

echo "[8/18] concurrent duplicate merchant_order_no should debit only once"
merchant_balance_before_concurrent="$(query_merchant_balance)"
available_balance_before_concurrent="$(json_get "${merchant_balance_before_concurrent}" "available_balance")"
concurrent_amount=1001
concurrent_merchant_order_no="PO-E2E-CONCURRENT-$(date +%s)-$RANDOM"
tmp_dir="$(mktemp -d)"
for i in $(seq 1 "${CONCURRENT_DUP_ATTEMPTS}"); do
  {
    req_body="$(build_signed_payout_body "${concurrent_merchant_order_no}" "${concurrent_amount}" "${PAYOUT_PRODUCT_CODE}")"
    curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${req_body}" -w $'\n%{http_code}' > "${tmp_dir}/resp_${i}.txt"
  } &
done
wait
success_count=0
pending_count=0
concurrent_order_no=""
unique_success_order_nos=""
for i in $(seq 1 "${CONCURRENT_DUP_ATTEMPTS}"); do
  raw="$(<"${tmp_dir}/resp_${i}.txt")"
  status_code="${raw##*$'\n'}"
  body_i="${raw%$'\n'*}"
  if [[ "${status_code}" == "200" ]]; then
    ono="$(json_get "${body_i}" "order_no")"
    if [[ -n "${ono}" ]]; then
      success_count=$((success_count + 1))
      concurrent_order_no="${ono}"
      if [[ ",${unique_success_order_nos}," != *",${ono},"* ]]; then
        if [[ -z "${unique_success_order_nos}" ]]; then
          unique_success_order_nos="${ono}"
        else
          unique_success_order_nos="${unique_success_order_nos},${ono}"
        fi
      fi
    fi
  elif [[ "${status_code}" == "422" && "$(json_get "${body_i}" "code")" == "PAYOUT_ORDER_ALREADY_EXISTS_PENDING" ]]; then
    pending_count=$((pending_count + 1))
  else
    echo "concurrent idempotent check failed: unexpected response status=${status_code}, body=${body_i}"
    rm -rf "${tmp_dir}"
    exit 1
  fi
done
if [[ "${success_count}" -lt 1 ]]; then
  echo "concurrent idempotent check failed: no success responses"
  rm -rf "${tmp_dir}"
  exit 1
fi
if [[ "${unique_success_order_nos}" == *","* ]]; then
  echo "concurrent idempotent check failed: multiple success order_no values=${unique_success_order_nos}"
  rm -rf "${tmp_dir}"
  exit 1
fi
if [[ -z "${concurrent_order_no}" ]]; then
  echo "concurrent idempotent check failed: empty order_no on success"
  rm -rf "${tmp_dir}"
  exit 1
fi
concurrent_query_params="$(python3 - <<PY
import json, time
print(json.dumps({"merchant_id":"${MERCHANT_ID}","order_no":"${concurrent_order_no}","timestamp":str(int(time.time()))}))
PY
)"
concurrent_query_url="$(signed_query_string "${concurrent_query_params}" "${MERCHANT_SECRET}")"
concurrent_query_resp="$(curl -fsS "${GATEWAY_BASE_URL}/v1/payout/query?${concurrent_query_url}")"
concurrent_fee_amount="$(json_get "${concurrent_query_resp}" "order.fee_amount")"
if [[ -z "${concurrent_fee_amount}" ]]; then
  echo "concurrent order query missing fee_amount: ${concurrent_query_resp}"
  rm -rf "${tmp_dir}"
  exit 1
fi
concurrent_expected_debit=$(( concurrent_amount + concurrent_fee_amount ))
merchant_balance_after_concurrent="$(query_merchant_balance)"
available_balance_after_concurrent="$(json_get "${merchant_balance_after_concurrent}" "available_balance")"
concurrent_actual_debit=$(( available_balance_before_concurrent - available_balance_after_concurrent ))
if [[ "${concurrent_actual_debit}" -ne "${concurrent_expected_debit}" ]]; then
  echo "concurrent debit mismatch: expected=${concurrent_expected_debit}, actual=${concurrent_actual_debit}"
  rm -rf "${tmp_dir}"
  exit 1
fi
rm -rf "${tmp_dir}"
echo "  concurrent_idempotent_ok attempts=${CONCURRENT_DUP_ATTEMPTS} success=${success_count} pending=${pending_count} debit=${concurrent_actual_debit}"

echo "[9/18] merchant payout_orders list visibility"
merchant_list_resp="$(curl -fsS "${GATEWAY_BASE_URL}/v1/merchant/payout_orders?order_no=${order_no}&limit=20" -H "X-Merchant-Token: ${merchant_token}")"
if [[ "$(json_has_nonempty_array "${merchant_list_resp}" "orders")" != "true" ]]; then
  echo "merchant payout orders list failed: ${merchant_list_resp}"
  exit 1
fi

echo "[10/18] admin payout_orders list visibility"
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

echo "[11/18] optional simulate payout success + verify no second debit"
if [[ "${SIMULATE_PAYOUT_SUCCESS}" == "1" ]]; then
  mock_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/admin/payout_orders/${order_no}/mock_success" -H "X-Admin-Token: ${admin_token}" -H "Content-Type: application/json" -d '{}')"
  mock_ok="$(json_get "${mock_resp}" "ok")"
  if [[ "${mock_ok}" != "True" && "${mock_ok}" != "true" ]]; then
    echo "mock payout success failed: ${mock_resp}"
    exit 1
  fi
  query_resp_after_success="$(query_payout_order "${MERCHANT_ID}" "${MERCHANT_SECRET}" "${order_no}" "")"
  query_status_after_success="$(json_get "${query_resp_after_success}" "order.status")"
  if [[ "${query_status_after_success}" != "1" ]]; then
    echo "simulate payout success failed, expected status=1: ${query_resp_after_success}"
    exit 1
  fi
  merchant_balance_after_success="$(query_merchant_balance)"
  available_balance_after_success="$(json_get "${merchant_balance_after_success}" "available_balance")"
  if [[ "${available_balance_after_success}" -ne "${available_balance_after_concurrent}" ]]; then
    echo "available balance changed unexpectedly after success simulation: before=${available_balance_after_concurrent}, after=${available_balance_after_success}"
    exit 1
  fi
  echo "  simulated_success_ok status=${query_status_after_success} available_balance_unchanged=${available_balance_after_success}"

  echo "[12/18] mock_success repeated call should be idempotent (changed=false)"
  mock_again_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/admin/payout_orders/${order_no}/mock_success" -H "X-Admin-Token: ${admin_token}" -H "Content-Type: application/json" -d '{}')"
  mock_again_ok="$(json_get "${mock_again_resp}" "ok")"
  mock_again_changed="$(json_get "${mock_again_resp}" "changed")"
  if [[ "${mock_again_ok}" != "True" && "${mock_again_ok}" != "true" ]]; then
    echo "mock payout success repeat failed: ${mock_again_resp}"
    exit 1
  fi
  if [[ "${mock_again_changed}" != "False" && "${mock_again_changed}" != "false" ]]; then
    echo "mock payout success repeat should be changed=false: ${mock_again_resp}"
    exit 1
  fi
  echo "  mock_repeat_ok changed=${mock_again_changed}"
else
  echo "  skip simulate payout success (set SIMULATE_PAYOUT_SUCCESS=1 to enable)"
fi

echo "[13/18] balance reconciliation summary"
merchant_balance_final="$(query_merchant_balance)"
available_balance_final="$(json_get "${merchant_balance_final}" "available_balance")"
total_expected_debit=$(( expected_debit + concurrent_expected_debit ))
total_actual_debit=$(( available_balance_before - available_balance_final ))
if [[ "${total_actual_debit}" -ne "${total_expected_debit}" ]]; then
  echo "final balance reconcile mismatch: expected_total_debit=${total_expected_debit}, actual_total_debit=${total_actual_debit}"
  exit 1
fi
echo "  balance_reconcile_ok expected_total_debit=${total_expected_debit} actual_total_debit=${total_actual_debit}"

echo "[14/18] invalid sign should return 401 INVALID_SIGN"
bad_sign_body="$(python3 - "${body}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p["sign"] = "deadbeef"
print(json.dumps(p))
PY
)"
bad_sign_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${bad_sign_body}" -w $'\n%{http_code}')"
assert_http_error "${bad_sign_raw}" "401" "INVALID_SIGN" "invalid_sign"

echo "[15/18] missing sign should return 400 SIGN_REQUIRED"
missing_sign_body="$(python3 - "${body}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p.pop("sign", None)
print(json.dumps(p))
PY
)"
missing_sign_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${missing_sign_body}" -w $'\n%{http_code}')"
assert_http_error "${missing_sign_raw}" "400" "SIGN_REQUIRED" "missing_sign"

echo "[16/18] missing merchant_id should return 400 MERCHANT_ID_REQUIRED"
missing_mid_body="$(python3 - "${body}" <<'PY'
import json, sys
p = json.loads(sys.argv[1]); p.pop("merchant_id", None)
print(json.dumps(p))
PY
)"
missing_mid_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${missing_mid_body}" -w $'\n%{http_code}')"
assert_http_error "${missing_mid_raw}" "400" "MERCHANT_ID_REQUIRED" "missing_merchant_id"

echo "[17/18] cross-merchant isolation check (query should be ORDER_NOT_FOUND)"
create_second_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/admin/merchants" -H "X-Admin-Token: ${admin_token}" -H "Content-Type: application/json" -d "{\"merchant_id\":\"${SECOND_MERCHANT_ID}\",\"api_secret\":\"${SECOND_MERCHANT_SECRET}\",\"ip_whitelist\":\"127.0.0.1\"}" -w $'\n%{http_code}')"
create_second_status="${create_second_raw##*$'\n'}"
if [[ "${create_second_status}" != "200" ]]; then
  create_second_body="${create_second_raw%$'\n'*}"
  echo "  create second merchant status=${create_second_status} (may already exist), continue"
  echo "  create second merchant body=${create_second_body}"
fi
second_login_resp="$(curl -fsS -X POST "${GATEWAY_BASE_URL}/v1/merchant/login" -H "Content-Type: application/json" -d "{\"merchant_id\":\"${SECOND_MERCHANT_ID}\",\"api_secret\":\"${SECOND_MERCHANT_SECRET}\"}")"
second_merchant_token="$(json_get "${second_login_resp}" "token")"
if [[ -z "${second_merchant_token}" ]]; then
  echo "second merchant login failed: ${second_login_resp}"
  exit 1
fi
cross_query_qs="$(build_signed_payout_query "${SECOND_MERCHANT_ID}" "${SECOND_MERCHANT_SECRET}" "${order_no}" "")"
cross_query_raw="$(curl -sS "${GATEWAY_BASE_URL}/v1/payout/query?${cross_query_qs}" -w $'\n%{http_code}')"
assert_http_error "${cross_query_raw}" "404" "ORDER_NOT_FOUND" "cross_merchant_query"
cross_list_resp="$(curl -fsS "${GATEWAY_BASE_URL}/v1/merchant/payout_orders?order_no=${order_no}&limit=20" -H "X-Merchant-Token: ${second_merchant_token}")"
if [[ "$(json_has_nonempty_array "${cross_list_resp}" "orders")" == "true" ]]; then
  echo "cross merchant list should not see foreign order: ${cross_list_resp}"
  exit 1
fi
echo "  cross_merchant_list_ok no foreign orders visible"

echo "[18/18] high concurrency duplicate test should stay stable"
high_tmp_dir="$(mktemp -d)"
high_amount=1003
high_order_no="PO-E2E-HIGHCON-$(date +%s)-$RANDOM"
for i in $(seq 1 "${CONCURRENT_DUP_ATTEMPTS_HIGH}"); do
  {
    req_body="$(build_signed_payout_body "${high_order_no}" "${high_amount}" "${PAYOUT_PRODUCT_CODE}")"
    curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${req_body}" -w $'\n%{http_code}' > "${high_tmp_dir}/resp_${i}.txt"
  } &
done
wait
high_success=0
high_pending=0
high_order_created=""
for i in $(seq 1 "${CONCURRENT_DUP_ATTEMPTS_HIGH}"); do
  raw="$(<"${high_tmp_dir}/resp_${i}.txt")"
  sc="${raw##*$'\n'}"
  bi="${raw%$'\n'*}"
  if [[ "${sc}" == "200" && -n "$(json_get "${bi}" "order_no")" ]]; then
    high_success=$((high_success + 1))
    high_order_created="$(json_get "${bi}" "order_no")"
  elif [[ "${sc}" == "422" && "$(json_get "${bi}" "code")" == "PAYOUT_ORDER_ALREADY_EXISTS_PENDING" ]]; then
    high_pending=$((high_pending + 1))
  else
    echo "high concurrency unexpected response status=${sc}, body=${bi}"
    rm -rf "${high_tmp_dir}"
    exit 1
  fi
done
rm -rf "${high_tmp_dir}"
if [[ "${high_success}" -lt 1 || -z "${high_order_created}" ]]; then
  echo "high concurrency failed: no valid success response"
  exit 1
fi
echo "  high_concurrency_ok attempts=${CONCURRENT_DUP_ATTEMPTS_HIGH} success=${high_success} pending=${high_pending}"

echo "[extra] invalid amount should return 400 INVALID_ARGUMENT"
invalid_amount_order_no="PO-E2E-INVALID-AMOUNT-$(date +%s)-$RANDOM"
invalid_amount_body="$(build_signed_payout_body "${invalid_amount_order_no}" "0" "${PAYOUT_PRODUCT_CODE}")"
invalid_amount_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${invalid_amount_body}" -w $'\n%{http_code}')"
assert_http_error "${invalid_amount_raw}" "400" "INVALID_ARGUMENT" "invalid_amount"

echo "[extra] malformed json should return 400 INVALID_PARAMS"
malformed_json_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d '{"merchant_id":' -w $'\n%{http_code}')"
assert_http_error "${malformed_json_raw}" "400" "INVALID_PARAMS" "malformed_json"

echo "[extra] replay same signed request should return 409 REPLAY_REQUEST"
replay_order_no="PO-E2E-REPLAY-$(date +%s)-$RANDOM"
replay_body="$(build_signed_payout_body "${replay_order_no}" "1011" "${PAYOUT_PRODUCT_CODE}")"
replay_first_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${replay_body}" -w $'\n%{http_code}')"
replay_first_status="${replay_first_raw##*$'\n'}"
if [[ "${replay_first_status}" != "200" ]]; then
  echo "replay baseline first request failed unexpectedly: ${replay_first_raw}"
  exit 1
fi
replay_second_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${replay_body}" -w $'\n%{http_code}')"
assert_http_error "${replay_second_raw}" "409" "REPLAY_REQUEST" "replay_request"

echo "[extra] wrong content-type should return 400 MERCHANT_ID_REQUIRED"
wrong_ct_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: text/plain" -d "${body}" -w $'\n%{http_code}')"
assert_http_error "${wrong_ct_raw}" "400" "MERCHANT_ID_REQUIRED" "wrong_content_type"

echo "[extra] amount type mismatch should return 400 INVALID_PARAMS"
type_mismatch_body="$(python3 - "${MERCHANT_ID}" "${MERCHANT_SECRET}" "${PAYOUT_PRODUCT_CODE}" <<'PY'
import hashlib, json, sys, time
merchant_id, secret, product = sys.argv[1], sys.argv[2], sys.argv[3]
p = {
  "merchant_id": merchant_id,
  "merchant_order_no": f"PO-E2E-TYPE-MISMATCH-{int(time.time())}",
  "amount": "abc",
  "currency": "CNY",
  "payout_product_code": product,
  "timestamp": int(time.time()),
  "nonce": __import__("uuid").uuid4().hex[:24],
}
pairs = []
for k in sorted(p.keys(), key=lambda x: x.lower()):
    v = str(p[k]) if p[k] is not None else ""
    if v:
        pairs.append(f"{k.lower()}={v}")
raw = "&".join(pairs)
if raw:
    raw += "&"
raw += f"key={secret}"
p["sign"] = hashlib.md5(raw.encode("utf-8")).hexdigest()
print(json.dumps(p))
PY
)"
type_mismatch_raw="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/payout/order" -H "Content-Type: application/json" -d "${type_mismatch_body}" -w $'\n%{http_code}')"
assert_http_error "${type_mismatch_raw}" "400" "INVALID_PARAMS" "amount_type_mismatch"

echo "PASS: payout flow verification completed (success + failure + concurrent idempotency + isolation + high concurrency + balance reconciliation + openapi error matrix + input anomaly matrix)"

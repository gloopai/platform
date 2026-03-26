#!/usr/bin/env bash
set -euo pipefail

# 全自动 RBAC 生效性回归脚本（创建临时角色 -> 授权 -> 绑定用户 -> 校验 -> 自动恢复）
#
# 默认目标：admin 用户（需提供账号密码）
#
# 示例：
#   GATEWAY_BASE_URL=http://127.0.0.1:8080 \
#   ADMIN_USERNAME=admin ADMIN_PASSWORD=admin123 \
#   TARGET_USERNAME=admin \
#   TEST_MENU_KEYS="menu.rbac_menus,menu.rbac_roles" \
#   TEST_PERM_KEYS="admin.rbac.my_menu" \
#   EXPECT_ALLOW_MENU_PATHS="/rbac/menus,/rbac/roles" \
#   EXPECT_DENY_MENU_PATHS="/merchants,/channels" \
#   EXPECT_ALLOW_ENDPOINTS=$'GET /v1/admin/rbac/my_menu\nGET /v1/admin/rbac/menus' \
#   EXPECT_DENY_ENDPOINTS=$'GET /v1/admin/merchants\nGET /v1/admin/channels' \
#   ./test_admin_rbac_effective_full.sh
#
# 约定：
# - allow endpoint 期望 2xx
# - deny endpoint 期望 401/403
# - 脚本会强制附加内部必需授权，确保能做“恢复现场”：
#     menu: menu.rbac_menus
#     perm: admin.rbac.manage

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"
TARGET_USERNAME="${TARGET_USERNAME:-admin}"
CACHE_WAIT_SECONDS="${CACHE_WAIT_SECONDS:-11}"
ALLOW_TARGET_ADMIN="${ALLOW_TARGET_ADMIN:-1}"
STRICT_SAFETY_CHECK="${STRICT_SAFETY_CHECK:-0}"
TARGET_USER_ID_OVERRIDE="${TARGET_USER_ID_OVERRIDE:-}"
CURL_MAX_TIME="${CURL_MAX_TIME:-15}"

TEST_MENU_KEYS="${TEST_MENU_KEYS:-menu.rbac_menus,menu.rbac_roles}"
TEST_PERM_KEYS="${TEST_PERM_KEYS:-admin.rbac.my_menu}"

EXPECT_ALLOW_MENU_PATHS="${EXPECT_ALLOW_MENU_PATHS:-/rbac/menus}"
EXPECT_DENY_MENU_PATHS="${EXPECT_DENY_MENU_PATHS:-/merchants}"

DEFAULT_ALLOW_ENDPOINTS=$'GET /v1/admin/rbac/my_menu\nGET /v1/admin/rbac/menus\nGET /v1/admin/rbac/permissions'
DEFAULT_DENY_ENDPOINTS=$'GET /v1/admin/merchants'
EXPECT_ALLOW_ENDPOINTS="${EXPECT_ALLOW_ENDPOINTS:-$DEFAULT_ALLOW_ENDPOINTS}"
EXPECT_DENY_ENDPOINTS="${EXPECT_DENY_ENDPOINTS:-$DEFAULT_DENY_ENDPOINTS}"

INTERNAL_MENU_KEYS="menu.rbac_menus"
INTERNAL_PERM_KEYS="admin.rbac.manage,admin.admin_users.manage,admin.rbac.my_menu"

TMP_ROLE_ID=""
TARGET_USER_ID=""
ORIGINAL_ROLE_IDS=""
RESTORE_NEEDED="0"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl not found"
  exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 not found"
  exit 1
fi

fail() {
  echo "FAIL: $*" >&2
  exit 1
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
parts = [p for p in sys.argv[2].split(".") if p]
cur = obj
for p in parts:
    if not isinstance(cur, dict) or p not in cur:
        print("")
        sys.exit(0)
    cur = cur[p]
if cur is None:
    print("")
else:
    print(cur)
PY
}

csv_normalize() {
  local csv="$1"
  python3 - "$csv" <<'PY'
import sys
s = (sys.argv[1] or "").strip()
if not s:
    print("")
    sys.exit(0)
out = []
seen = set()
for p in s.split(","):
    k = p.strip()
    if not k or k in seen:
        continue
    seen.add(k)
    out.append(k)
print(",".join(out))
PY
}

csv_union() {
  local a="$1"
  local b="$2"
  python3 - "$a" "$b" <<'PY'
import sys
out = []
seen = set()
for raw in [sys.argv[1], sys.argv[2]]:
    if not raw:
        continue
    for p in raw.split(","):
        k = p.strip()
        if not k or k in seen:
            continue
        seen.add(k)
        out.append(k)
print(",".join(out))
PY
}

split_csv_to_lines() {
  local csv="$1"
  python3 - "$csv" <<'PY'
import sys
s = (sys.argv[1] or "").strip()
if not s:
    sys.exit(0)
for p in s.split(","):
    x = p.strip()
    if x:
        print(x)
PY
}

api_call() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local token="${4:-$ADMIN_TOKEN}"
  local output
  if [[ "${method}" == "GET" || "${method}" == "DELETE" ]]; then
    output="$(curl -sS --max-time "${CURL_MAX_TIME}" -X "${method}" \
      "${GATEWAY_BASE_URL}${path}" \
      -H "X-Admin-Token: ${token}" \
      -w $'\n%{http_code}')"
  else
    output="$(curl -sS --max-time "${CURL_MAX_TIME}" -X "${method}" \
      "${GATEWAY_BASE_URL}${path}" \
      -H "X-Admin-Token: ${token}" \
      -H "Content-Type: application/json" \
      -d "${body:-{}}" \
      -w $'\n%{http_code}')"
  fi
  local status="${output##*$'\n'}"
  local resp="${output%$'\n'*}"
  printf '%s\t%s' "${status}" "${resp}"
}

assert_2xx() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local token="${4:-$ADMIN_TOKEN}"
  local result status resp
  result="$(api_call "${method}" "${path}" "${body}" "${token}")"
  status="${result%%$'\t'*}"
  resp="${result#*$'\t'}"
  if [[ "${status}" != 2* ]]; then
    echo "FAIL: ${method} ${path} expected 2xx, got ${status}"
    echo "  resp=${resp}"
    exit 1
  fi
  printf '%s' "${resp}"
}

resolve_user_id_by_username() {
  local username="$1"
  local result status users_json
  result="$(api_call GET "/v1/admin/admin_users")"
  status="${result%%$'\t'*}"
  users_json="${result#*$'\t'}"
  if [[ "${status}" != 2* ]]; then
    echo "FAIL: GET /v1/admin/admin_users expected 2xx, got ${status}" >&2
    echo "  resp=${users_json}" >&2
    return 1
  fi
  python3 - "$users_json" "$username" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
username = sys.argv[2]
for u in (obj.get("users") or []):
    if isinstance(u, dict) and u.get("username") == username:
        print(u.get("id") or "")
        sys.exit(0)
print("")
PY
}

resolve_menu_ids_from_keys() {
  local key_csv="$1"
  local result status menus_json
  result="$(api_call GET "/v1/admin/rbac/menus")"
  status="${result%%$'\t'*}"
  menus_json="${result#*$'\t'}"
  if [[ "${status}" != 2* ]]; then
    echo "FAIL: GET /v1/admin/rbac/menus expected 2xx, got ${status}" >&2
    echo "  resp=${menus_json}" >&2
    return 1
  fi
  python3 - "$menus_json" "$key_csv" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
keys = [k.strip() for k in (sys.argv[2] or "").split(",") if k.strip()]
m = {}
id_to_parent = {}
for row in (obj.get("menus") or []):
    if isinstance(row, dict):
        mk = (row.get("menu_key") or "").strip()
        mid = row.get("id")
        pid = row.get("parent_id")
        if mk:
            m[mk] = mid
        if mid:
            id_to_parent[int(mid)] = int(pid or 0)
missing = []
expanded = set()
for k in keys:
    if k not in m or not m[k]:
        missing.append(k)
    else:
        cur = int(m[k])
        while cur and cur not in expanded:
            expanded.add(cur)
            cur = id_to_parent.get(cur, 0)
if missing:
    print("MISSING:" + ",".join(missing))
else:
    ids = sorted(expanded)
    print(",".join(str(i) for i in ids))
PY
}

resolve_menu_paths_from_keys() {
  local key_csv="$1"
  local result status menus_json
  result="$(api_call GET "/v1/admin/rbac/menus")"
  status="${result%%$'\t'*}"
  menus_json="${result#*$'\t'}"
  if [[ "${status}" != 2* ]]; then
    echo "FAIL: GET /v1/admin/rbac/menus expected 2xx, got ${status}" >&2
    echo "  resp=${menus_json}" >&2
    return 1
  fi
  python3 - "$menus_json" "$key_csv" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
keys = [k.strip() for k in (sys.argv[2] or "").split(",") if k.strip()]
m = {}
for row in (obj.get("menus") or []):
    if not isinstance(row, dict):
        continue
    mk = (row.get("menu_key") or "").strip()
    p = (row.get("path") or "").strip()
    if mk:
        m[mk] = p
for k in keys:
    p = m.get(k, "")
    if p:
        print(p)
PY
}

flatten_menu_paths() {
  local my_menu_json="$1"
  python3 - "$my_menu_json" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
seen = set()
def emit(p):
    if isinstance(p, str) and p.strip() and p not in seen:
        seen.add(p)
        print(p)
for x in (obj.get("sidebar") or []):
    if isinstance(x, dict):
        emit(x.get("to"))
        for c in (x.get("children") or []):
            if isinstance(c, dict):
                emit(c.get("to"))
for x in (obj.get("avatar_links") or []):
    if isinstance(x, dict):
        emit(x.get("to"))
PY
}

has_path() {
  local all_paths="$1"
  local target="$2"
  python3 - "$all_paths" "$target" <<'PY'
import sys
raw_paths = sys.argv[1] or ""
target = (sys.argv[2] or "").strip()
target_norm = target.rstrip("/") if target != "/" else "/"
paths = set()
for line in raw_paths.splitlines():
    p = line.strip()
    if not p:
        continue
    p_norm = p.rstrip("/") if p != "/" else "/"
    paths.add(p_norm)
sys.exit(0 if target_norm in paths else 1)
PY
}

csv_contains_value() {
  local csv="$1"
  local value="$2"
  python3 - "$csv" "$value" <<'PY'
import sys
vals = {x.strip() for x in (sys.argv[1] or "").split(",") if x.strip()}
target = (sys.argv[2] or "").strip()
sys.exit(0 if target in vals else 1)
PY
}

assert_allow_endpoint() {
  local method="$1"
  local path="$2"
  local result status resp
  result="$(api_call "${method}" "${path}" "" "${TARGET_TOKEN}")"
  status="${result%%$'\t'*}"
  resp="${result#*$'\t'}"
  if [[ "${status}" != 2* ]]; then
    echo "FAIL: expect ALLOW but got ${status} for ${method} ${path}"
    echo "  resp=${resp}"
    exit 1
  fi
  echo "  api_allow_ok ${method} ${path} -> ${status}"
}

assert_deny_endpoint() {
  local method="$1"
  local path="$2"
  local result status resp
  result="$(api_call "${method}" "${path}" "" "${TARGET_TOKEN}")"
  status="${result%%$'\t'*}"
  resp="${result#*$'\t'}"
  if [[ "${status}" != "401" && "${status}" != "403" ]]; then
    echo "FAIL: expect DENY(401/403) but got ${status} for ${method} ${path}"
    echo "  resp=${resp}"
    exit 1
  fi
  echo "  api_deny_ok ${method} ${path} -> ${status}"
}

cleanup() {
  local exit_code=$?
  if [[ "${RESTORE_NEEDED}" == "1" && -n "${TARGET_USER_ID}" && -n "${ORIGINAL_ROLE_IDS}" ]]; then
    local body
    body="$(python3 - "$ORIGINAL_ROLE_IDS" <<'PY'
import json, sys
ids = [int(x) for x in (sys.argv[1] or "").split(",") if x.strip()]
print(json.dumps({"role_ids": ids}))
PY
)"
    local r
    r="$(api_call PUT "/v1/admin/rbac/admin_users/${TARGET_USER_ID}/roles" "${body}" "${ADMIN_TOKEN}")" || true
    echo "  cleanup_restore_roles status=${r%%$'\t'*}"
  fi
  if [[ -n "${TMP_ROLE_ID}" ]]; then
    local r
    r="$(api_call DELETE "/v1/admin/rbac/roles/${TMP_ROLE_ID}" "" "${ADMIN_TOKEN}")" || true
    echo "  cleanup_delete_tmp_role status=${r%%$'\t'*}"
  fi
  exit "${exit_code}"
}
trap cleanup EXIT

echo "[1/10] admin login"
LOGIN_BODY="$(python3 - <<PY
import json
print(json.dumps({"username":"${ADMIN_USERNAME}","password":"${ADMIN_PASSWORD}"}))
PY
)"
LOGIN_RESP="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/admin/login" -H "Content-Type: application/json" -d "${LOGIN_BODY}")"
ADMIN_TOKEN="$(json_get "${LOGIN_RESP}" "token")"
if [[ -z "${ADMIN_TOKEN}" ]]; then
  fail "admin login failed: ${LOGIN_RESP}"
fi
echo "  login_ok"

PRECHECK_USERS_RESULT="$(api_call GET "/v1/admin/admin_users")"
PRECHECK_USERS_STATUS="${PRECHECK_USERS_RESULT%%$'\t'*}"
if [[ "${PRECHECK_USERS_STATUS}" != 2* ]]; then
  fail "current admin token has no access to /v1/admin/admin_users (status=${PRECHECK_USERS_STATUS}). Please restore admin.admin_users.manage first, then rerun."
fi

if [[ "${TARGET_USERNAME}" == "${ADMIN_USERNAME}" ]]; then
  if [[ "${ALLOW_TARGET_ADMIN}" != "1" || "${STRICT_SAFETY_CHECK}" == "1" ]]; then
    echo "FAIL: TARGET_USERNAME is same as ADMIN_USERNAME (${ADMIN_USERNAME})."
    echo "      Set ALLOW_TARGET_ADMIN=1 and STRICT_SAFETY_CHECK=0 to continue."
    exit 1
  fi
  echo "WARN: TARGET_USERNAME equals ADMIN_USERNAME; script will continue and rely on automatic cleanup."
fi

echo "[2/10] resolve target user and backup role_ids"
if [[ -n "${TARGET_USER_ID_OVERRIDE}" ]]; then
  TARGET_USER_ID="${TARGET_USER_ID_OVERRIDE}"
else
  TARGET_USER_ID="$(resolve_user_id_by_username "${TARGET_USERNAME}")"
fi
if [[ -z "${TARGET_USER_ID}" ]]; then
  fail "target user not found: ${TARGET_USERNAME}"
fi
ORIG_ROLE_RESULT="$(api_call GET "/v1/admin/rbac/admin_users/${TARGET_USER_ID}/roles")"
ORIG_ROLE_STATUS="${ORIG_ROLE_RESULT%%$'\t'*}"
ORIG_ROLE_JSON="${ORIG_ROLE_RESULT#*$'\t'}"
if [[ "${ORIG_ROLE_STATUS}" != 2* ]]; then
  fail "GET /v1/admin/rbac/admin_users/${TARGET_USER_ID}/roles expected 2xx, got ${ORIG_ROLE_STATUS}; resp=${ORIG_ROLE_JSON}"
fi
ORIGINAL_ROLE_IDS="$(python3 - "$ORIG_ROLE_JSON" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
print(",".join(str(int(x)) for x in (obj.get("role_ids") or [])))
PY
)"
RESTORE_NEEDED="1"
echo "  target_user_id=${TARGET_USER_ID} backup_roles=[${ORIGINAL_ROLE_IDS}]"

echo "[3/10] create temp role"
TMP_ROLE_CODE="rbac_e2e_$(date +%s)_$RANDOM"
CREATE_ROLE_BODY="$(python3 - <<PY
import json
print(json.dumps({"code":"${TMP_ROLE_CODE}","name":"RBAC回归临时角色","status":1}))
PY
)"
CREATE_ROLE_RESP="$(assert_2xx POST "/v1/admin/rbac/roles" "${CREATE_ROLE_BODY}")"
TMP_ROLE_ID="$(json_get "${CREATE_ROLE_RESP}" "role.id")"
if [[ -z "${TMP_ROLE_ID}" || "${TMP_ROLE_ID}" == "0" ]]; then
  echo "FAIL: create temp role failed: ${CREATE_ROLE_RESP}"
  exit 1
fi
echo "  tmp_role_id=${TMP_ROLE_ID}"

echo "[4/10] grant menus/perms to temp role"
ALL_MENU_KEYS="$(csv_union "$(csv_normalize "${TEST_MENU_KEYS}")" "${INTERNAL_MENU_KEYS}")"
ALL_PERM_KEYS="$(csv_union "$(csv_normalize "${TEST_PERM_KEYS}")" "${INTERNAL_PERM_KEYS}")"

MENU_IDS_RAW="$(resolve_menu_ids_from_keys "${ALL_MENU_KEYS}")"
if [[ "${MENU_IDS_RAW}" == MISSING:* ]]; then
  echo "FAIL: some menu_key not found: ${MENU_IDS_RAW#MISSING:}"
  exit 1
fi
SET_MENUS_BODY="$(python3 - "$MENU_IDS_RAW" <<'PY'
import json, sys
ids = [int(x) for x in (sys.argv[1] or "").split(",") if x.strip()]
print(json.dumps({"menu_ids": ids}))
PY
)"
assert_2xx PUT "/v1/admin/rbac/roles/${TMP_ROLE_ID}/menus" "${SET_MENUS_BODY}" >/dev/null

SET_PERMS_BODY="$(python3 - "$ALL_PERM_KEYS" <<'PY'
import json, sys
keys = [x.strip() for x in (sys.argv[1] or "").split(",") if x.strip()]
print(json.dumps({"perm_keys": keys}))
PY
)"
assert_2xx PUT "/v1/admin/rbac/roles/${TMP_ROLE_ID}/perm_keys" "${SET_PERMS_BODY}" >/dev/null
echo "  grant_ok menus=[${ALL_MENU_KEYS}] perms=[${ALL_PERM_KEYS}]"

echo "[5/10] bind temp role to target user (replace mode)"
SET_USER_ROLES_BODY="$(python3 - "$TMP_ROLE_ID" <<'PY'
import json, sys
print(json.dumps({"role_ids":[int(sys.argv[1])]}))
PY
)"
assert_2xx PUT "/v1/admin/rbac/admin_users/${TARGET_USER_ID}/roles" "${SET_USER_ROLES_BODY}" >/dev/null

CHECK_USER_ROLES_JSON="$(assert_2xx GET "/v1/admin/rbac/admin_users/${TARGET_USER_ID}/roles")"
CHECK_USER_ROLE_IDS="$(python3 - "$CHECK_USER_ROLES_JSON" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
print(",".join(str(int(x)) for x in (obj.get("role_ids") or [])))
PY
)"
if ! csv_contains_value "${CHECK_USER_ROLE_IDS}" "${TMP_ROLE_ID}"; then
  fail "target user role binding verification failed; expected role_id=${TMP_ROLE_ID}, got [${CHECK_USER_ROLE_IDS}]"
fi

CHECK_ROLE_MENUS_JSON="$(assert_2xx GET "/v1/admin/rbac/roles/${TMP_ROLE_ID}/menus")"
CHECK_ROLE_MENU_IDS="$(python3 - "$CHECK_ROLE_MENUS_JSON" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
print(",".join(str(int(x)) for x in (obj.get("menu_ids") or [])))
PY
)"
for mid in $(echo "${MENU_IDS_RAW}" | tr ',' ' '); do
  if [[ -n "${mid}" ]] && ! csv_contains_value "${CHECK_ROLE_MENU_IDS}" "${mid}"; then
    fail "role menu binding verification failed; expected menu_id=${mid}, got [${CHECK_ROLE_MENU_IDS}]"
  fi
done

CHECK_ROLE_PERMS_JSON="$(assert_2xx GET "/v1/admin/rbac/roles/${TMP_ROLE_ID}/perm_keys")"
CHECK_ROLE_PERMS="$(python3 - "$CHECK_ROLE_PERMS_JSON" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
print(",".join(str(x).strip() for x in (obj.get("perm_keys") or []) if str(x).strip()))
PY
)"
for pk in $(echo "${ALL_PERM_KEYS}" | tr ',' ' '); do
  if [[ -n "${pk}" ]] && ! csv_contains_value "${CHECK_ROLE_PERMS}" "${pk}"; then
    fail "role permission binding verification failed; expected perm_key=${pk}, got [${CHECK_ROLE_PERMS}]"
  fi
done

if [[ "${CACHE_WAIT_SECONDS}" -gt 0 ]]; then
  echo "  waiting ${CACHE_WAIT_SECONDS}s for RBAC middleware cache TTL..."
  sleep "${CACHE_WAIT_SECONDS}"
fi

echo "[6/10] target user relogin"
TARGET_LOGIN_RESP="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/admin/login" -H "Content-Type: application/json" -d "${LOGIN_BODY}")"
TARGET_TOKEN="$(json_get "${TARGET_LOGIN_RESP}" "token")"
if [[ -z "${TARGET_TOKEN}" ]]; then
  echo "FAIL: target relogin failed: ${TARGET_LOGIN_RESP}"
  exit 1
fi
echo "  target_login_ok"

echo "[7/10] verify menu visibility"
TARGET_MY_MENU_RESULT="$(api_call GET "/v1/admin/rbac/my_menu" "" "${TARGET_TOKEN}")"
TARGET_MY_MENU_STATUS="${TARGET_MY_MENU_RESULT%%$'\t'*}"
TARGET_MY_MENU="${TARGET_MY_MENU_RESULT#*$'\t'}"
if [[ "${TARGET_MY_MENU_STATUS}" != 2* ]]; then
  echo "FAIL: GET /v1/admin/rbac/my_menu expected 2xx, got ${TARGET_MY_MENU_STATUS}"
  echo "  resp=${TARGET_MY_MENU}"
  exit 1
fi
TARGET_MENU_PATHS="$(flatten_menu_paths "${TARGET_MY_MENU}")"

while IFS= read -r p; do
  [[ -z "${p}" ]] && continue
  if ! has_path "${TARGET_MENU_PATHS}" "${p}"; then
    echo "  debug_target_menu_paths=$(printf '%s' "${TARGET_MENU_PATHS}" | tr '\n' ',' | sed 's/,$//')"
    echo "  debug_target_my_menu=${TARGET_MY_MENU}"
    echo "FAIL: expected allow menu path missing: ${p}"
    exit 1
  fi
  echo "  menu_allow_ok ${p}"
done < <(split_csv_to_lines "${EXPECT_ALLOW_MENU_PATHS}")

while IFS= read -r p; do
  [[ -z "${p}" ]] && continue
  if has_path "${TARGET_MENU_PATHS}" "${p}"; then
    echo "FAIL: expected deny menu path still visible: ${p}"
    exit 1
  fi
  echo "  menu_deny_ok ${p}"
done < <(split_csv_to_lines "${EXPECT_DENY_MENU_PATHS}")

echo "[8/10] verify API allow matrix"
while IFS= read -r line; do
  [[ -z "${line}" ]] && continue
  method="$(echo "${line}" | awk '{print $1}')"
  path="$(echo "${line}" | awk '{print $2}')"
  if [[ -z "${method}" || -z "${path}" ]]; then
    echo "FAIL: bad allow endpoint format: ${line}"
    exit 1
  fi
  assert_allow_endpoint "${method}" "${path}"
done <<< "${EXPECT_ALLOW_ENDPOINTS}"

echo "[9/10] verify API deny matrix"
while IFS= read -r line; do
  [[ -z "${line}" ]] && continue
  method="$(echo "${line}" | awk '{print $1}')"
  path="$(echo "${line}" | awk '{print $2}')"
  if [[ -z "${method}" || -z "${path}" ]]; then
    echo "FAIL: bad deny endpoint format: ${line}"
    exit 1
  fi
  assert_deny_endpoint "${method}" "${path}"
done <<< "${EXPECT_DENY_ENDPOINTS}"

echo "[10/10] done (cleanup will run automatically)"
echo "PASS: RBAC full effectiveness test passed"

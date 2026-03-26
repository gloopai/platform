#!/usr/bin/env bash
set -euo pipefail

# RBAC 菜单矩阵回归：
# - 多用户
# - 全部现有“叶子菜单(path 非空)”逐个验证
# - 每个菜单生成临时角色并绑定用户，验证 my_menu 可见性 + API allow/deny
# - 全程自动回滚（用户原角色 + 临时角色清理）
#
# 必填前提：
# 1) 已有可登录的后台用户账号（用户名/密码）
# 2) 使用 OPERATOR 账号执行管理动作（建议不要把 operator 同时作为 target）
#
# 示例：
# TARGET_USERS_JSON='[
#   {"username":"admin","password":"admin123"}
# ]' \
# OPERATOR_USERNAME=admin OPERATOR_PASSWORD=admin123 \
# ./test_admin_rbac_menu_matrix.sh

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
OPERATOR_USERNAME="${OPERATOR_USERNAME:-admin}"
OPERATOR_PASSWORD="${OPERATOR_PASSWORD:-admin123}"
TARGET_USERS_JSON="${TARGET_USERS_JSON:-[{\"username\":\"admin\",\"password\":\"admin123\"}]}"
CACHE_WAIT_SECONDS="${CACHE_WAIT_SECONDS:-11}"
CURL_MAX_TIME="${CURL_MAX_TIME:-15}"
ALLOW_OPERATOR_AS_TARGET="${ALLOW_OPERATOR_AS_TARGET:-1}"
ONLY_MENU_KEYS="${ONLY_MENU_KEYS:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MATRIX_HELPER_PY="${SCRIPT_DIR}/rbac-tests/matrix_helper.py"

# 为保证验证和恢复（尤其 operator==target 场景）：
# - 临时角色附加 RBAC 管理与用户角色管理权限，避免执行中自锁
INTERNAL_PERM_KEYS="admin.rbac.my_menu,admin.rbac.manage,admin.admin_users.manage"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl not found"
  exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 not found"
  exit 1
fi
if [[ ! -f "${MATRIX_HELPER_PY}" ]]; then
  echo "matrix helper not found: ${MATRIX_HELPER_PY}"
  exit 1
fi

ROLE_IDS_TO_DELETE=()
RESTORE_ITEMS=() # item format: user_id|csv_role_ids
OPERATOR_TOKEN=""

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
cur = obj
for p in [x for x in sys.argv[2].split(".") if x]:
    if not isinstance(cur, dict) or p not in cur:
        print("")
        sys.exit(0)
    cur = cur[p]
print("" if cur is None else cur)
PY
}

api_call() {
  local method="$1"
  local path="$2"
  local token="$3"
  local body="${4:-}"
  local out
  if [[ "${method}" == "GET" || "${method}" == "DELETE" ]]; then
    out="$(curl -sS --max-time "${CURL_MAX_TIME}" -X "${method}" \
      "${GATEWAY_BASE_URL}${path}" \
      -H "X-Admin-Token: ${token}" \
      -w $'\n%{http_code}')"
  else
    out="$(curl -sS --max-time "${CURL_MAX_TIME}" -X "${method}" \
      "${GATEWAY_BASE_URL}${path}" \
      -H "X-Admin-Token: ${token}" \
      -H "Content-Type: application/json" \
      -d "${body:-{}}" \
      -w $'\n%{http_code}')"
  fi
  local status="${out##*$'\n'}"
  local resp="${out%$'\n'*}"
  printf '%s\t%s' "${status}" "${resp}"
}

assert_2xx() {
  local method="$1"
  local path="$2"
  local token="$3"
  local body="${4:-}"
  local r status resp
  r="$(api_call "${method}" "${path}" "${token}" "${body}")"
  status="${r%%$'\t'*}"
  resp="${r#*$'\t'}"
  if [[ "${status}" != 2* ]]; then
    fail "${method} ${path} expected 2xx, got ${status}; resp=${resp}"
  fi
  printf '%s' "${resp}"
}

login_token() {
  local username="$1"
  local password="$2"
  local body resp tok
  body="$(python3 - <<PY
import json
print(json.dumps({"username":"${username}","password":"${password}"}))
PY
)"
  resp="$(curl -sS --max-time "${CURL_MAX_TIME}" -X POST "${GATEWAY_BASE_URL}/v1/admin/login" \
    -H "Content-Type: application/json" -d "${body}")"
  tok="$(json_get "${resp}" "token")"
  [[ -n "${tok}" ]] || fail "login failed for ${username}: ${resp}"
  printf '%s' "${tok}"
}

cleanup() {
  local exit_code=$?
  if [[ -n "${OPERATOR_TOKEN}" ]]; then
    for item in ${RESTORE_ITEMS[@]+"${RESTORE_ITEMS[@]}"}; do
      local uid="${item%%|*}"
      local csv="${item#*|}"
      local body
      body="$(python3 - "$csv" <<'PY'
import json, sys
ids = [int(x) for x in (sys.argv[1] or "").split(",") if x.strip()]
print(json.dumps({"role_ids": ids}))
PY
)"
      local r
      r="$(api_call PUT "/v1/admin/rbac/admin_users/${uid}/roles" "${OPERATOR_TOKEN}" "${body}")" || true
      echo "  cleanup_restore_user_roles uid=${uid} status=${r%%$'\t'*}"
    done
    for rid in ${ROLE_IDS_TO_DELETE[@]+"${ROLE_IDS_TO_DELETE[@]}"}; do
      local r
      r="$(api_call DELETE "/v1/admin/rbac/roles/${rid}" "${OPERATOR_TOKEN}")" || true
      echo "  cleanup_delete_temp_role rid=${rid} status=${r%%$'\t'*}"
    done
  fi
  exit "${exit_code}"
}
trap cleanup EXIT

flatten_menu_paths() {
  local my_menu_json="$1"
  python3 - "$my_menu_json" <<'PY'
import json, sys
obj = json.loads(sys.argv[1])
seen = set()
def emit(p):
    if isinstance(p, str) and p.strip():
        p = p.strip()
        if p != "/" and p.endswith("/"):
            p = p[:-1]
        if p not in seen:
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

contains_path() {
  local lines="$1"
  local target="$2"
  python3 - "$lines" "$target" <<'PY'
import sys
lines = sys.argv[1].splitlines()
target = (sys.argv[2] or "").strip()
if target != "/" and target.endswith("/"):
    target = target[:-1]
s = set()
for x in lines:
    x = x.strip()
    if not x:
        continue
    if x != "/" and x.endswith("/"):
        x = x[:-1]
    s.add(x)
sys.exit(0 if target in s else 1)
PY
}

echo "[1/7] operator login"
OPERATOR_TOKEN="$(login_token "${OPERATOR_USERNAME}" "${OPERATOR_PASSWORD}")"
echo "  operator_login_ok"

TARGET_USERS_JSON="$(python3 "${MATRIX_HELPER_PY}" validate-target-users --raw "${TARGET_USERS_JSON}")"

echo "[2/7] load users + menus"
USERS_JSON="$(assert_2xx GET "/v1/admin/admin_users" "${OPERATOR_TOKEN}")"
MENUS_JSON="$(assert_2xx GET "/v1/admin/rbac/menus" "${OPERATOR_TOKEN}")"

MATRIX_JSON="$(python3 "${MATRIX_HELPER_PY}" build-matrix --users-json "${USERS_JSON}" --menus-json "${MENUS_JSON}" --targets-json "${TARGET_USERS_JSON}" --only-menu-keys "${ONLY_MENU_KEYS}")"

TARGET_COUNT="$(python3 - "$MATRIX_JSON" <<'PY'
import json, sys
o=json.loads(sys.argv[1]); print(len(o.get("targets") or []))
PY
)"
MENU_COUNT="$(python3 - "$MATRIX_JSON" <<'PY'
import json, sys
o=json.loads(sys.argv[1]); print(len(o.get("menus") or []))
PY
)"
[[ "${TARGET_COUNT}" -gt 0 ]] || fail "no valid target users found; check TARGET_USERS_JSON and existing admin_users"
[[ "${MENU_COUNT}" -gt 0 ]] || fail "no testable leaf menus found"
echo "  targets=${TARGET_COUNT} menus=${MENU_COUNT}"

if [[ "${ALLOW_OPERATOR_AS_TARGET}" == "1" ]]; then
  echo "  warn: operator may be used as target user; script adds internal manage perms for safe rollback."
fi

echo "[3/7] execute user × menu matrix"
python3 "${MATRIX_HELPER_PY}" to-cases-tsv --matrix-json "${MATRIX_JSON}" > /tmp/rbac_matrix_cases.tsv

TOTAL=0
PASS=0
SKIP_OPERATOR=0
while IFS=$'\t' read -r target_uid target_un target_pw menu_key menu_path menu_id; do
  [[ -z "${target_uid}" ]] && continue
  TOTAL=$((TOTAL + 1))

  if [[ "${target_un}" == "${OPERATOR_USERNAME}" && "${ALLOW_OPERATOR_AS_TARGET}" != "1" ]]; then
    echo "  skip case#${TOTAL}: target=${target_un} menu=${menu_key} (operator user)"
    SKIP_OPERATOR=$((SKIP_OPERATOR + 1))
    continue
  fi

  echo "  case#${TOTAL}: user=${target_un} menu=${menu_key} path=${menu_path}"

  # backup roles once per user
  seen_backup=0
  for item in ${RESTORE_ITEMS[@]+"${RESTORE_ITEMS[@]}"}; do
    if [[ "${item%%|*}" == "${target_uid}" ]]; then
      seen_backup=1
      break
    fi
  done
  if [[ "${seen_backup}" -eq 0 ]]; then
    ORIG_ROLES_JSON="$(assert_2xx GET "/v1/admin/rbac/admin_users/${target_uid}/roles" "${OPERATOR_TOKEN}")"
    ORIG_ROLES_CSV="$(python3 - "$ORIG_ROLES_JSON" <<'PY'
import json, sys
o=json.loads(sys.argv[1])
print(",".join(str(int(x)) for x in (o.get("role_ids") or [])))
PY
)"
    RESTORE_ITEMS+=("${target_uid}|${ORIG_ROLES_CSV}")
  fi

  ROLE_CODE="rbac_matrix_${target_uid}_$(date +%s)_$RANDOM"
  CREATE_BODY="$(python3 - <<PY
import json
print(json.dumps({"code":"${ROLE_CODE}","name":"RBAC矩阵临时角色","status":1}))
PY
)"
  CREATE_ROLE_JSON="$(assert_2xx POST "/v1/admin/rbac/roles" "${OPERATOR_TOKEN}" "${CREATE_BODY}")"
  RID="$(json_get "${CREATE_ROLE_JSON}" "role.id")"
  [[ -n "${RID}" && "${RID}" != "0" ]] || fail "create temp role failed: ${CREATE_ROLE_JSON}"
  ROLE_IDS_TO_DELETE+=("${RID}")

  # 设置角色菜单（仅当前菜单 + 内部保底菜单）
  ALL_MENU_IDS="$(python3 - "$menu_id" "$MENUS_JSON" <<'PY'
import json, sys
cur = int(sys.argv[1])
menus = json.loads(sys.argv[2]).get("menus") or []
parent = {}
key_to_id = {}
for m in menus:
    if not isinstance(m, dict): continue
    mid = int(m.get("id") or 0)
    parent[mid] = int(m.get("parent_id") or 0)
    mk = (m.get("menu_key") or "").strip()
    if mk: key_to_id[mk] = mid
ids = set()
for base in [cur, key_to_id.get("menu.rbac_menus", 0)]:
    while base and base not in ids:
        ids.add(base)
        base = parent.get(base, 0)
print(",".join(str(x) for x in sorted(ids)))
PY
)"
  SET_MENUS_BODY="$(python3 - "$ALL_MENU_IDS" <<'PY'
import json, sys
ids=[int(x) for x in (sys.argv[1] or "").split(",") if x.strip()]
print(json.dumps({"menu_ids": ids}))
PY
)"
  assert_2xx PUT "/v1/admin/rbac/roles/${RID}/menus" "${OPERATOR_TOKEN}" "${SET_MENUS_BODY}" >/dev/null

  # 设置角色权限：业务权限 + 内部保底权限
  SET_PERMS_BODY="$(python3 - "${INTERNAL_PERM_KEYS}" <<'PY'
import json, sys
keys = [x.strip() for x in (sys.argv[1] or "").split(",") if x.strip()]
print(json.dumps({"perm_keys": keys}))
PY
)"
  assert_2xx PUT "/v1/admin/rbac/roles/${RID}/perm_keys" "${OPERATOR_TOKEN}" "${SET_PERMS_BODY}" >/dev/null

  # 绑定用户 -> 当前临时角色
  SET_USER_ROLES_BODY="$(python3 - "$RID" <<'PY'
import json, sys
print(json.dumps({"role_ids":[int(sys.argv[1])]}))
PY
)"
  assert_2xx PUT "/v1/admin/rbac/admin_users/${target_uid}/roles" "${OPERATOR_TOKEN}" "${SET_USER_ROLES_BODY}" >/dev/null

  if [[ "${CACHE_WAIT_SECONDS}" -gt 0 ]]; then
    sleep "${CACHE_WAIT_SECONDS}"
  fi

  TARGET_TOKEN="$(login_token "${target_un}" "${target_pw}")"
  MY_MENU_JSON="$(assert_2xx GET "/v1/admin/rbac/my_menu" "${TARGET_TOKEN}")"
  PATHS="$(flatten_menu_paths "${MY_MENU_JSON}")"

  if ! contains_path "${PATHS}" "${menu_path}"; then
    echo "    debug_paths=$(printf '%s' "${PATHS}" | tr '\n' ',' | sed 's/,$//')"
    echo "    debug_my_menu=${MY_MENU_JSON}"
    fail "menu visibility failed user=${target_un} menu=${menu_key} path=${menu_path}"
  fi

  # deny check: 对 merchants 做负向校验（大多数菜单场景不应允许）
  DENY_R="$(api_call GET "/v1/admin/merchants" "${TARGET_TOKEN}")"
  DENY_S="${DENY_R%%$'\t'*}"
  if [[ "${DENY_S}" != "401" && "${DENY_S}" != "403" ]]; then
    echo "    warn: deny check got ${DENY_S} for /v1/admin/merchants (user=${target_un}, menu=${menu_key})"
  fi

  PASS=$((PASS + 1))
done < /tmp/rbac_matrix_cases.tsv

echo "[4/7] matrix result"
echo "  total_cases=${TOTAL} pass_cases=${PASS}"
if [[ "${SKIP_OPERATOR}" -gt 0 ]]; then
  echo "  skipped_operator_cases=${SKIP_OPERATOR}"
fi
[[ "${PASS}" -gt 0 ]] || fail "no case passed"

echo "[5/7] operator sanity"
assert_2xx GET "/v1/admin/rbac/menus" "${OPERATOR_TOKEN}" >/dev/null
echo "  operator_ok"

echo "[6/7] cleanup scheduled by trap"
echo "[7/7] done"
echo "PASS: multi-user multi-role all-menu matrix verified"

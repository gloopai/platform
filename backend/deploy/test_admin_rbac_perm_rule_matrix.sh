#!/usr/bin/env bash
set -euo pipefail

# RBAC 权限点 + 接口规则矩阵回归
# - 按 perm_key 分组，校验该权限下所有接口规则在“已授权”时不再被 RBAC 拦截（非 401/403）
# - 同时抽样一个“非当前权限”的接口规则，校验在“未授权”时被 RBAC 拒绝（401/403）
#
# 说明：
# - 本脚本关注“RBAC 是否生效”，不关注业务参数正确性
# - 所以 allow 判定采用：HTTP != 401/403 即通过（常见为 200/400/404/422）
#
# 依赖：
# - 使用 operator 执行管理动作
# - target 用户用于承载被测角色

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
OPERATOR_USERNAME="${OPERATOR_USERNAME:-admin}"
OPERATOR_PASSWORD="${OPERATOR_PASSWORD:-admin123}"
TARGET_USERS_JSON="${TARGET_USERS_JSON:-[{\"username\":\"admin\",\"password\":\"admin123\"}]}"
CACHE_WAIT_SECONDS="${CACHE_WAIT_SECONDS:-11}"
CURL_MAX_TIME="${CURL_MAX_TIME:-15}"
ALLOW_OPERATOR_AS_TARGET="${ALLOW_OPERATOR_AS_TARGET:-1}"
ONLY_PERM_KEYS="${ONLY_PERM_KEYS:-}"
MAX_PERM_CASES="${MAX_PERM_CASES:-0}" # 0=all

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MATRIX_HELPER_PY="${SCRIPT_DIR}/rbac-tests/matrix_helper.py"

# 为避免执行中自锁，临时角色固定附加管理权限
INTERNAL_PERM_KEYS="admin.rbac.manage,admin.admin_users.manage,admin.rbac.my_menu"

ROLE_IDS_TO_DELETE=()
RESTORE_ITEMS=() # user_id|csv_role_ids
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
  resp="$(curl -sS --max-time "${CURL_MAX_TIME}" -X POST "${GATEWAY_BASE_URL}/v1/admin/login" -H "Content-Type: application/json" -d "${body}")"
  tok="$(json_get "${resp}" "token")"
  [[ -n "${tok}" ]] || fail "login failed for ${username}: ${resp}"
  printf '%s' "${tok}"
}

cleanup() {
  local ec=$?
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
  exit "${ec}"
}
trap cleanup EXIT

resolve_rule_path() {
  local path_pattern="$1"
  python3 - "$path_pattern" <<'PY'
import re, sys
p = (sys.argv[1] or "").strip()
if not p.startswith("/"):
    p = "/" + p
out = re.sub(r":[A-Za-z_][A-Za-z0-9_]*", "1", p)
print(out)
PY
}

echo "[1/8] preflight"
command -v curl >/dev/null 2>&1 || fail "curl not found"
command -v python3 >/dev/null 2>&1 || fail "python3 not found"
[[ -f "${MATRIX_HELPER_PY}" ]] || fail "matrix helper not found: ${MATRIX_HELPER_PY}"

echo "[2/8] operator login"
OPERATOR_TOKEN="$(login_token "${OPERATOR_USERNAME}" "${OPERATOR_PASSWORD}")"
echo "  operator_login_ok"

TARGET_USERS_JSON="$(python3 "${MATRIX_HELPER_PY}" validate-target-users --raw "${TARGET_USERS_JSON}")"

echo "[3/8] load users / permissions / api rules"
USERS_JSON="$(assert_2xx GET "/v1/admin/admin_users" "${OPERATOR_TOKEN}")"
PERMS_JSON="$(assert_2xx GET "/v1/admin/rbac/permissions" "${OPERATOR_TOKEN}")"
RULES_JSON="$(assert_2xx GET "/v1/admin/rbac/api_rules" "${OPERATOR_TOKEN}")"
MENUS_JSON="$(assert_2xx GET "/v1/admin/rbac/menus" "${OPERATOR_TOKEN}")"

python3 - "$USERS_JSON" "$TARGET_USERS_JSON" "$PERMS_JSON" "$RULES_JSON" "$ONLY_PERM_KEYS" "$MAX_PERM_CASES" > /tmp/rbac_perm_cases.tsv <<'PY'
import json, sys
users = json.loads(sys.argv[1]).get("users") or []
targets_raw = json.loads(sys.argv[2]) or []
perms = json.loads(sys.argv[3]).get("permissions") or []
rules = json.loads(sys.argv[4]).get("rules") or []
only = {x.strip() for x in (sys.argv[5] or "").split(",") if x.strip()}
max_cases = int(sys.argv[6] or 0)

user_map = {u.get("username"): int(u.get("id") or 0) for u in users if isinstance(u, dict)}
targets = []
for t in targets_raw:
    if not isinstance(t, dict):
        continue
    un = str(t.get("username") or "").strip()
    pw = str(t.get("password") or "").strip()
    uid = user_map.get(un, 0)
    if un and pw and uid > 0:
        targets.append((uid, un, pw))

perm_set = {str(p.get("perm_key") or "").strip() for p in perms if isinstance(p, dict)}
group = {}
for r in rules:
    if not isinstance(r, dict):
        continue
    if int(r.get("status") or 0) != 1:
        continue
    pk = str(r.get("perm_key") or "").strip()
    method = str(r.get("method") or "").strip().upper()
    path = str(r.get("path_pattern") or "").strip()
    if not pk or not method or not path:
        continue
    if only and pk not in only:
        continue
    if pk not in perm_set:
        continue
    group.setdefault(pk, []).append((method, path))

perm_keys = sorted(group.keys())
if max_cases > 0:
    perm_keys = perm_keys[:max_cases]

for uid, un, pw in targets:
    for pk in perm_keys:
        # target_user_id,username,password,perm_key,method,path_pattern
        for method, path in group[pk]:
            print("\t".join([str(uid), un, pw, pk, method, path]))
PY

CASE_COUNT="$(wc -l < /tmp/rbac_perm_cases.tsv | tr -d ' ')"
[[ "${CASE_COUNT}" -gt 0 ]] || fail "no perm/rule cases generated"
echo "  cases=${CASE_COUNT}"

echo "[4/8] prepare menu id set for temp role"
RBAC_MENU_IDS="$(python3 - "$MENUS_JSON" <<'PY'
import json, sys
menus = json.loads(sys.argv[1]).get("menus") or []
key_to = {}
parent = {}
for m in menus:
    if not isinstance(m, dict):
        continue
    mid = int(m.get("id") or 0)
    mk = str(m.get("menu_key") or "").strip()
    parent[mid] = int(m.get("parent_id") or 0)
    if mk:
        key_to[mk] = mid
seed = []
for k in ["menu.rbac_menus", "menu.rbac_roles", "menu.rbac_features", "menu.rbac_api_rules", "menu.rbac_overview"]:
    if key_to.get(k):
        seed.append(key_to[k])
ids = set()
for x in seed:
    while x and x not in ids:
        ids.add(x)
        x = parent.get(x, 0)
print(",".join(str(i) for i in sorted(ids)))
PY
)"
[[ -n "${RBAC_MENU_IDS}" ]] || fail "cannot resolve RBAC menu ids"

echo "[5/8] execute perm/rule matrix"
PASS=0
TOTAL=0
LAST_USER_ID=""
LAST_USER_TOKEN=""

while IFS=$'\t' read -r uid un pw pk method pattern; do
  [[ -z "${uid}" || -z "${pk}" || -z "${method}" || -z "${pattern}" ]] && continue
  TOTAL=$((TOTAL + 1))

  if [[ "${un}" == "${OPERATOR_USERNAME}" && "${ALLOW_OPERATOR_AS_TARGET}" != "1" ]]; then
    continue
  fi

  # backup original roles once per user
  seen=0
  for item in ${RESTORE_ITEMS[@]+"${RESTORE_ITEMS[@]}"}; do
    [[ "${item%%|*}" == "${uid}" ]] && seen=1 && break
  done
  if [[ "${seen}" -eq 0 ]]; then
    ORIG_ROLES_JSON="$(assert_2xx GET "/v1/admin/rbac/admin_users/${uid}/roles" "${OPERATOR_TOKEN}")"
    ORIG_CSV="$(python3 - "$ORIG_ROLES_JSON" <<'PY'
import json, sys
o=json.loads(sys.argv[1]); print(",".join(str(int(x)) for x in (o.get("role_ids") or [])))
PY
)"
    RESTORE_ITEMS+=("${uid}|${ORIG_CSV}")
  fi

  ROLE_CODE="rbac_perm_${uid}_$(date +%s)_$RANDOM"
  CREATE_BODY="$(python3 - <<PY
import json
print(json.dumps({"code":"${ROLE_CODE}","name":"RBAC权限规则临时角色","status":1}))
PY
)"
  CREATE_JSON="$(assert_2xx POST "/v1/admin/rbac/roles" "${OPERATOR_TOKEN}" "${CREATE_BODY}")"
  RID="$(json_get "${CREATE_JSON}" "role.id")"
  [[ -n "${RID}" && "${RID}" != "0" ]] || fail "create role failed for perm=${pk}"
  ROLE_IDS_TO_DELETE+=("${RID}")

  SET_MENU_BODY="$(python3 - "${RBAC_MENU_IDS}" <<'PY'
import json, sys
ids=[int(x) for x in (sys.argv[1] or "").split(",") if x.strip()]
print(json.dumps({"menu_ids": ids}))
PY
)"
  assert_2xx PUT "/v1/admin/rbac/roles/${RID}/menus" "${OPERATOR_TOKEN}" "${SET_MENU_BODY}" >/dev/null

  SET_PERM_BODY="$(python3 - "${pk}" "${INTERNAL_PERM_KEYS}" <<'PY'
import json, sys
main = (sys.argv[1] or "").strip()
extra = [x.strip() for x in (sys.argv[2] or "").split(",") if x.strip()]
keys = [main] + [x for x in extra if x != main]
print(json.dumps({"perm_keys": keys}))
PY
)"
  assert_2xx PUT "/v1/admin/rbac/roles/${RID}/perm_keys" "${OPERATOR_TOKEN}" "${SET_PERM_BODY}" >/dev/null

  SET_USER_BODY="$(python3 - "${RID}" <<'PY'
import json, sys
print(json.dumps({"role_ids":[int(sys.argv[1])]}))
PY
)"
  assert_2xx PUT "/v1/admin/rbac/admin_users/${uid}/roles" "${OPERATOR_TOKEN}" "${SET_USER_BODY}" >/dev/null

  sleep "${CACHE_WAIT_SECONDS}"

  if [[ "${LAST_USER_ID}" != "${uid}" ]]; then
    LAST_USER_TOKEN="$(login_token "${un}" "${pw}")"
    LAST_USER_ID="${uid}"
  fi

  path="$(resolve_rule_path "${pattern}")"
  r="$(api_call "${method}" "${path}" "${LAST_USER_TOKEN}" "{}")"
  status="${r%%$'\t'*}"
  if [[ "${status}" == "401" || "${status}" == "403" ]]; then
    fail "allow check blocked by RBAC: user=${un} perm=${pk} rule=${method} ${pattern} status=${status}"
  fi

  # 抽样一条其他权限规则做 deny 检查（如果存在）
  other="$(python3 - "$RULES_JSON" "$pk" <<'PY'
import json, sys
rules = json.loads(sys.argv[1]).get("rules") or []
cur = sys.argv[2]
for r in rules:
    if not isinstance(r, dict):
        continue
    if int(r.get("status") or 0) != 1:
        continue
    pk = str(r.get("perm_key") or "").strip()
    m = str(r.get("method") or "").strip().upper()
    p = str(r.get("path_pattern") or "").strip()
    if not pk or not m or not p:
        continue
    if pk != cur and pk not in {"admin.rbac.manage","admin.admin_users.manage","admin.rbac.my_menu"}:
        print(m + "\t" + p)
        break
PY
)"
  if [[ -n "${other}" ]]; then
    om="${other%%$'\t'*}"
    op="${other#*$'\t'}"
    opath="$(resolve_rule_path "${op}")"
    rr="$(api_call "${om}" "${opath}" "${LAST_USER_TOKEN}" "{}")"
    st="${rr%%$'\t'*}"
    if [[ "${st}" != "401" && "${st}" != "403" ]]; then
      fail "deny check failed: user=${un} perm=${pk} unexpectedly got ${st} for ${om} ${op}"
    fi
  fi

  PASS=$((PASS + 1))
  echo "  case#${TOTAL} ok user=${un} perm=${pk} rule=${method} ${pattern}"
done < /tmp/rbac_perm_cases.tsv

echo "[6/8] result"
echo "  total_cases=${TOTAL} pass_cases=${PASS}"
[[ "${PASS}" -gt 0 ]] || fail "no case passed"

echo "[7/8] operator sanity"
assert_2xx GET "/v1/admin/rbac/menus" "${OPERATOR_TOKEN}" >/dev/null
echo "  operator_ok"

echo "[8/8] done (cleanup by trap)"
echo "PASS: perm_key + api_rules matrix verified"

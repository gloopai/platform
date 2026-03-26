#!/usr/bin/env bash
set -euo pipefail

# RBAC 生效性验证脚本（菜单可见 + 接口可访问）
#
# 用法（最小）：
#   GATEWAY_BASE_URL=http://127.0.0.1:8080 \
#   ADMIN_USERNAME=admin ADMIN_PASSWORD=admin123 \
#   ./test_admin_rbac_effective.sh
#
# 可选环境变量：
#   EXPECT_ALLOW_MENU_PATHS="/rbac/menus,/rbac/roles"
#   EXPECT_DENY_MENU_PATHS="/merchants,/channels"
#   EXPECT_ALLOW_ENDPOINTS=$'GET /v1/admin/rbac/menus\nGET /v1/admin/rbac/roles'
#   EXPECT_DENY_ENDPOINTS=$'GET /v1/admin/merchants\nGET /v1/admin/channels'
#
# 说明：
# - allow endpoint 期望 HTTP 2xx
# - deny endpoint 期望 HTTP 401/403

GATEWAY_BASE_URL="${GATEWAY_BASE_URL:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"

EXPECT_ALLOW_MENU_PATHS="${EXPECT_ALLOW_MENU_PATHS:-}"
EXPECT_DENY_MENU_PATHS="${EXPECT_DENY_MENU_PATHS:-}"

DEFAULT_ALLOW_ENDPOINTS=$'GET /v1/admin/rbac/my_menu\nGET /v1/admin/rbac/menus\nGET /v1/admin/rbac/permissions\nGET /v1/admin/rbac/api_rules'
DEFAULT_DENY_ENDPOINTS=""
EXPECT_ALLOW_ENDPOINTS="${EXPECT_ALLOW_ENDPOINTS:-$DEFAULT_ALLOW_ENDPOINTS}"
EXPECT_DENY_ENDPOINTS="${EXPECT_DENY_ENDPOINTS:-$DEFAULT_DENY_ENDPOINTS}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl not found"
  exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 not found"
  exit 1
fi

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

split_csv_to_lines() {
  local csv="$1"
  python3 - "$csv" <<'PY'
import sys
s = (sys.argv[1] or "").strip()
if not s:
    sys.exit(0)
for part in s.split(","):
    p = part.strip()
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
    if isinstance(p, str) and p.strip():
        if p not in seen:
            seen.add(p)
            print(p)

for item in (obj.get("sidebar") or []):
    if not isinstance(item, dict):
        continue
    emit(item.get("to"))
    for ch in (item.get("children") or []):
        if isinstance(ch, dict):
            emit(ch.get("to"))
for item in (obj.get("avatar_links") or []):
    if isinstance(item, dict):
        emit(item.get("to"))
PY
}

require_path_present() {
  local all_paths="$1"
  local target="$2"
  if ! printf '%s\n' "${all_paths}" | python3 - "$target" <<'PY'
import sys
target = sys.argv[1]
paths = {line.strip() for line in sys.stdin if line.strip()}
sys.exit(0 if target in paths else 1)
PY
  then
    echo "FAIL: expected menu path visible but missing: ${target}"
    exit 1
  fi
  echo "  menu_allow_ok ${target}"
}

require_path_absent() {
  local all_paths="$1"
  local target="$2"
  if printf '%s\n' "${all_paths}" | python3 - "$target" <<'PY'
import sys
target = sys.argv[1]
paths = {line.strip() for line in sys.stdin if line.strip()}
sys.exit(0 if target in paths else 1)
PY
  then
    echo "FAIL: expected menu path hidden but still visible: ${target}"
    exit 1
  fi
  echo "  menu_deny_ok ${target}"
}

call_admin_api() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local output
  if [[ "${method}" == "GET" || "${method}" == "DELETE" ]]; then
    output="$(curl -sS -X "${method}" \
      "${GATEWAY_BASE_URL}${path}" \
      -H "X-Admin-Token: ${ADMIN_TOKEN}" \
      -w $'\n%{http_code}')"
  else
    local payload="${body:-{}}"
    output="$(curl -sS -X "${method}" \
      "${GATEWAY_BASE_URL}${path}" \
      -H "X-Admin-Token: ${ADMIN_TOKEN}" \
      -H "Content-Type: application/json" \
      -d "${payload}" \
      -w $'\n%{http_code}')"
  fi
  local status="${output##*$'\n'}"
  local resp="${output%$'\n'*}"
  printf '%s\t%s' "${status}" "${resp}"
}

assert_allow_endpoint() {
  local method="$1"
  local path="$2"
  local result status
  result="$(call_admin_api "${method}" "${path}")"
  status="${result%%$'\t'*}"
  if [[ "${status}" != 2* ]]; then
    echo "FAIL: expect ALLOW but got ${status} for ${method} ${path}"
    echo "  resp=${result#*$'\t'}"
    exit 1
  fi
  echo "  api_allow_ok ${method} ${path} -> ${status}"
}

assert_deny_endpoint() {
  local method="$1"
  local path="$2"
  local result status
  result="$(call_admin_api "${method}" "${path}")"
  status="${result%%$'\t'*}"
  if [[ "${status}" != "401" && "${status}" != "403" ]]; then
    echo "FAIL: expect DENY(401/403) but got ${status} for ${method} ${path}"
    echo "  resp=${result#*$'\t'}"
    exit 1
  fi
  echo "  api_deny_ok ${method} ${path} -> ${status}"
}

echo "[1/6] admin login"
LOGIN_BODY="$(python3 - <<PY
import json
print(json.dumps({"username":"${ADMIN_USERNAME}","password":"${ADMIN_PASSWORD}"}))
PY
)"
LOGIN_RESP="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/admin/login" -H "Content-Type: application/json" -d "${LOGIN_BODY}")"
ADMIN_TOKEN="$(json_get "${LOGIN_RESP}" "token")"
if [[ -z "${ADMIN_TOKEN}" ]]; then
  echo "FAIL: admin login failed: ${LOGIN_RESP}"
  exit 1
fi
echo "  login_ok"

echo "[2/6] basic RBAC config endpoint sanity"
INVALID_PERM_RESP="$(curl -sS -X POST "${GATEWAY_BASE_URL}/v1/admin/rbac/permissions" \
  -H "X-Admin-Token: ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}' \
  -w $'\n%{http_code}')"
INVALID_PERM_STATUS="${INVALID_PERM_RESP##*$'\n'}"
if [[ "${INVALID_PERM_STATUS}" != "400" ]]; then
  echo "FAIL: expect 400 for invalid permission payload, got ${INVALID_PERM_STATUS}"
  echo "  resp=${INVALID_PERM_RESP%$'\n'*}"
  exit 1
fi
echo "  invalid_payload_check_ok /v1/admin/rbac/permissions -> 400"

echo "[3/6] fetch my_menu"
MY_MENU_RESP="$(curl -sS "${GATEWAY_BASE_URL}/v1/admin/rbac/my_menu" -H "X-Admin-Token: ${ADMIN_TOKEN}")"
MENU_PATHS="$(flatten_menu_paths "${MY_MENU_RESP}")"
if [[ -z "${MENU_PATHS}" ]]; then
  echo "WARN: my_menu has no visible paths (may be expected for fully restricted role)"
fi

echo "[4/6] verify expected menu visibility"
if [[ -n "${EXPECT_ALLOW_MENU_PATHS}" ]]; then
  while IFS= read -r p; do
    [[ -z "${p}" ]] && continue
    require_path_present "${MENU_PATHS}" "${p}"
  done < <(split_csv_to_lines "${EXPECT_ALLOW_MENU_PATHS}")
else
  echo "  skip (EXPECT_ALLOW_MENU_PATHS not set)"
fi

if [[ -n "${EXPECT_DENY_MENU_PATHS}" ]]; then
  while IFS= read -r p; do
    [[ -z "${p}" ]] && continue
    require_path_absent "${MENU_PATHS}" "${p}"
  done < <(split_csv_to_lines "${EXPECT_DENY_MENU_PATHS}")
else
  echo "  skip (EXPECT_DENY_MENU_PATHS not set)"
fi

echo "[5/6] verify expected API ALLOW matrix"
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

echo "[6/6] verify expected API DENY matrix"
if [[ -n "${EXPECT_DENY_ENDPOINTS}" ]]; then
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
else
  echo "  skip (EXPECT_DENY_ENDPOINTS not set)"
fi

echo
echo "PASS: RBAC permission effectiveness verified"

#!/usr/bin/env bash
set -euo pipefail

# 一键跑三层 RBAC 检查：
# 1) 菜单层（多用户 × 全菜单）：test_admin_rbac_menu_matrix.sh
# 2) 功能点 + 接口规则层：test_admin_rbac_perm_rule_matrix.sh
#
# 透传同名环境变量到子脚本（operator/target/cache/base_url 等）

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "[full-stack 1/2] menu matrix"
"${SCRIPT_DIR}/test_admin_rbac_menu_matrix.sh"

echo "[full-stack 2/2] perm+rule matrix"
"${SCRIPT_DIR}/test_admin_rbac_perm_rule_matrix.sh"

echo "PASS: RBAC full-stack checks completed"

#!/usr/bin/env bash
set -euo pipefail
# 初始化 MySQL：建表 + 脚手架演示数据（管理员 / RBAC / global_settings）
# 使用前请设置 MYSQL_PWD 或编辑下方连接参数。

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_DIR="${ROOT}/sql"

MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_DB="${MYSQL_DB:-platform}"

MYSQL_ARGS=( -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" )
if [ -n "${MYSQL_PWD:-}" ]; then
  MYSQL_ARGS+=( -p"${MYSQL_PWD}" )
fi

echo "Applying schema to ${MYSQL_USER}@${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB} ..."
mysql "${MYSQL_ARGS[@]}" "${MYSQL_DB}" <"${SQL_DIR}/schema.sql"

echo "Applying seed_demo.sql ..."
mysql "${MYSQL_ARGS[@]}" "${MYSQL_DB}" <"${SQL_DIR}/seed_demo.sql"

echo "Done. Default admin: admin / admin123"

#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SQL_DIR="${ROOT_DIR}/deploy/sql"

MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASS="${MYSQL_PASS:-your_password}"
MYSQL_DB="${MYSQL_DB:-pay}"

if ! command -v mysql >/dev/null 2>&1; then
  echo "mysql client not found"
  exit 1
fi

mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASS}" \
  -e "CREATE DATABASE IF NOT EXISTS \`${MYSQL_DB}\` DEFAULT CHARSET=utf8mb4;"

for sql_file in "${SQL_DIR}"/migration_*.sql; do
  if [[ -f "${sql_file}" ]]; then
    echo "apply: $(basename "${sql_file}")"
    mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASS}" "${MYSQL_DB}" < "${sql_file}"
  fi
done

echo "ok: migrations applied to ${MYSQL_USER}@${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB}"

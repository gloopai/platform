#!/usr/bin/env bash
# 代收/代付 DB 不变量检查。**不依赖 mysql 客户端**，通过 Go 驱动连库。
#
# 环境变量（与 gateway 一致）：
#   PAY_PLATFORM_MYSQL_DSN — 优先；与 gateway Mysql.DataSource 相同即可
#   MYSQL_HOST MYSQL_USER MYSQL_PASSWORD MYSQL_DB MYSQL_PORT
#   STUCK_PAYOUT_MINUTES — 代付 pending 超时告警分钟数，0 关闭（默认 180）
#
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/services/gateway"
exec go run ./cmd/check-pay-invariants "$@"

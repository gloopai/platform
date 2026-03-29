#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${ROOT_DIR}/.dev-logs"
mkdir -p "${LOG_DIR}"

PIDS=()
NAMES=()

cleanup() {
  for i in "${!PIDS[@]}"; do
    pid="${PIDS[$i]}"
    if kill -0 "${pid}" >/dev/null 2>&1; then
      kill "${pid}" >/dev/null 2>&1 || true
    fi
  done
  for i in "${!PIDS[@]}"; do
    pid="${PIDS[$i]}"
    if kill -0 "${pid}" >/dev/null 2>&1; then
      kill -9 "${pid}" >/dev/null 2>&1 || true
    fi
  done
}

trap cleanup EXIT INT TERM

is_listening() {
  local port="$1"
  lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
}

print_url() {
  local name="$1"
  local url="$2"
  printf "  %-12s %s\n" "${name}:" "${url}"
}

start_bg() {
  local name="$1"
  local cwd="$2"
  local cmd="$3"
  local logfile="${LOG_DIR}/${name}.log"

  (
    cd "${cwd}"
    exec bash -lc "${cmd}"
  ) >"${logfile}" 2>&1 &

  local pid="$!"
  PIDS+=("${pid}")
  NAMES+=("${name}")

  sleep 0.2
  if ! kill -0 "${pid}" >/dev/null 2>&1; then
    echo "[${name}] failed to start. logs: ${logfile}"
    tail -n 80 "${logfile}" || true
    exit 1
  fi

  echo "[${name}] pid=${pid} logs=${logfile}"
}

export CGO_ENABLED=0

if ! is_listening 8500; then
  if command -v consul >/dev/null 2>&1; then
    start_bg "consul" "${ROOT_DIR}" "consul agent -dev -client=127.0.0.1 -bind=127.0.0.1 -ui -log-level=warn"
    sleep 0.8
  else
    echo "[consul] not running on :8500 and consul binary not found"
  fi
fi

if ! is_listening 6379; then
  if command -v redis-server >/dev/null 2>&1; then
    start_bg "redis" "${ROOT_DIR}" "redis-server --port 6379"
    sleep 0.3
  else
    echo "[redis] not running on :6379 and redis-server binary not found"
  fi
fi

if ! is_listening 4150; then
  if command -v nsqlookupd >/dev/null 2>&1 && command -v nsqd >/dev/null 2>&1; then
    start_bg "nsqlookupd" "${ROOT_DIR}" "nsqlookupd -tcp-address=127.0.0.1:4160 -http-address=127.0.0.1:4161"
    sleep 0.3
    start_bg "nsqd" "${ROOT_DIR}" "nsqd -tcp-address=127.0.0.1:4150 -http-address=127.0.0.1:4151 -lookupd-tcp-address=127.0.0.1:4160"
    sleep 0.6
  else
    echo "[nsq] not running on :4150 and nsq binaries not found"
  fi
fi

if ! is_listening 3306; then
  echo "[mysql] not listening on :3306 (services may fail)"
fi

start_bg "trade" "${ROOT_DIR}/backend/services/trade" "go run . -f etc/trade.yaml"
start_bg "core" "${ROOT_DIR}/backend/services/core" "go run . -f etc/core.yaml"
start_bg "service-hub" "${ROOT_DIR}/backend/services/service-hub" "go run . -f etc/service-hub.yaml"
start_bg "gateway" "${ROOT_DIR}/backend/services/gateway" "go run . -f etc/gateway-api.yaml"
start_bg "notice-consumer" "${ROOT_DIR}/backend/services/notice-consumer" "go run . -f etc/notice-consumer.yaml"

if [ -f "${ROOT_DIR}/frontend/package.json" ]; then
  if ! is_listening 5176; then
    start_bg "fe-admin" "${ROOT_DIR}/frontend" "npm run dev"
  fi
fi

echo "running. logs: ${LOG_DIR}"
echo "urls:"
echo "  gateway（scaffold/platform-admin：仅 Admin HTTP）:"
print_url "gateway-admin" "http://127.0.0.1:8080/  (Admin: /v1/admin/*)"
print_url "service-hub" "grpc://127.0.0.1:8094 (Consul: payment.rpc.service-hub)"
print_url "admin" "http://127.0.0.1:5176/"
echo "db init:"
echo "  bash backend/deploy/init_demo.sh"
echo "demo accounts:"
echo "  admin:   admin / admin123"
echo "  merchant:m_demo / demo_secret"
for i in "${!PIDS[@]}"; do
  echo "  ${NAMES[$i]}: ${PIDS[$i]}"
done

wait

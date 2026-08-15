#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"
load_env

duration="${1:-120}"
interval="${2:-2}"
[[ "$duration" =~ ^[0-9]+$ && "$interval" =~ ^[0-9]+$ && "$interval" -gt 0 ]] || die "usage: capture-runtime-stats.sh [duration-seconds] [interval-seconds]"
mkdir -p "${DEPLOY_ROOT}/state"
output="${DEPLOY_ROOT}/state/runtime-stats-$(date -u +%Y%m%dT%H%M%SZ).csv"
printf 'timestamp,container,cpu_percent,memory_usage,network_io,block_io\n' > "$output"
end=$((SECONDS + duration))
while (( SECONDS < end )); do
  timestamp="$(date -u +%FT%TZ)"
  docker stats --no-stream --format "${timestamp},{{.Name}},{{.CPUPerc}},{{.MemUsage}},{{.NetIO}},{{.BlockIO}}" cs2match-nakama cs2match-db >> "$output"
  sleep "$interval"
done
log "runtime stats recorded: $output"

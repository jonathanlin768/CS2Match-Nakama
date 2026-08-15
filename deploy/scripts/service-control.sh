#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"
load_env
case "${1:-status}" in
  close) compose stop cloudflared nakama; log "public API and application stopped; database volume and instance billing remain" ;;
  open) compose up -d db nakama cloudflared ;;
  status) compose ps ;;
  *) die "usage: service-control.sh open|close|status" ;;
esac

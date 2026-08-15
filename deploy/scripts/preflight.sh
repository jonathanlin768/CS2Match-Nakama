#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"

skip_permissions=false
backend_image_override=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-permissions) skip_permissions=true ;;
    --backend-image)
      shift
      [[ $# -gt 0 ]] || die "--backend-image requires a value"
      backend_image_override="$1"
      ;;
    *) die "usage: preflight.sh [--skip-permissions] [--backend-image IMAGE@sha256:DIGEST]" ;;
  esac
  shift
done
load_env
if [[ -n "$backend_image_override" ]]; then
  BACKEND_IMAGE_REF="$backend_image_override"
  export BACKEND_IMAGE_REF
fi

required=(COMPOSE_PROJECT_NAME BACKEND_IMAGE_REF POSTGRES_IMAGE CLOUDFLARED_IMAGE_REF DB_NAME DB_USER DB_PASSWORD NAKAMA_SERVER_KEY CONSOLE_USERNAME CONSOLE_PASSWORD CONSOLE_SIGNING_KEY SESSION_ENCRYPTION_KEY SESSION_REFRESH_ENCRYPTION_KEY RUNTIME_HTTP_KEY API_HOST FRONTEND_ORIGIN TAILSCALE_IP CLOUDFLARE_TUNNEL_TOKEN RESTIC_IMAGE RESTIC_REPOSITORY RESTIC_PASSWORD AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_DEFAULT_REGION)
for name in "${required[@]}"; do
  [[ -n "${!name:-}" ]] || die "$name is required"
  case "${!name}" in *REPLACE*|*replace*|*example.com*|*changeme*|*CHANGE_ME*) die "$name still contains a placeholder";; esac
done

[[ "$BACKEND_IMAGE_REF" =~ ^ghcr\.io/.+@sha256:[0-9a-f]{64}$ ]] || die "BACKEND_IMAGE_REF must be a GHCR sha256 digest"
[[ "$CLOUDFLARED_IMAGE_REF" == *@sha256:* ]] || die "CLOUDFLARED_IMAGE_REF must use an image digest"
[[ "$POSTGRES_IMAGE" != *latest* && "$RESTIC_IMAGE" != *latest* ]] || die "production images must not use latest"
[[ "$API_HOST" != http* && "$API_HOST" != */* ]] || die "API_HOST must be a hostname only"
[[ "$FRONTEND_ORIGIN" =~ ^https://[^/]+$ ]] || die "FRONTEND_ORIGIN must be one HTTPS origin"
[[ "$TAILSCALE_IP" =~ ^100\.[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "TAILSCALE_IP must be the host's private Tailscale IPv4 address"
[[ "$NAKAMA_SERVER_KEY" != defaultkey && "$DB_PASSWORD" != localpass ]] || die "development defaults are forbidden"
[[ "$CONSOLE_USERNAME" != admin ]] || die "the default Console username is forbidden"
[[ ${#SESSION_ENCRYPTION_KEY} -ge 32 && ${#SESSION_REFRESH_ENCRYPTION_KEY} -ge 32 && ${#RUNTIME_HTTP_KEY} -ge 32 && ${#CONSOLE_SIGNING_KEY} -ge 32 ]] || die "session, refresh, runtime and Console signing keys must be at least 32 characters"
[[ "$SESSION_ENCRYPTION_KEY" != "$SESSION_REFRESH_ENCRYPTION_KEY" && "$SESSION_ENCRYPTION_KEY" != "$RUNTIME_HTTP_KEY" && "$SESSION_REFRESH_ENCRYPTION_KEY" != "$RUNTIME_HTTP_KEY" ]] || die "session, refresh and runtime keys must be distinct"

if [[ "$skip_permissions" != true ]]; then
  mode="$(stat -c '%a' "$ENV_FILE")"
  [[ "$mode" == "600" ]] || die "$ENV_FILE must have mode 600 (current: $mode)"
fi

command -v docker >/dev/null || die "docker is not installed"
compose config --quiet
log "preflight passed"

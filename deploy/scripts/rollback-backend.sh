#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"

failed_image="${1:?usage: rollback-backend.sh ghcr.io/owner/image@sha256:digest}"
[[ "$failed_image" =~ ^ghcr\.io/.+@sha256:[0-9a-f]{64}$ ]] || die "rollback requires the failed immutable GHCR digest"
load_env
mkdir -p "${DEPLOY_ROOT}/state"
exec 8>"${DEPLOY_ROOT}/state/deploy.lock"
flock -n 8 || die "another deployment is already running"

pending_file="${DEPLOY_ROOT}/state/pending-backend-deployment"
[[ -f "$pending_file" ]] || die "no pending backend deployment exists"
mapfile -t pending < "$pending_file"
[[ "${#pending[@]}" -eq 2 ]] || die "pending backend deployment state is invalid"
[[ "${pending[0]}" == "$failed_image" ]] || die "pending backend image does not match the failed digest"
[[ "$BACKEND_IMAGE_REF" == "$failed_image" ]] || die "production environment no longer points to the failed image"
previous="${pending[1]}"

update_image() {
  local value="$1" tmp
  tmp="$(mktemp "${ENV_FILE}.XXXXXX")"
  awk -v image="$value" 'BEGIN{done=0} /^BACKEND_IMAGE_REF=/{print "BACKEND_IMAGE_REF=" image; done=1; next} {print} END{if(!done) print "BACKEND_IMAGE_REF=" image}' "$ENV_FILE" > "$tmp"
  chmod 600 "$tmp"
  mv "$tmp" "$ENV_FILE"
  BACKEND_IMAGE_REF="$value"
  export BACKEND_IMAGE_REF
}
healthy() {
  local status="not-run"
  for _ in {1..30}; do
    status="$(curl --silent --output /dev/null --write-out '%{http_code}' \
      --connect-timeout 2 --max-time 5 \
      -u "${RUNTIME_HTTP_KEY}:" \
      -H "Content-Type: application/json" \
      -d '{}' \
      "http://127.0.0.1:7350/v2/rpc/healthcheck?unwrap" || true)"
    [[ "$status" == 200 ]] && return 0
    sleep 2
  done
  log "rollback runtime RPC health probe failed (last HTTP status: ${status:-transport-error})"
  return 1
}

if [[ "$previous" =~ ^ghcr\.io/.+@sha256:[0-9a-f]{64}$ ]]; then
  update_image "$previous"
  if compose pull nakama && compose up -d --remove-orphans db nakama cloudflared && healthy; then
    printf '%s\n' "$previous" > "${DEPLOY_ROOT}/state/last-known-good-image"
    rm -f -- "$pending_file"
    log "public smoke failed; rollback succeeded to $previous"
    exit 0
  fi
  compose logs --tail=200 nakama >&2 || true
  die "public smoke failed and rollback to the previous backend also failed"
fi

if compose down --remove-orphans; then
  update_image "$previous"
  rm -f -- "$pending_file" "${DEPLOY_ROOT}/state/last-known-good-image"
  log "initial public smoke failed; incomplete stack stopped and database volume preserved"
  exit 0
fi
die "initial public smoke failed and the incomplete stack could not be stopped; close the Tunnel and application manually"

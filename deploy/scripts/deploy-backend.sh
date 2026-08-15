#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"

new_image="${1:?usage: deploy-backend.sh ghcr.io/owner/image@sha256:digest}"
[[ "$new_image" =~ ^ghcr\.io/.+@sha256:[0-9a-f]{64}$ ]] || die "deployment requires an immutable GHCR digest"
load_env
"$SCRIPT_DIR/preflight.sh" --backend-image "$new_image"
mkdir -p "${DEPLOY_ROOT}/state"
exec 8>"${DEPLOY_ROOT}/state/deploy.lock"
flock -n 8 || die "another deployment is already running"

previous="$BACKEND_IMAGE_REF"
has_previous=false
bootstrap=false
if [[ "$previous" =~ ^ghcr\.io/.+@sha256:[0-9a-f]{64}$ ]]; then
  has_previous=true
fi
if docker volume inspect cs2match-postgres-data >/dev/null 2>&1; then
  log "existing database volume detected; starting database for pre-deploy backup"
  compose up -d db
  for _ in {1..30}; do compose exec -T db pg_isready -h 127.0.0.1 -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1 && break; sleep 1; done
  if ! compose exec -T db pg_isready -h 127.0.0.1 -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
    compose logs --tail=100 db >&2 || true
    die "database TCP service did not become ready for pre-deploy backup"
  fi
  "$SCRIPT_DIR/backup-db.sh" pre-deploy
else
  bootstrap=true
  log "no database volume exists; treating this as the first bootstrap deployment"
fi
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
  for _ in {1..30}; do
    curl --fail --silent --show-error -u "${NAKAMA_SERVER_KEY}:" -H "Content-Type: application/json" -d '{}' "http://127.0.0.1:7350/v2/rpc/HealthCheck?unwrap" >/dev/null 2>&1 && return 0
    sleep 2
  done
  return 1
}

update_image "$new_image"
if compose pull nakama && compose up -d --remove-orphans db nakama cloudflared && healthy; then
  if [[ "$bootstrap" == true ]]; then
    "$SCRIPT_DIR/backup-db.sh" initial || die "application is healthy, but the required initial offsite backup failed"
  fi
  printf '%s\n' "$new_image" > "${DEPLOY_ROOT}/state/last-known-good-image"
  log "backend deployment healthy: $new_image"
  exit 0
fi

log "new deployment failed; rolling back to previous image"
compose logs --tail=200 nakama >&2 || true
if [[ "$has_previous" != true ]]; then
  update_image "$previous"
  die "initial deployment failed and no previous healthy image exists; fix the reported error and rerun the workflow"
fi
update_image "$previous"
if compose pull nakama && compose up -d nakama cloudflared && healthy; then
  die "deployment failed; rollback succeeded to $previous"
fi
compose logs --tail=200 nakama >&2 || true
die "deployment and rollback both failed; inspect logs, then run: $SCRIPT_DIR/deploy-backend.sh <known-good-digest>"

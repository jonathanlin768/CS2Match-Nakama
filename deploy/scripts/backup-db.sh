#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"
load_env

tag="${1:-daily}"
[[ "$tag" =~ ^[a-zA-Z0-9._-]+$ ]] || die "invalid backup tag"
mkdir -p "${DEPLOY_ROOT}/state" "${DEPLOY_ROOT}/backups"
exec 9>"${DEPLOY_ROOT}/state/backup.lock"
flock -n 9 || die "another backup is already running"

dump="${DEPLOY_ROOT}/backups/postgres-${tag}-$(date -u +%Y%m%dT%H%M%SZ).dump"
cleanup() { rm -f -- "$dump"; }
trap cleanup EXIT

log "creating a consistent PostgreSQL dump"
compose exec -T db pg_dump -U "$DB_USER" -d "$DB_NAME" --format=custom --no-owner --no-acl > "$dump"
[[ -s "$dump" ]] || die "database dump is empty"

restic=(docker run --rm --env-file "$ENV_FILE" -v "${DEPLOY_ROOT}/backups:/backup:ro" "$RESTIC_IMAGE")
if ! "${restic[@]}" snapshots >/dev/null 2>&1; then
  log "initializing encrypted backup repository"
  "${restic[@]}" init >/dev/null
fi
log "uploading encrypted backup"
"${restic[@]}" backup "/backup/$(basename "$dump")" --tag "$tag" --tag postgres
"${restic[@]}" forget --tag postgres --keep-daily "${BACKUP_KEEP_DAILY:-7}" --keep-weekly "${BACKUP_KEEP_WEEKLY:-4}" --prune
log "backup completed: tag=$tag"

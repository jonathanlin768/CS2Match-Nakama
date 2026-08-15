#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"
load_env

snapshot="${1:-latest}"
work="$(mktemp -d "${TMPDIR:-/tmp}/cs2match-restore.XXXXXX")"
container="cs2match-restore-$RANDOM"
cleanup() { docker rm -f "$container" >/dev/null 2>&1 || true; rm -rf -- "$work"; }
trap cleanup EXIT

log "restoring snapshot into an isolated directory"
docker run --rm --env-file "$ENV_FILE" -v "$work:/restore" "$RESTIC_IMAGE" restore "$snapshot" --target /restore
dump="$(find "$work" -type f -name '*.dump' -print -quit)"
[[ -n "$dump" && -s "$dump" ]] || die "snapshot contains no PostgreSQL dump"

docker run -d --name "$container" -e POSTGRES_DB=verify -e POSTGRES_USER=verify -e POSTGRES_PASSWORD=verify "$POSTGRES_IMAGE" >/dev/null
for _ in {1..30}; do docker exec "$container" pg_isready -h 127.0.0.1 -U verify -d verify >/dev/null 2>&1 && break; sleep 1; done
if ! docker exec "$container" pg_isready -h 127.0.0.1 -U verify -d verify >/dev/null 2>&1; then
  docker logs "$container" >&2 || true
  die "temporary PostgreSQL TCP service did not become ready"
fi
docker cp "$dump" "$container:/tmp/restore.dump"
docker exec "$container" pg_restore -U verify -d verify --no-owner --no-acl /tmp/restore.dump
docker exec "$container" psql -U verify -d verify -v ON_ERROR_STOP=1 -Atc "SELECT count(*) FROM users; SELECT count(*) FROM storage;" >/dev/null
log "restore verification passed in isolated PostgreSQL"

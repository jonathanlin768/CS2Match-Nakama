#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

DEPLOY_ROOT="${DEPLOY_ROOT:-/opt/cs2match}"
ENV_FILE="${ENV_FILE:-${DEPLOY_ROOT}/.env.production}"
COMPOSE_FILE="${COMPOSE_FILE:-${DEPLOY_ROOT}/docker-compose.prod.yml}"

die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
log() { printf '[cs2match] %s\n' "$*"; }

load_env() {
  [[ -f "$ENV_FILE" ]] || die "missing environment file: $ENV_FILE"
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
}

compose() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
work="$(mktemp -d "${TMPDIR:-/tmp}/cs2match-script-test.XXXXXX")"
cleanup() { rm -rf -- "$work"; }
trap cleanup EXIT
mkdir -p "$work/bin" "$work/root/state" "$work/root/backups"

cat > "$work/bin/docker" <<'MOCK'
#!/usr/bin/env bash
set -eu
printf '%s\n' "$*" >> "${MOCK_CALLS:?}"
if [[ "${1:-}" == compose ]]; then
  printf 'compose-image=%s\n' "${BACKEND_IMAGE_REF:-}" >> "${MOCK_CALLS:?}"
fi
case "$*" in
  "volume inspect cs2match-postgres-data")
    [[ "${MOCK_VOLUME_EXISTS:-1}" == 1 ]]
    ;;
  "build "*)
    build_count="$(grep -c '^build ' "${MOCK_CALLS:?}" || true)"
    if [[ "$build_count" -lt "${MOCK_BUILD_SUCCEEDS_ON:-1}" ]]; then
      printf '%s\n' "${MOCK_BUILD_ERROR:-429 Too Many Requests}" >&2
      exit 42
    fi
    ;;
  *"compose"*"pg_dump"*)
    [[ "${MOCK_DUMP_FAIL:-0}" == 0 ]] || exit 40
    printf 'mock-pg-dump'
    ;;
  *" snapshots") exit "${MOCK_RESTIC_LIST_FAIL:-0}" ;;
  *" init") exit 0 ;;
  *" backup "*) exit "${MOCK_BACKUP_FAIL:-0}" ;;
  *" forget "*) touch "${MOCK_FORGET_MARKER:?}" ;;
  *" restore "*)
    [[ "${MOCK_RESTORE_FAIL:-0}" == 0 ]] || exit 41
    if [[ "${MOCK_CORRUPT:-0}" == 0 ]]; then mkdir -p "${MOCK_RESTORE_ROOT:?}/backup"; printf dump > "${MOCK_RESTORE_ROOT}/backup/test.dump"; fi
    ;;
  *) exit 0 ;;
esac
MOCK
chmod +x "$work/bin/docker"
cat > "$work/bin/curl" <<'MOCK'
#!/usr/bin/env bash
exit "${MOCK_CURL_STATUS:-0}"
MOCK
chmod +x "$work/bin/curl"
cat > "$work/bin/flock" <<'MOCK'
#!/usr/bin/env bash
[[ "${MOCK_LOCKED:-0}" == 0 ]]
MOCK
chmod +x "$work/bin/flock"

cat > "$work/prod.env" <<'ENV'
COMPOSE_PROJECT_NAME=cs2match-test
BACKEND_IMAGE_REF=ghcr.io/test/cs2match@sha256:0000000000000000000000000000000000000000000000000000000000000000
POSTGRES_IMAGE=postgres:15.18-alpine
CLOUDFLARED_IMAGE_REF=cloudflare/cloudflared:2026.7.2@sha256:1111111111111111111111111111111111111111111111111111111111111111
DB_NAME=nakama
DB_USER=nakama
DB_PASSWORD=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
NAKAMA_SERVER_KEY=unit-test-public-key
CONSOLE_USERNAME=operator
CONSOLE_PASSWORD=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
CONSOLE_SIGNING_KEY=cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
SESSION_ENCRYPTION_KEY=dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
SESSION_REFRESH_ENCRYPTION_KEY=eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
RUNTIME_HTTP_KEY=ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
API_HOST=api.ci.invalid
FRONTEND_ORIGIN=https://play.ci.invalid
TAILSCALE_IP=100.64.0.10
CLOUDFLARE_TUNNEL_TOKEN=unit-test-tunnel-token
RESTIC_IMAGE=restic/restic:0.18.0
RESTIC_REPOSITORY=s3:https://invalid.example/bucket
RESTIC_PASSWORD=unit-test-only
AWS_ACCESS_KEY_ID=unit-test
AWS_SECRET_ACCESS_KEY=unit-test
AWS_DEFAULT_REGION=auto
BACKUP_KEEP_DAILY=7
BACKUP_KEEP_WEEKLY=4
ENV
touch "$work/compose.yml" "$work/calls"

export PATH="$work/bin:$PATH" DEPLOY_ROOT="$work/root" ENV_FILE="$work/prod.env" COMPOSE_FILE="$work/compose.yml"
export MOCK_CALLS="$work/calls" MOCK_FORGET_MARKER="$work/forget-ran"

expect_failure() {
  set +e
  "$@" >/dev/null 2>&1
  status=$?
  set -e
  [[ $status -ne 0 ]] || { echo "expected command to fail: $*" >&2; exit 1; }
}

bash "$repo_root/deploy/scripts/preflight.sh" --skip-permissions >/dev/null
cp "$work/prod.env" "$work/bootstrap.env"
chmod 600 "$work/bootstrap.env"
sed -i 's#^BACKEND_IMAGE_REF=.*#BACKEND_IMAGE_REF=ghcr.io/owner/cs2match-nakama@sha256:REPLACE_ME#' "$work/bootstrap.env"
ENV_FILE="$work/bootstrap.env" bash "$repo_root/deploy/scripts/preflight.sh" --skip-permissions \
  --backend-image ghcr.io/test/cs2match@sha256:2222222222222222222222222222222222222222222222222222222222222222 >/dev/null
bootstrap_image=ghcr.io/test/cs2match@sha256:2222222222222222222222222222222222222222222222222222222222222222
: > "$work/calls"
MOCK_VOLUME_EXISTS=0 ENV_FILE="$work/bootstrap.env" bash "$repo_root/deploy/scripts/deploy-backend.sh" "$bootstrap_image" >/dev/null
grep -Fxq "BACKEND_IMAGE_REF=$bootstrap_image" "$work/bootstrap.env"
! grep -q 'OWNER/REPOSITORY\|REPLACE_ME' "$work/calls"
cp "$work/prod.env" "$work/default.env"
printf '\nNAKAMA_SERVER_KEY=defaultkey\n' >> "$work/default.env"
ENV_FILE="$work/default.env" expect_failure bash "$repo_root/deploy/scripts/preflight.sh" --skip-permissions
cp "$work/prod.env" "$work/duplicate.env"
printf '\nSESSION_REFRESH_ENCRYPTION_KEY=dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd\n' >> "$work/duplicate.env"
ENV_FILE="$work/duplicate.env" expect_failure bash "$repo_root/deploy/scripts/preflight.sh" --skip-permissions
if [[ "$(uname -s)" == Linux ]]; then
  chmod 644 "$work/prod.env"
  expect_failure bash "$repo_root/deploy/scripts/preflight.sh"
  chmod 600 "$work/prod.env"
fi

MOCK_DUMP_FAIL=1 expect_failure bash "$repo_root/deploy/scripts/backup-db.sh" test
[[ ! -e "$work/forget-ran" && -z "$(find "$work/root/backups" -type f -print -quit)" ]]

MOCK_BACKUP_FAIL=1 expect_failure bash "$repo_root/deploy/scripts/backup-db.sh" test
[[ ! -e "$work/forget-ran" && -z "$(find "$work/root/backups" -type f -print -quit)" ]]

MOCK_LOCKED=1 expect_failure bash "$repo_root/deploy/scripts/backup-db.sh" concurrent

MOCK_RESTORE_FAIL=1 expect_failure bash "$repo_root/deploy/scripts/restore-verify.sh" latest
MOCK_RESTORE_FAIL=0 MOCK_CORRUPT=1 expect_failure bash "$repo_root/deploy/scripts/restore-verify.sh" latest
! grep -q 'compose.*exec.*db' "$work/calls"

: > "$work/calls"
MOCK_BUILD_SUCCEEDS_ON=3 DOCKER_BUILD_ATTEMPTS=3 DOCKER_BUILD_RETRY_DELAY_SECONDS=0 \
  bash "$repo_root/deploy/scripts/docker-build-retry.sh" -t retry-test .
[[ "$(grep -c '^build ' "$work/calls")" == 3 ]]

: > "$work/calls"
MOCK_BUILD_SUCCEEDS_ON=3 DOCKER_BUILD_ATTEMPTS=2 DOCKER_BUILD_RETRY_DELAY_SECONDS=0 \
  expect_failure bash "$repo_root/deploy/scripts/docker-build-retry.sh" -t retry-test .
[[ "$(grep -c '^build ' "$work/calls")" == 2 ]]

: > "$work/calls"
MOCK_BUILD_SUCCEEDS_ON=3 MOCK_BUILD_ERROR='ERROR: failed to build: invalid tag "ghcr.io/test/Uppercase:sha": repository name must be lowercase' \
  DOCKER_BUILD_ATTEMPTS=3 DOCKER_BUILD_RETRY_DELAY_SECONDS=0 \
  expect_failure bash "$repo_root/deploy/scripts/docker-build-retry.sh" -t retry-test .
[[ "$(grep -c '^build ' "$work/calls")" == 1 ]]

: > "$work/calls"
MOCK_BUILD_SUCCEEDS_ON=3 MOCK_BUILD_ERROR='# example/server: server/main.go:10:2: undefined: missingSymbol' \
  DOCKER_BUILD_ATTEMPTS=3 DOCKER_BUILD_RETRY_DELAY_SECONDS=0 \
  expect_failure bash "$repo_root/deploy/scripts/docker-build-retry.sh" -t retry-test .
[[ "$(grep -c '^build ' "$work/calls")" == 1 ]]

GITHUB_REPOSITORY='jonathanlin768/CS2Match-Nakama'
image="ghcr.io/${GITHUB_REPOSITORY,,}"
[[ "$image" == 'ghcr.io/jonathanlin768/cs2match-nakama' ]]

echo "deployment script failure and retry tests passed"

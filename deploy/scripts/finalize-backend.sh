#!/usr/bin/env bash
set -Eeuo pipefail
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "$SCRIPT_DIR/common.sh"

expected_image="${1:?usage: finalize-backend.sh ghcr.io/owner/image@sha256:digest}"
[[ "$expected_image" =~ ^ghcr\.io/.+@sha256:[0-9a-f]{64}$ ]] || die "finalization requires an immutable GHCR digest"
load_env
mkdir -p "${DEPLOY_ROOT}/state"
exec 8>"${DEPLOY_ROOT}/state/deploy.lock"
flock -n 8 || die "another deployment is already running"

pending_file="${DEPLOY_ROOT}/state/pending-backend-deployment"
[[ -f "$pending_file" ]] || die "no pending backend deployment exists"
mapfile -t pending < "$pending_file"
[[ "${#pending[@]}" -eq 2 ]] || die "pending backend deployment state is invalid"
[[ "${pending[0]}" == "$expected_image" ]] || die "pending backend image does not match the requested digest"
[[ "$BACKEND_IMAGE_REF" == "$expected_image" ]] || die "production environment no longer points to the pending image"

state_tmp="$(mktemp "${DEPLOY_ROOT}/state/last-known-good-image.XXXXXX")"
printf '%s\n' "$expected_image" > "$state_tmp"
chmod 600 "$state_tmp"
mv "$state_tmp" "${DEPLOY_ROOT}/state/last-known-good-image"
rm -f -- "$pending_file"
log "public smoke passed; backend deployment finalized: $expected_image"

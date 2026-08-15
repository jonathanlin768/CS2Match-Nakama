#!/usr/bin/env bash
set -Eeuo pipefail

max_attempts="${DOCKER_BUILD_ATTEMPTS:-3}"
base_delay_seconds="${DOCKER_BUILD_RETRY_DELAY_SECONDS:-20}"

[[ "$max_attempts" =~ ^[1-9][0-9]*$ ]] || {
  echo "DOCKER_BUILD_ATTEMPTS must be a positive integer" >&2
  exit 2
}
[[ "$base_delay_seconds" =~ ^[0-9]+$ ]] || {
  echo "DOCKER_BUILD_RETRY_DELAY_SECONDS must be a non-negative integer" >&2
  exit 2
}
(( $# > 0 )) || {
  echo "usage: docker-build-retry.sh [docker build arguments...]" >&2
  exit 2
}

attempt=1
while true; do
  if docker build "$@"; then
    exit 0
  else
    status=$?
  fi

  if (( attempt >= max_attempts )); then
    echo "docker build failed after ${attempt} attempt(s)" >&2
    exit "$status"
  fi

  delay_seconds=$((base_delay_seconds * attempt))
  echo "::warning::docker build attempt ${attempt}/${max_attempts} failed; retrying in ${delay_seconds}s" >&2
  sleep "$delay_seconds"
  attempt=$((attempt + 1))
done

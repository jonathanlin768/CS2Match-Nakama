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

is_retryable_failure() {
  local output_file="$1"
  grep -Eqi \
    '429[[:space:]]+Too Many Requests|too many requests|TOOMANYREQUESTS|502[[:space:]]+Bad Gateway|503[[:space:]]+Service Unavailable|504[[:space:]]+Gateway Timeout|HTTP[^[:cntrl:]]*(502|503|504)|unexpected status[^[:cntrl:]]*(502|503|504)|status( code)?[^[:digit:]]*(502|503|504)|i/o timeout|TLS handshake timeout|Client\.Timeout exceeded|context deadline exceeded|request canceled[^[:cntrl:]]*waiting for connection|operation timed out|ETIMEDOUT|connection reset( by peer)?|ECONNRESET|temporary failure in name resolution|could not resolve host|server misbehaving' \
    "$output_file"
}

build_log="$(mktemp "${TMPDIR:-/tmp}/cs2match-docker-build.XXXXXX")"
cleanup() { rm -f -- "$build_log"; }
trap cleanup EXIT

attempt=1
while true; do
  : > "$build_log"
  set +e
  docker build "$@" 2>&1 | tee "$build_log"
  pipeline_status=("${PIPESTATUS[@]}")
  set -e
  status="${pipeline_status[0]}"

  if (( status == 0 )); then
    exit 0
  fi

  if ! is_retryable_failure "$build_log"; then
    echo "docker build failed with a non-retryable error; stopping after ${attempt} attempt(s)" >&2
    exit "$status"
  fi

  if (( attempt >= max_attempts )); then
    echo "docker build failed with a retryable network error after ${attempt} attempt(s)" >&2
    exit "$status"
  fi

  delay_seconds=$((base_delay_seconds * attempt))
  echo "::warning::docker build attempt ${attempt}/${max_attempts} failed; retrying in ${delay_seconds}s" >&2
  sleep "$delay_seconds"
  attempt=$((attempt + 1))
done

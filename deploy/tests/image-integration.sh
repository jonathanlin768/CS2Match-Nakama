#!/usr/bin/env bash
set -Eeuo pipefail
image="${1:?usage: image-integration.sh IMAGE}"
runtime_http_key=33333333333333333333333333333333
suffix="${GITHUB_RUN_ID:-local}-$RANDOM"
network="cs2match-test-$suffix"
db="cs2match-db-test-$suffix"
app="cs2match-app-test-$suffix"
cleanup() {
  docker rm -f "$app" "$db" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker network create "$network" >/dev/null
docker run -d --name "$db" --network "$network" --network-alias db -e POSTGRES_DB=nakama -e POSTGRES_USER=nakama -e POSTGRES_PASSWORD=integration postgres:15.18-alpine >/dev/null
for _ in {1..30}; do docker exec "$db" pg_isready -h 127.0.0.1 -U nakama -d nakama >/dev/null 2>&1 && break; sleep 1; done
if ! docker exec "$db" pg_isready -h 127.0.0.1 -U nakama -d nakama >/dev/null 2>&1; then
  docker logs "$db" >&2 || true
  exit 1
fi
test "$(docker exec -e PGPASSWORD=integration "$db" psql -h 127.0.0.1 -w -U nakama -d nakama -v ON_ERROR_STOP=1 -Atc 'SELECT 1')" = 1

docker run --rm --network "$network" --entrypoint /nakama/nakama "$image" migrate up --database.address 'nakama:integration@db:5432/nakama?sslmode=disable'
docker run -d --name "$app" --network "$network" -p 127.0.0.1::7350 "$image" --config /nakama/data/nakama-config.yml --database.address 'nakama:integration@db:5432/nakama?sslmode=disable' --socket.server_key integration-public-key --runtime.http_key "$runtime_http_key" --session.encryption_key 11111111111111111111111111111111 --session.refresh_encryption_key 22222222222222222222222222222222 --console.username integration --console.password integration-console-password --console.signing_key 44444444444444444444444444444444 >/dev/null

for _ in {1..30}; do
  logs="$(docker logs "$app" 2>&1 || true)"
  host_port="$(docker port "$app" 7350/tcp 2>/dev/null | awk -F: 'NR == 1 {print $NF}')"
  response=""
  if [[ "$host_port" =~ ^[0-9]+$ ]]; then
    response="$(curl --fail --silent --show-error \
      --connect-timeout 2 --max-time 5 \
      -u "${runtime_http_key}:" \
      -H 'Content-Type: application/json' \
      -d '{}' \
      "http://127.0.0.1:${host_port}/v2/rpc/healthcheck?unwrap" 2>/dev/null || true)"
  fi
  if grep -q 'HealthCheck RPC registered' <<<"$logs" \
    && grep -q 'SimuMatch RPC registered' <<<"$logs" \
    && grep -q 'Social RPCs registered' <<<"$logs" \
    && grep -q '"status":"ok"' <<<"$response"; then
    exit 0
  fi
  sleep 1
done
docker logs "$app" >&2
printf 'HealthCheck RPC did not return a successful response\n' >&2
exit 1

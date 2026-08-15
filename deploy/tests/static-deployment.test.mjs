import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import test from "node:test"

const compose = readFileSync(new URL("../../docker-compose.prod.yml", import.meta.url), "utf8")
const devCompose = readFileSync(new URL("../../docker-compose.yml", import.meta.url), "utf8")
const tunnel = readFileSync(new URL("../cloudflared/config.yml.example", import.meta.url), "utf8")
const env = readFileSync(new URL("../../.env.production.example", import.meta.url), "utf8")
const workflow = readFileSync(new URL("../../.github/workflows/deploy-production.yml", import.meta.url), "utf8")
const deploy = readFileSync(new URL("../scripts/deploy-backend.sh", import.meta.url), "utf8")
const backup = readFileSync(new URL("../scripts/backup-db.sh", import.meta.url), "utf8")
const dockerBuildRetry = readFileSync(new URL("../scripts/docker-build-retry.sh", import.meta.url), "utf8")
const backendDockerfile = readFileSync(new URL("../../server/Dockerfile.prod", import.meta.url), "utf8")
const nginx = readFileSync(new URL("../../client/nginx.conf", import.meta.url), "utf8")
const redirects = readFileSync(new URL("../../client/public/_redirects", import.meta.url), "utf8")

test("production database is internal and has no published port", () => {
  const db = compose.match(/  db:\n([\s\S]*?)\n  nakama:/)?.[1] ?? ""
  assert.match(db, /networks:\n\s+- backend/)
  assert.doesNotMatch(db, /ports:/)
  assert.match(compose, /backend:\n\s+internal: true/)
  assert.match(compose, /mem_limit:/)
  assert.match(compose, /cpus:/)
})

test("Nakama API binds to loopback and Console binds only to Tailscale", () => {
  assert.match(compose, /127\.0\.0\.1:7350:7350/)
  assert.match(compose, /TAILSCALE_IP[^}]*}:7351:7351/)
  assert.doesNotMatch(compose, /- ["']?(?:0\.0\.0\.0:)?5432:/)
})

test("Tunnel has one API ingress and a terminal 404", () => {
  assert.equal((tunnel.match(/hostname:/g) ?? []).length, 1)
  assert.match(tunnel, /service: http:\/\/nakama:7350/)
  assert.match(tunnel, /service: http_status:404\s*$/)
  assert.doesNotMatch(tunnel, /7351/)
})

test("production template contains no real credentials and pins immutable backend", () => {
  assert.match(env, /BACKEND_IMAGE_REF=ghcr\.io\/owner\/cs2match-nakama@sha256:REPLACE_ME/)
  assert.doesNotMatch(env, /defaultkey|localpass/)
})

test("development compose remains a separate local stack", () => {
  assert.match(devCompose, /7350:7350/)
  assert.doesNotMatch(devCompose, /cloudflared/)
})

test("both Nginx and Pages preserve SPA route fallback", () => {
  assert.match(nginx, /try_files \$uri \$uri\/ \/index\.html/)
  assert.match(redirects, /^\/\* \/index\.html 200\s*$/)
})

test("workflow actions are immutable and deployment uses Tailscale OIDC", () => {
  for (const use of workflow.matchAll(/uses:\s*([^\s#]+)/g)) assert.match(use[1], /@[0-9a-f]{40}$/)
  assert.match(workflow, /default: master/)
  assert.match(workflow, /origin\/master/)
  assert.doesNotMatch(workflow, /origin\/main|REQUESTED_REF" != main/)
  assert.match(workflow, /id-token: write/)
  assert.match(workflow, /audience:/)
  assert.match(workflow, /tailscale ssh/)
  assert.match(workflow, /packages: read/)
  assert.match(workflow, /docker login ghcr\.io/)
  assert.match(workflow, /docker logout ghcr\.io/)
  assert.doesNotMatch(workflow, /LIGHTSAIL_IP|SSH_PRIVATE_KEY/)
})

test("backend builds bypass the rate-limited gateway and retry transient failures", () => {
  assert.match(backendDockerfile, /FROM heroiclabs\/nakama-pluginbuilder:3\.30\.0/)
  assert.match(backendDockerfile, /FROM heroiclabs\/nakama:3\.30\.0/)
  assert.doesNotMatch(backendDockerfile, /registry\.heroiclabs\.com/)
  assert.equal((workflow.match(/bash deploy\/scripts\/docker-build-retry\.sh/g) ?? []).length, 2)
  assert.match(workflow, /image="ghcr\.io\/\$\{GITHUB_REPOSITORY,,\}"/)
  assert.doesNotMatch(workflow, /\$\{GITHUB_REPOSITORY#\*\/\}-nakama/)
  assert.match(dockerBuildRetry, /DOCKER_BUILD_ATTEMPTS:-3/)
  assert.match(dockerBuildRetry, /docker build "\$@"/)
  assert.match(dockerBuildRetry, /429\[\[:space:\]\]\+Too Many Requests/)
  assert.match(dockerBuildRetry, /connection reset/)
  assert.match(dockerBuildRetry, /temporary failure in name resolution/)
  assert.match(dockerBuildRetry, /502\[\[:space:\]\]\+Bad Gateway/)
  assert.match(dockerBuildRetry, /non-retryable error/)
})

test("deployment backs up, locks, health checks and rolls back", () => {
  assert.match(deploy, /flock -n/)
  assert.match(deploy, /preflight\.sh" --backend-image "\$new_image"/)
  assert.match(deploy, /docker volume inspect cs2match-postgres-data/)
  assert.match(deploy, /backup-db\.sh.*pre-deploy/)
  assert.match(deploy, /backup-db\.sh" initial/)
  assert.match(deploy, /previous=.*BACKEND_IMAGE_REF/)
  assert.match(deploy, /rolling back/)
  assert.match(deploy, /HealthCheck/)
})

test("backup is locked, encrypted offsite and retained", () => {
  assert.match(backup, /flock -n/)
  assert.match(backup, /RESTIC_IMAGE/)
  assert.match(backup, /forget.*keep-daily/)
  assert.doesNotMatch(backup, /echo.*(?:PASSWORD|SECRET|TOKEN)/i)
})

import assert from "node:assert/strict"
import test from "node:test"
import { validateProductionEnv } from "../../scripts/validate-production-env.mjs"

const valid = {
  VITE_NAKAMA_HOST: "api.ci.invalid",
  VITE_NAKAMA_PORT: "443",
  VITE_NAKAMA_SERVER_KEY: "public-app-key",
  VITE_NAKAMA_USE_SSL: "true",
}

test("accepts an explicit TLS production endpoint", () => {
  assert.doesNotThrow(() => validateProductionEnv(valid))
})

for (const host of ["localhost", "127.0.0.1", "nakama"]) {
  test(`rejects development host ${host}`, () => {
    assert.throws(() => validateProductionEnv({ ...valid, VITE_NAKAMA_HOST: host }))
  })
}

test("rejects documentation placeholders", () => {
  assert.throws(() => validateProductionEnv({ ...valid, VITE_NAKAMA_HOST: "api.example.com" }))
  assert.throws(() => validateProductionEnv({ ...valid, VITE_NAKAMA_SERVER_KEY: "REPLACE_ME" }))
})

test("rejects missing or insecure production values", () => {
  assert.throws(() => validateProductionEnv({ ...valid, VITE_NAKAMA_HOST: "" }))
  assert.throws(() => validateProductionEnv({ ...valid, VITE_NAKAMA_PORT: "7350" }))
  assert.throws(() => validateProductionEnv({ ...valid, VITE_NAKAMA_USE_SSL: "false" }))
  assert.throws(() => validateProductionEnv({ ...valid, VITE_NAKAMA_SERVER_KEY: "defaultkey" }))
})

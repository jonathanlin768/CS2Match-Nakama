import { randomUUID } from "node:crypto"
import { performance } from "node:perf_hooks"
import { Client } from "@heroiclabs/nakama-js"
import { validateProductionEnv } from "./validate-production-env.mjs"

validateProductionEnv(process.env)
const concurrency = Number(process.argv[2] ?? 20)
if (![20, 50].includes(concurrency)) throw new Error("usage: production-load.mjs 20|50")
const client = new Client(process.env.VITE_NAKAMA_SERVER_KEY, process.env.VITE_NAKAMA_HOST, process.env.VITE_NAKAMA_PORT, true)

const results = await Promise.all(Array.from({ length: concurrency }, async () => {
  const started = performance.now()
  try {
    const session = await client.authenticateDevice(randomUUID(), true)
    const response = await client.rpc(session, "SimuMatch", { mode: "computer" })
    const text = typeof response.payload === "string" ? response.payload : JSON.stringify(response.payload)
    return { ok: true, ms: performance.now() - started, bytes: Buffer.byteLength(text) }
  } catch (error) {
    return { ok: false, ms: performance.now() - started, bytes: 0, error: String(error) }
  }
}))
const sorted = results.map((item) => item.ms).sort((a, b) => a - b)
const percentile = (p) => sorted[Math.min(sorted.length - 1, Math.ceil(sorted.length * p) - 1)]
const ok = results.filter((item) => item.ok)
const report = { concurrency, successRate: ok.length / concurrency, p50Ms: percentile(0.5), p95Ms: percentile(0.95), averageResponseBytes: ok.length ? Math.round(ok.reduce((sum, item) => sum + item.bytes, 0) / ok.length) : 0, failures: results.filter((item) => !item.ok).map((item) => item.error) }
console.log(JSON.stringify(report, null, 2))
if (ok.length !== concurrency) process.exitCode = 1

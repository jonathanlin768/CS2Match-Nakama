import { randomUUID } from "node:crypto"
import { Client } from "@heroiclabs/nakama-js"
import { validateProductionEnv } from "./validate-production-env.mjs"

validateProductionEnv(process.env)
const client = new Client(process.env.VITE_NAKAMA_SERVER_KEY, process.env.VITE_NAKAMA_HOST, process.env.VITE_NAKAMA_PORT, true)

async function describeFailure(error) {
  if (error && typeof error === "object" && typeof error.status === "number" && typeof error.text === "function") {
    let detail = ""
    try {
      const body = (await error.text()).trim()
      if (body) detail = `: ${body.slice(0, 1000)}`
    } catch {
      // Preserve the HTTP status even when the response body cannot be read.
    }
    return `HTTP ${error.status}${error.statusText ? ` ${error.statusText}` : ""}${detail}`
  }
  return error instanceof Error ? error.message : String(error)
}

async function stage(name, operation) {
  try {
    const result = await operation()
    console.log(`[smoke] ${name}: passed`)
    return result
  } catch (error) {
    throw new Error(`[smoke] ${name}: ${await describeFailure(error)}`)
  }
}

try {
  // Do not pass a custom username: the production identity hook generates the
  // required eight-character player code for newly created accounts.
  const session = await stage("device authentication", () => client.authenticateDevice(randomUUID(), true))

  const health = await stage("HealthCheck RPC", () => client.rpc(session, "HealthCheck", {}))
  const healthPayload = typeof health.payload === "string" ? JSON.parse(health.payload) : health.payload
  if (healthPayload?.status !== "ok") throw new Error("[smoke] HealthCheck RPC: unexpected response")

  const match = await stage("SimuMatch RPC", () => client.rpc(session, "SimuMatch", { mode: "computer" }))
  if (!match.payload) throw new Error("[smoke] SimuMatch RPC: response has no payload")

  const socket = client.createSocket(false, false)
  await stage("WebSocket", () => socket.connect(session, true))
  socket.disconnect(true)
  console.log("Production authentication, RPC and WebSocket smoke passed")
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error))
  process.exitCode = 1
}

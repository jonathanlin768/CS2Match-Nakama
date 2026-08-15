import { randomUUID } from "node:crypto"
import { Client } from "@heroiclabs/nakama-js"
import { validateProductionEnv } from "./validate-production-env.mjs"

validateProductionEnv(process.env)
const client = new Client(process.env.VITE_NAKAMA_SERVER_KEY, process.env.VITE_NAKAMA_HOST, process.env.VITE_NAKAMA_PORT, true)
const session = await client.authenticateDevice(randomUUID(), true, `smoke-${Date.now()}`)

const health = await client.rpc(session, "HealthCheck", {})
const healthPayload = typeof health.payload === "string" ? JSON.parse(health.payload) : health.payload
if (healthPayload?.status !== "ok") throw new Error("HealthCheck returned an unexpected response")

const match = await client.rpc(session, "SimuMatch", { mode: "computer" })
if (!match.payload) throw new Error("SimuMatch returned no payload")

const socket = client.createSocket(false, false)
await socket.connect(session, true)
socket.disconnect(true)
console.log("Production authentication, RPC and WebSocket smoke passed")

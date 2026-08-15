const required = [
  "VITE_NAKAMA_HOST",
  "VITE_NAKAMA_PORT",
  "VITE_NAKAMA_SERVER_KEY",
  "VITE_NAKAMA_USE_SSL",
]

export function validateProductionEnv(env = process.env) {
  const missing = required.filter((name) => !env[name]?.trim())
  if (missing.length > 0) {
    throw new Error(`Missing production frontend variables: ${missing.join(", ")}`)
  }
  if (["localhost", "127.0.0.1", "nakama"].includes(env.VITE_NAKAMA_HOST)) {
    throw new Error("VITE_NAKAMA_HOST must be a public production hostname")
  }
  if (env.VITE_NAKAMA_HOST.endsWith(".example.com")) {
    throw new Error("VITE_NAKAMA_HOST still uses the documentation placeholder domain")
  }
  if (env.VITE_NAKAMA_PORT !== "443") {
    throw new Error("VITE_NAKAMA_PORT must be 443 in production")
  }
  if (env.VITE_NAKAMA_USE_SSL !== "true") {
    throw new Error("VITE_NAKAMA_USE_SSL must be true in production")
  }
  if (env.VITE_NAKAMA_SERVER_KEY === "defaultkey") {
    throw new Error("VITE_NAKAMA_SERVER_KEY must not use the Nakama default")
  }
  if (/replace|change[_-]?me/i.test(env.VITE_NAKAMA_SERVER_KEY)) {
    throw new Error("VITE_NAKAMA_SERVER_KEY still contains a placeholder")
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  validateProductionEnv()
  console.log("Production frontend environment is valid")
}
import { pathToFileURL } from "node:url"

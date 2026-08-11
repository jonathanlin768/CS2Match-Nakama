import { useEffect, useState } from "react"
import { loadConfig, type Tables } from "../config"

let cached: Tables | null = null
let pending: Promise<Tables> | null = null

function getConfig() {
  if (cached) return Promise.resolve(cached)
  pending ??= loadConfig().then((tables) => { cached = tables; return tables })
  return pending
}

export function useGameConfig() {
  const [tables, setTables] = useState<Tables | null>(cached)
  const [error, setError] = useState<string | null>(null)
  useEffect(() => {
    let active = true
    getConfig().then((value) => { if (active) setTables(value) }).catch((reason: unknown) => {
      if (active) setError(reason instanceof Error ? reason.message : String(reason))
    })
    return () => { active = false }
  }, [])
  return { tables, error, loading: !tables && !error }
}

import { useCallback, useEffect, useState } from "react"
import { loadConfig, type Tables } from "../config"

let cached: Tables | null = null
let pending: Promise<Tables> | null = null

function getConfig(force = false) {
  if (force) {
    cached = null
    pending = null
  }
  if (cached) return Promise.resolve(cached)
  pending ??= loadConfig()
    .then((tables) => { cached = tables; return tables })
    .catch((reason) => { pending = null; throw reason })
  return pending
}

export function useGameConfig() {
  const [tables, setTables] = useState<Tables | null>(cached)
  const [error, setError] = useState<string | null>(null)
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    let active = true
    getConfig(attempt > 0).then((value) => { if (active) setTables(value) }).catch((reason: unknown) => {
      if (active) setError(reason instanceof Error ? reason.message : String(reason))
    })
    return () => { active = false }
  }, [attempt])
  const retry = useCallback(() => {
    setTables(null)
    setError(null)
    setAttempt((value) => value + 1)
  }, [])
  return { tables, error, loading: !tables && !error, retry }
}

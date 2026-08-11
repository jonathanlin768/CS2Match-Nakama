import { useCallback, useEffect, useState } from "react"
import type { Session } from "@heroiclabs/nakama-js"
import { linkEmailAccount, loginWithEmail, logoutToFreshGuest, restoreOrCreateSession, type IdentityKind } from "../api/auth"

type AuthStatus = "restoring" | "authenticated" | "guest"

interface AuthState {
  status: AuthStatus
  session: Session | null
  kind: IdentityKind
  error: string | null
}

export function useNakamaAuth() {
  const [state, setState] = useState<AuthState>({ status: "restoring", session: null, kind: "guest", error: null })

  useEffect(() => {
    let cancelled = false
    restoreOrCreateSession()
      .then(({ session, kind }) => {
        if (!cancelled) setState({ status: kind === "account" ? "authenticated" : "guest", session, kind, error: null })
      })
      .catch((error: unknown) => {
        if (!cancelled) setState({ status: "guest", session: null, kind: "guest", error: error instanceof Error ? error.message : String(error) })
      })
    return () => { cancelled = true }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    try {
      const session = await loginWithEmail(email, password)
      setState({ status: "authenticated", session, kind: "account", error: null })
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      setState((current) => ({ ...current, error: message }))
      throw error
    }
  }, [])

  const register = useCallback(async (email: string, password: string) => {
    if (!state.session) throw new Error("访客会话尚未准备好")
    try {
      const session = await linkEmailAccount(state.session, email, password)
      setState({ status: "authenticated", session, kind: "account", error: null })
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      setState((current) => ({ ...current, error: message }))
      throw error
    }
  }, [state.session])

  const logout = useCallback(async () => {
    const session = await logoutToFreshGuest()
    setState({ status: "guest", session, kind: "guest", error: null })
  }, [])

  return { ...state, isGuest: state.kind === "guest", login, register, logout }
}

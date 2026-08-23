import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react"
import { listContactExchangeInbox, toSocialApiError, type ContactExchangeInbox } from "../api/social"
import type { SocialApiError } from "../api/social-error"
import { useAuth } from "./AuthContext"
import { subscribeChannelMessages, subscribeNotifications, useSocket } from "../hooks/useSocket"
import { emptyContactInbox, mergeContactInboxPages } from "../social/contact-inbox"
import { shouldRefreshContactInbox } from "../social/contact-card"

interface ContactExchangeContextValue {
  inbox: ContactExchangeInbox
  status: "idle" | "loading" | "success" | "error"
  error: SocialApiError | null
  revision: number
  refresh: () => Promise<void>
}

const ContactExchangeContext = createContext<ContactExchangeContextValue | null>(null)

export function ContactExchangeProvider({ children }: { children: ReactNode }) {
  const { session, isGuest } = useAuth()
  const socket = useSocket(session)
  const [inbox, setInbox] = useState<ContactExchangeInbox>(emptyContactInbox)
  const [status, setStatus] = useState<ContactExchangeContextValue["status"]>("idle")
  const [error, setError] = useState<SocialApiError | null>(null)
  const [revision, setRevision] = useState(0)
  const requestID = useRef(0)

  const refresh = useCallback(async () => {
    if (!session || isGuest) {
      setInbox(emptyContactInbox)
      setStatus("idle")
      setError(null)
      return
    }
    const currentRequest = ++requestID.current
    setStatus((current) => current === "success" ? current : "loading")
    try {
      const pages: ContactExchangeInbox[] = []
      let cursor: string | undefined
      for (let page = 0; page < 10; page++) {
        const result = await listContactExchangeInbox(session, 100, cursor)
        pages.push(result)
        cursor = result.cursor
        if (!cursor) break
      }
      if (currentRequest !== requestID.current) return
      setInbox(mergeContactInboxPages(pages))
      setError(null)
      setStatus("success")
      setRevision((value) => value + 1)
    } catch (cause) {
      if (currentRequest !== requestID.current) return
      setError(await toSocialApiError(cause))
      setStatus("error")
    }
  }, [isGuest, session])

  useEffect(() => {
    const timer = window.setTimeout(() => { if (shouldRefreshContactInbox("login")) void refresh() }, 0)
    return () => window.clearTimeout(timer)
  }, [refresh])
  useEffect(() => {
    if (!session || isGuest) return
    const onVisible = () => { if (document.visibilityState === "visible" && shouldRefreshContactInbox("visible")) void refresh() }
    const timer = window.setInterval(() => void refresh(), 15_000)
    window.addEventListener("focus", onVisible)
    document.addEventListener("visibilitychange", onVisible)
    return () => {
      window.clearInterval(timer)
      window.removeEventListener("focus", onVisible)
      document.removeEventListener("visibilitychange", onVisible)
    }
  }, [isGuest, refresh, session])
  useEffect(() => {
    if (!socket) return
    const timer = window.setTimeout(() => { if (shouldRefreshContactInbox("reconnect")) void refresh() }, 0)
    const unsubscribeMessages = subscribeChannelMessages((message) => {
      if (shouldRefreshContactInbox("card", message)) void refresh()
    })
    const unsubscribeNotifications = subscribeNotifications(() => { if (shouldRefreshContactInbox("notification")) void refresh() })
    return () => { window.clearTimeout(timer); unsubscribeMessages(); unsubscribeNotifications() }
  }, [refresh, socket])

  const value = useMemo(() => ({ inbox, status, error, revision, refresh }), [error, inbox, refresh, revision, status])
  return <ContactExchangeContext.Provider value={value}>{children}</ContactExchangeContext.Provider>
}

export function useContactExchangeInbox() {
  const value = useContext(ContactExchangeContext)
  if (!value) throw new Error("useContactExchangeInbox must be used within <ContactExchangeProvider>")
  return value
}

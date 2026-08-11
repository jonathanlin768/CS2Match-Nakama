import { Session } from "@heroiclabs/nakama-js"
import client from "../nakama"

const TOKEN_KEY = "nakama_token"
const REFRESH_KEY = "nakama_refresh"
const DEVICE_KEY = "nakama_device_id"
const IDENTITY_KEY = "nakama_identity_kind"

export type IdentityKind = "guest" | "account"

function persistSession(session: Session, kind: IdentityKind) {
  localStorage.setItem(TOKEN_KEY, session.token)
  localStorage.setItem(REFRESH_KEY, session.refresh_token)
  localStorage.setItem(IDENTITY_KEY, kind)
}

function deviceID() {
  let value = localStorage.getItem(DEVICE_KEY)
  if (!value) {
    value = crypto.randomUUID()
    localStorage.setItem(DEVICE_KEY, value)
  }
  return value
}

export async function createGuestSession(freshDevice = false): Promise<Session> {
  if (freshDevice) localStorage.setItem(DEVICE_KEY, crypto.randomUUID())
  let lastError: unknown
  for (let attempt = 0; attempt < 3; attempt += 1) {
    try {
      const session = await client.authenticateDevice(deviceID(), true)
      persistSession(session, "guest")
      return session
    } catch (error) { lastError = error }
  }
  throw lastError instanceof Error ? lastError : new Error("无法创建游客账号")
}

export async function loginWithEmail(email: string, password: string): Promise<Session> {
  const session = await client.authenticateEmail(email.trim(), password, false)
  persistSession(session, "account")
  return session
}

// Link formal credentials to the current guest so its user id and history stay intact.
export async function linkEmailAccount(session: Session, email: string, password: string): Promise<Session> {
  await client.linkEmail(session, { email: email.trim(), password })
  persistSession(session, "account")
  return session
}

export async function restoreOrCreateSession(): Promise<{ session: Session; kind: IdentityKind }> {
  const token = localStorage.getItem(TOKEN_KEY)
  const refreshToken = localStorage.getItem(REFRESH_KEY)
  const kind = localStorage.getItem(IDENTITY_KEY) === "account" ? "account" : "guest"
  if (token && refreshToken) {
    try {
      let session = Session.restore(token, refreshToken)
      if (session.isexpired(Date.now())) session = await client.sessionRefresh(session)
      persistSession(session, kind)
      return { session, kind }
    } catch {
      clearSession()
    }
  }
  return { session: await createGuestSession(), kind: "guest" }
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_KEY)
  localStorage.removeItem(IDENTITY_KEY)
}

export async function logoutToFreshGuest(): Promise<Session> {
  clearSession()
  return createGuestSession(true)
}

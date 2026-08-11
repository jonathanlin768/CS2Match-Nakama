import type { Session } from "@heroiclabs/nakama-js"
import client from "../nakama"

export interface ContactProfile { qq?: string; wechat?: string; updated_at?: number }
export interface ContactProfileSummary { qq_configured: boolean; wechat_configured: boolean; qq_masked?: string; wechat_masked?: string }
export interface ContactExchange {
  request_id?: string
  requester_id?: string
  recipient_id?: string
  channels?: Array<"qq" | "wechat">
  status: "none" | "pending" | "accepted" | "declined" | "expired"
  version?: number
  requested_at?: number
  my_contact?: ContactProfile
  friend_contact?: ContactProfile
}

async function rpc<T>(session: Session, id: string, input: object): Promise<T> {
  const response = await client.rpc(session, id, input)
  const payload = typeof response.payload === "string" ? response.payload : JSON.stringify(response.payload)
  return JSON.parse(payload) as T
}

export const setContactProfile = (session: Session, profile: ContactProfile) => rpc<ContactProfileSummary>(session, "SocialSetContactProfile", profile)
export const getContactExchange = (session: Session, friendId: string) => rpc<ContactExchange>(session, "SocialGetContactExchange", { friend_id: friendId })
export const requestContactExchange = (session: Session, friendId: string, channels: Array<"qq" | "wechat">) => rpc<ContactExchange>(session, "SocialRequestContactExchange", { friend_id: friendId, channels })
export const respondContactExchange = (session: Session, friendId: string, requestId: string, accept: boolean) => rpc<ContactExchange>(session, "SocialRespondContactExchange", { friend_id: friendId, request_id: requestId, accept })

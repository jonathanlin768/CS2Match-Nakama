import type { Session } from "@heroiclabs/nakama-js"
import client from "../nakama"
import { decodeSocialRpcPayload, toSocialApiError } from "./social-error"
export { decodeSocialRpcPayload, SocialApiError, socialErrorMessage, toSocialApiError } from "./social-error"
export type { SocialApiError as SocialApiErrorType } from "./social-error"

export type ContactChannel = "qq" | "wechat"
export type ContactExchangeStatus = "none" | "pending" | "accepted" | "declined" | "expired" | "cancelled" | "stale" | "revoked"

export interface ContactProfile {
  qq?: string
  wechat?: string
  updated_at?: number
  revision?: number
}

export interface ContactProfileSummary {
  qq_configured: boolean
  wechat_configured: boolean
  qq_masked?: string
  wechat_masked?: string
}

export interface ContactProfileView extends ContactProfile {
  summary: ContactProfileSummary
}

export interface ContactExchange {
  request_id?: string
  requester_id?: string
  recipient_id?: string
  channels?: ContactChannel[]
  status: ContactExchangeStatus
  version?: number
  requested_at?: number
  responded_at?: number
  expires_at?: number
  requester_profile_revision?: number
  accepted_requester_profile_revision?: number
  accepted_recipient_profile_revision?: number
  my_contact?: ContactProfile
  friend_contact?: ContactProfile
}

export interface ContactInboxItem {
  friend_id: string
  username: string
  request_id?: string
  requester_id?: string
  recipient_id?: string
  channels?: ContactChannel[]
  status: ContactExchangeStatus
  version?: number
  requested_at?: number
  expires_at?: number
}

export interface ContactExchangeInbox {
  received: ContactInboxItem[]
  sent: ContactInboxItem[]
  reapproval: ContactInboxItem[]
  incoming_pending_count: number
  cursor?: string
}

async function rpc<T>(session: Session, id: string, input: object): Promise<T> {
  try {
    const response = await client.rpc(session, id, input)
    return decodeSocialRpcPayload<T>(response.payload)
  } catch (error) {
    throw await toSocialApiError(error)
  }
}

export const getContactProfile = (session: Session) => rpc<ContactProfileView>(session, "SocialGetContactProfile", {})
export const setContactProfile = (session: Session, profile: ContactProfile) => rpc<ContactProfileView>(session, "SocialSetContactProfile", { qq: profile.qq ?? "", wechat: profile.wechat ?? "" })
export const getContactExchange = (session: Session, friendId: string) => rpc<ContactExchange>(session, "SocialGetContactExchange", { friend_id: friendId })
export const requestContactExchange = (session: Session, friendId: string, channels: ContactChannel[]) => rpc<ContactExchange>(session, "SocialRequestContactExchange", { friend_id: friendId, channels })
export const respondContactExchange = (session: Session, friendId: string, requestId: string, accept: boolean) => rpc<ContactExchange>(session, "SocialRespondContactExchange", { friend_id: friendId, request_id: requestId, accept })
export const listContactExchangeInbox = (session: Session, limit = 100, cursor?: string) => rpc<ContactExchangeInbox>(session, "SocialListContactExchangeInbox", { limit, ...(cursor ? { cursor } : {}) })

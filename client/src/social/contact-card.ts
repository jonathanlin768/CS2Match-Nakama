export interface ContactCardSource {
  message_id?: string
  content?: unknown
  create_time?: string
  sender_id?: string
}

export interface ContactCardMessage {
  messageId: string
  requestId: string
  action: "requested" | "accepted" | "declined" | "stale" | "revoked"
  version: number
  createTime?: string
  senderId?: string
}

export function parseContactCard(message: ContactCardSource): ContactCardMessage | null {
  try {
    const content = typeof message.content === "string" ? JSON.parse(message.content) as Record<string, unknown> : message.content as Record<string, unknown> | null | undefined
    if (!content || content.type !== "contact_exchange" || typeof content.request_id !== "string" || !["requested", "accepted", "declined", "stale", "revoked"].includes(String(content.action))) return null
    return { messageId: message.message_id ?? `${content.request_id}:${content.version}`, requestId: content.request_id, action: content.action as ContactCardMessage["action"], version: Number(content.version ?? 0), createTime: message.create_time, senderId: message.sender_id }
  } catch { return null }
}

export function countUnreadCards(cards: ContactCardMessage[], userId: string, readAt: string) {
  return cards.filter((card) => card.senderId !== userId && (!readAt || (card.createTime ?? "") > readAt)).length
}

export type ContactInboxRefreshReason = "login" | "explicit" | "visible" | "reconnect" | "notification" | "card" | "card_history_failure"

export function shouldRefreshContactInbox(reason: ContactInboxRefreshReason, message?: ContactCardSource) {
  if (reason === "card") return Boolean(message && parseContactCard(message))
  if (reason === "card_history_failure") return false
  return true
}

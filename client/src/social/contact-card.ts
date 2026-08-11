export interface ContactCardSource {
  message_id?: string
  content?: unknown
  create_time?: string
  sender_id?: string
}

export interface ContactCardMessage {
  messageId: string
  requestId: string
  action: "requested" | "accepted" | "declined"
  version: number
  createTime?: string
  senderId?: string
}

export function parseContactCard(message: ContactCardSource): ContactCardMessage | null {
  try {
    const content = typeof message.content === "string" ? JSON.parse(message.content) as Record<string, unknown> : message.content as Record<string, unknown> | null | undefined
    if (!content || content.type !== "contact_exchange" || typeof content.request_id !== "string" || !["requested", "accepted", "declined"].includes(String(content.action))) return null
    return { messageId: message.message_id ?? `${content.request_id}:${content.version}`, requestId: content.request_id, action: content.action as ContactCardMessage["action"], version: Number(content.version ?? 0), createTime: message.create_time, senderId: message.sender_id }
  } catch { return null }
}

export function countUnreadCards(cards: ContactCardMessage[], userId: string, readAt: string) {
  return cards.filter((card) => card.senderId !== userId && (!readAt || (card.createTime ?? "") > readAt)).length
}

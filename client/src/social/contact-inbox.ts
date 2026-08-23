import type { ContactExchange, ContactExchangeInbox, ContactInboxItem, ContactProfile } from "../api/social"

export const emptyContactInbox: ContactExchangeInbox = {
  received: [],
  sent: [],
  reapproval: [],
  incoming_pending_count: 0,
}

export function mergeContactInboxPages(pages: ContactExchangeInbox[]): ContactExchangeInbox {
  return {
    received: pages.flatMap((page) => page.received ?? []),
    sent: pages.flatMap((page) => page.sent ?? []),
    reapproval: pages.flatMap((page) => page.reapproval ?? []),
    incoming_pending_count: pages.reduce((sum, page) => sum + (page.incoming_pending_count ?? 0), 0),
  }
}

export function contactProfileChanged(profile: ContactProfile | null, qq: string, wechat: string) {
  if (!profile) return false
  return (profile.qq ?? "").trim() !== qq.trim() || (profile.wechat ?? "").trim() !== wechat.trim()
}

export function contactChannels(profile: ContactProfile): Array<"qq" | "wechat"> {
  return [profile.qq?.trim() ? "qq" as const : null, profile.wechat?.trim() ? "wechat" as const : null]
    .filter((channel): channel is "qq" | "wechat" => channel !== null)
}

export function formatContactTime(epochSeconds?: number) {
  if (!epochSeconds) return "时间未知"
  return new Intl.DateTimeFormat("zh-CN", {
    month: "numeric",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(epochSeconds * 1000))
}

export function contactItemExpired(item: ContactInboxItem, nowSeconds = Date.now() / 1000) {
  return item.status === "expired" || Boolean(item.expires_at && item.expires_at <= nowSeconds)
}

export function disclosedFriendContact(exchange: ContactExchange) {
  if (exchange.status !== "accepted" || !exchange.friend_contact) return null
  return { qq: exchange.friend_contact.qq ?? "", wechat: exchange.friend_contact.wechat ?? "" }
}

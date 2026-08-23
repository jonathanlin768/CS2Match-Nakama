import { useEffect, useMemo, useRef, useState } from "react"
import type { ChannelMessage, Session } from "@heroiclabs/nakama-js"
import { joinDMChannel, leaveDMChannel, listChannelMessages } from "../api/chat"
import { countUnreadCards, parseContactCard, type ContactCardMessage } from "../social/contact-card"
import { subscribeChannelMessages, useSocket } from "./useSocket"
import { useContactExchangeInbox } from "../context/ContactExchangeContext"

export interface ContactInboxSummary {
  latest?: ContactCardMessage
  unread: number
}

function readKey(userId: string, friendId: string) {
  return `cs2match:contact-read:${userId}:${friendId}`
}

function markRead(userId: string, friendId: string) {
  localStorage.setItem(readKey(userId, friendId), new Date().toISOString())
}

export function useContactInbox(session: Session | null, friendIds: string[], selectedFriendId?: string) {
  const socket = useSocket(session)
  const { refresh: refreshAuthority } = useContactExchangeInbox()
  const idsKey = [...friendIds].sort().join(",")
  const normalizedIds = useMemo(() => idsKey ? idsKey.split(",") : [], [idsKey])
  const [cardsByFriend, setCardsByFriend] = useState<Record<string, ContactCardMessage[]>>({})
  const cardsRef = useRef<Record<string, ContactCardMessage[]>>({})
  const [summaries, setSummaries] = useState<Record<string, ContactInboxSummary>>({})
  const [revision, setRevision] = useState(0)
  const [cardError, setCardError] = useState<string | null>(null)
  const [retryRevision, setRetryRevision] = useState(0)

  useEffect(() => {
    const userId = session?.user_id
    if (!userId || !selectedFriendId) return
    const timer = window.setTimeout(() => {
      markRead(userId, selectedFriendId)
      setSummaries((current) => current[selectedFriendId]
        ? { ...current, [selectedFriendId]: { ...current[selectedFriendId], unread: 0 } }
        : current)
    }, 0)
    return () => window.clearTimeout(timer)
  }, [session, selectedFriendId])

  useEffect(() => {
    const userId = session?.user_id
    if (!socket || !session || !userId || normalizedIds.length === 0) return
    let active = true
    const channels = new Map<string, string>()

    const storeCards = (friendId: string, cards: ContactCardMessage[]) => {
      if (!active) return
      const ordered = [...cards].sort((a, b) => (a.createTime ?? "").localeCompare(b.createTime ?? ""))
      if (friendId === selectedFriendId) markRead(userId, friendId)
      const readAt = localStorage.getItem(readKey(userId, friendId)) ?? ""
      cardsRef.current = { ...cardsRef.current, [friendId]: ordered }
      setCardsByFriend(cardsRef.current)
      setSummaries((current) => ({ ...current, [friendId]: {
        latest: ordered.at(-1),
        unread: friendId === selectedFriendId ? 0 : countUnreadCards(ordered, userId, readAt),
      } }))
      setRevision((current) => current + 1)
    }

    const receive = (message: ChannelMessage) => {
      if (!message.channel_id) return
      const friendId = channels.get(message.channel_id)
      const card = parseContactCard(message)
      if (!friendId || !card) return
      const current = cardsRef.current[friendId] ?? []
      storeCards(friendId, [...current.filter((item) => item.messageId !== card.messageId), card])
      void refreshAuthority()
    }
    const unsubscribe = subscribeChannelMessages(receive)
    const joined: string[] = []

    void Promise.allSettled(normalizedIds.map(async (friendId) => {
      const channel = await joinDMChannel(socket, friendId)
      if (!active) { await leaveDMChannel(socket, channel.id); return }
      channels.set(channel.id, friendId)
      joined.push(channel.id)
      const history = await listChannelMessages(session, channel.id, 50)
      storeCards(friendId, (history.messages ?? []).map(parseContactCard).filter((item): item is ContactCardMessage => Boolean(item)))
    })).then((results) => {
      if (!active) return
      setCardError(results.some((result) => result.status === "rejected") ? "部分交换事件历史加载失败，可重试；权威申请状态不受影响。" : null)
    })

    return () => {
      active = false
      unsubscribe()
      joined.forEach((channelId) => { void leaveDMChannel(socket, channelId) })
    }
  }, [socket, session, normalizedIds, selectedFriendId, retryRevision, refreshAuthority])

  return { cardsByFriend, summaries, revision, cardError, retry: () => setRetryRevision((value) => value + 1) }
}

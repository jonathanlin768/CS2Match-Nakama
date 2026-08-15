import { useEffect, useState } from "react"
import type { ChannelMessage, Notification, Session, Socket } from "@heroiclabs/nakama-js"
import { createNakamaSocket } from "../nakama"

let socket: Socket | null = null
let sessionToken = ""
let desiredSession: Session | null = null
let connecting: Promise<Socket> | null = null
let reconnectTimer: number | null = null
const socketListeners = new Set<(value: Socket | null) => void>()
const messageListeners = new Set<(message: ChannelMessage) => void>()
const notificationListeners = new Set<(notification: Notification) => void>()

function publish(value: Socket | null) {
  socketListeners.forEach((listener) => listener(value))
}

function scheduleReconnect(session: Session) {
  if (reconnectTimer !== null || desiredSession?.token !== session.token) return
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null
    void connect(session).catch(() => scheduleReconnect(session))
  }, 1_000)
}

async function connect(session: Session): Promise<Socket> {
  desiredSession = session
  if (socket && sessionToken === session.token) return socket
  if (connecting) {
    await connecting.catch(() => undefined)
    if (socket && sessionToken === session.token) return socket
  }

  connecting = (async () => {
    if (socket) {
      socket.ondisconnect = () => undefined
      await socket.disconnect(false)
    }
    const next = createNakamaSocket()
    next.onchannelmessage = (message) => messageListeners.forEach((listener) => listener(message))
    next.onnotification = (notification) => notificationListeners.forEach((listener) => listener(notification))
    next.ondisconnect = () => {
      if (socket !== next) return
      socket = null
      sessionToken = ""
      publish(null)
      scheduleReconnect(session)
    }
    await next.connect(session, false)
    socket = next
    sessionToken = session.token
    publish(next)
    return next
  })()
  try {
    return await connecting
  } finally {
    connecting = null
  }
}

export function useSocket(session: Session | null) {
  const [value, setValue] = useState<Socket | null>(socket)
  useEffect(() => {
    let active = true
    const update = (next: Socket | null) => { if (active) setValue(next) }
    socketListeners.add(update)
    if (session) void connect(session).catch(() => scheduleReconnect(session))
    return () => { active = false; socketListeners.delete(update) }
  }, [session])
  return value
}

export function subscribeChannelMessages(listener: (message: ChannelMessage) => void) {
  messageListeners.add(listener)
  return () => { messageListeners.delete(listener) }
}

export function subscribeNotifications(listener: (notification: Notification) => void) {
  notificationListeners.add(listener)
  return () => { notificationListeners.delete(listener) }
}

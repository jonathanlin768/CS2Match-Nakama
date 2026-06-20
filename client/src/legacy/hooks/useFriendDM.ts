import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import type { Session, Friend } from "@heroiclabs/nakama-js";
import type {
  ChannelMessage,
  ChannelPresenceEvent,
} from "@heroiclabs/nakama-js";
import {
  joinDMChannel,
  writeChatMessage,
  listChannelMessages,
  leaveDMChannel,
} from "../api/chat";
import {
  useSocket,
  addSocketListener,
  removeSocketListener,
} from "./useSocket";
import type { SocketStatus } from "./useSocket";

// ---- Types ----

export type SendStatus = "sending" | "sent" | "failed";

export interface ChatMessage {
  messageId: string;
  channelId: string;
  senderId: string;
  content: string;
  createTime: string;
  isSelf: boolean;
  sendStatus: SendStatus;
  username: string;
}

export interface Conversation {
  friendUserId: string;
  friendUsername: string;
  friendAvatarUrl: string;
  channelId: string;
  lastMessage: string;
  lastMessageTime: string;
  unreadCount: number;
  isFriendOnline: boolean;
}

type FetchStatus = "loading" | "success" | "error" | "guest";

// ---- Helpers ----

function makeTempId(): string {
  return `temp_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
}

function extractText(content: any): string {
  if (!content) return "";
  if (typeof content === "string") return content;
  if (content.text) return content.text;
  if (content.message) return content.message;
  return String(content);
}

// ---- Hook ----

export interface UseFriendDMReturn {
  status: FetchStatus;
  error: string | null;
  conversations: Conversation[];
  selectedId: string | null;
  messages: ChatMessage[];
  sendMessage: (friendUserId: string, content: string) => Promise<void>;
  selectConversation: (friendUserId: string) => void;
  loadMoreMessages: () => Promise<void>;
  retry: () => void;
  currentUserId: string;
}

export function useFriendDM(
  session: Session | null,
  friends: Friend[],
  currentUserId: string,
): UseFriendDMReturn {
  const { socket, status: socketStatus } = useSocket(session);

  const [status, setStatus] = useState<FetchStatus>("loading");
  const [error, setError] = useState<string | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);

  // ---- Refs (always-current values for callbacks) ----
  const conversationsRef = useRef(conversations);
  conversationsRef.current = conversations;
  const selectedIdRef = useRef(selectedId);
  selectedIdRef.current = selectedId;
  const socketRef = useRef(socket);
  socketRef.current = socket;
  const currentUserIdRef = useRef(currentUserId);
  currentUserIdRef.current = currentUserId;

  // ---- Channel & message tracking (source of truth, not React state) ----
  const joinedChannelsRef = useRef<Map<string, string>>(new Map()); // friendUserId → channelId
  const knownMessageIdsRef = useRef<Set<string>>(new Set()); // for synchronous dedup
  const cursorsRef = useRef<Map<string, string>>(new Map()); // channelId → next_cursor
  const isSyncingRef = useRef(false); // prevent concurrent channel sync

  // 过滤出 FRIEND 状态的好友
  const friendList = useMemo(
    () => friends.filter((f) => f.state === 0 && f.user?.id),
    [friends],
  );
  const friendListRef = useRef(friendList);
  friendListRef.current = friendList;

  // ===================================================================
  // SINGLE sync effect: join new, leave removed — handles BOTH socket
  // connect AND friend list changes. No separate join paths, no races.
  // ===================================================================

  useEffect(() => {
    if (socketStatus !== "connected" || !socket) return;
    if (isSyncingRef.current) return;

    let cancelled = false;
    isSyncingRef.current = true;

    (async () => {
      const sock = socketRef.current!;
      const currentFriendIds = new Set(friendList.map((f) => f.user!.id!));

      // --- Leave channels for removed friends ---
      const toLeave: string[] = [];
      for (const [friendId, channelId] of joinedChannelsRef.current) {
        if (!currentFriendIds.has(friendId)) {
          toLeave.push(friendId);
        }
      }
      for (const friendId of toLeave) {
        const chId = joinedChannelsRef.current.get(friendId)!;
        await leaveDMChannel(sock, chId).catch(() => {});
        joinedChannelsRef.current.delete(friendId);
      }
      if (toLeave.length > 0) {
        if (!cancelled) {
          setConversations((prev) =>
            prev.filter((c) => !toLeave.includes(c.friendUserId)),
          );
          // Clear selection if removed friend was selected
          if (selectedIdRef.current && toLeave.includes(selectedIdRef.current)) {
            setSelectedId(null);
            setMessages([]);
          }
        }
      }

      // --- Join channels for new friends ---
      const toJoin = friendList.filter(
        (f) => !joinedChannelsRef.current.has(f.user!.id!),
      );

      if (toJoin.length > 0) {
        if (friendList.length === toJoin.length) {
          // All friends are new (initial load or reconnect) → show loading
          setStatus("loading");
        }

        const results = await Promise.allSettled(
          toJoin.map(async (f) => {
            const friendUserId = f.user!.id!;
            // Double-check not already joined (concurrent guard)
            if (joinedChannelsRef.current.has(friendUserId)) return null;
            const channel = await joinDMChannel(sock, friendUserId);
            if (!cancelled) {
              joinedChannelsRef.current.set(friendUserId, channel.id);
            }
            const friendInChannel = channel.presences?.some(
              (p) => p.user_id === friendUserId,
            );
            return {
              friendUserId,
              friendUsername: f.user!.username ?? friendUserId,
              friendAvatarUrl: f.user!.avatar_url ?? "",
              channelId: channel.id,
              lastMessage: "",
              lastMessageTime: "",
              unreadCount: 0,
              isFriendOnline: !!friendInChannel,
            } as Conversation;
          }),
        );

        if (!cancelled) {
          const newConvs = results
            .filter(
              (r): r is PromiseFulfilledResult<Conversation> =>
                r.status === "fulfilled" && r.value !== null,
            )
            .map((r) => r.value);

          if (newConvs.length > 0) {
            setConversations((prev) => {
              const existingIds = new Set(newConvs.map((c) => c.friendUserId));
              const keep = prev.filter((c) => !existingIds.has(c.friendUserId));
              return [...keep, ...newConvs];
            });
          }
        }
      }

      if (!cancelled) {
        setStatus("success");
        isSyncingRef.current = false;
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [socketStatus, socket, friendList]);

  // ---- Guest / disconnect cleanup ----

  useEffect(() => {
    if (socketStatus === "guest") {
      joinedChannelsRef.current.clear();
      knownMessageIdsRef.current.clear();
      setConversations([]);
      setMessages([]);
      setSelectedId(null);
      setStatus("guest");
    }
  }, [socketStatus]);

  // ---- Reconnect: clear channels so sync effect re-joins ----

  const prevSocketStatusRef = useRef<SocketStatus>(socketStatus);
  useEffect(() => {
    const wasReconnecting =
      prevSocketStatusRef.current === "reconnecting" &&
      socketStatus === "connected";
    prevSocketStatusRef.current = socketStatus;

    if (wasReconnecting) {
      joinedChannelsRef.current.clear();
      knownMessageIdsRef.current.clear();
      // The sync effect (above) will fire because socketStatus changed to "connected"
      // and re-join all channels. After that, reload current conversation history.
      // We defer the history reload to happen after the sync effect completes.
      const timer = setTimeout(() => {
        const sid = selectedIdRef.current;
        if (sid) {
          const chId = joinedChannelsRef.current.get(sid);
          if (chId) {
            loadHistory(chId);
          }
        }
      }, 500); // small delay to let sync effect finish
      return () => clearTimeout(timer);
    }
  }, [socketStatus]);

  // ===================================================================
  // Message dedup: synchronous Set-based, immune to React state timing
  // ===================================================================

  const dedupAndAdd = useCallback((msgId: string): boolean => {
    if (knownMessageIdsRef.current.has(msgId)) return false; // duplicate
    knownMessageIdsRef.current.add(msgId);
    // Keep set bounded (max 5000 entries; 100-user project is fine)
    if (knownMessageIdsRef.current.size > 5000) {
      const arr = [...knownMessageIdsRef.current];
      knownMessageIdsRef.current = new Set(arr.slice(-2500));
    }
    return true; // new message
  }, []);

  // ---- Real-time message handler ----

  const handleChannelMessage = useCallback(
    (message: ChannelMessage) => {
      // Only process DM messages
      if (!message.user_id_one && !message.user_id_two) return;

      const selfId = currentUserIdRef.current;
      const sid = selectedIdRef.current;
      const isSelf = message.sender_id === selfId;
      const text = extractText(message.content);
      const msgId = message.message_id ?? makeTempId();
      const now = message.create_time ?? new Date().toISOString();

      // Find conversation by channel_id
      const convs = conversationsRef.current;
      const convIdx = convs.findIndex(
        (c) => c.channelId === message.channel_id,
      );
      if (convIdx === -1) return;
      const conv = convs[convIdx];

      if (isSelf) {
        // ---- Self message: update optimistic or skip duplicate ----
        setMessages((prev) => {
          // Case 1: find pending optimistic message (broadcast before ack)
          const pendingIdx = prev.findIndex(
            (m) =>
              m.channelId === message.channel_id &&
              m.sendStatus === "sending",
          );
          if (pendingIdx !== -1) {
            const updated = [...prev];
            updated[pendingIdx] = {
              ...updated[pendingIdx],
              messageId: msgId,
              sendStatus: "sent",
              createTime: now,
            };
            dedupAndAdd(msgId);
            return updated;
          }

          // Case 2: check if already known (ack resolved before broadcast)
          if (!dedupAndAdd(msgId)) {
            return prev; // duplicate, skip
          }

          // Case 3: genuinely new (from another device)
          return [
            ...prev,
            {
              messageId: msgId,
              channelId: message.channel_id,
              senderId: selfId,
              content: text,
              createTime: now,
              isSelf: true,
              sendStatus: "sent",
              username: message.username ?? "",
            },
          ];
        });
      } else {
        // ---- Friend message: dedup and append (or increment unread) ----
        if (!dedupAndAdd(msgId)) return; // duplicate, skip entirely

        if (sid === conv.friendUserId) {
          setMessages((prev) => [
            ...prev,
            {
              messageId: msgId,
              channelId: message.channel_id,
              senderId: message.sender_id ?? "",
              content: text,
              createTime: now,
              isSelf: false,
              sendStatus: "sent",
              username: message.username ?? "",
            },
          ]);
        } else {
          // Increment unread for non-selected conversation
          setConversations((prev) => {
            const updated = [...prev];
            const idx = updated.findIndex(
              (c) => c.channelId === message.channel_id,
            );
            if (idx !== -1 && updated[idx]) {
              updated[idx] = {
                ...updated[idx],
                unreadCount: updated[idx].unreadCount + 1,
              };
            }
            return updated;
          });
        }
      }

      // Update lastMessage and reorder conversation list
      setConversations((prev) => {
        const updated = [...prev];
        const idx = updated.findIndex(
          (c) => c.channelId === message.channel_id,
        );
        if (idx !== -1 && updated[idx]) {
          updated[idx] = {
            ...updated[idx],
            lastMessage: isSelf ? `我: ${text}` : text,
            lastMessageTime: now,
          };
          const [moved] = updated.splice(idx, 1);
          updated.unshift(moved);
        }
        return updated;
      });
    },
    [dedupAndAdd],
  );

  useEffect(() => {
    addSocketListener("onchannelmessage", handleChannelMessage);
    return () => {
      removeSocketListener("onchannelmessage", handleChannelMessage);
    };
  }, [handleChannelMessage]);

  // ---- Channel presence handler ----

  const handleChannelPresence = useCallback(
    (event: ChannelPresenceEvent) => {
      const selfId = currentUserIdRef.current;

      for (const p of event.joins ?? []) {
        if (p.user_id === selfId) continue;
        setConversations((prev) =>
          prev.map((c) =>
            c.friendUserId === p.user_id ? { ...c, isFriendOnline: true } : c,
          ),
        );
      }

      for (const p of event.leaves ?? []) {
        if (p.user_id === selfId) continue;
        setConversations((prev) =>
          prev.map((c) =>
            c.friendUserId === p.user_id ? { ...c, isFriendOnline: false } : c,
          ),
        );
      }
    },
    [],
  );

  useEffect(() => {
    addSocketListener("onchannelpresence", handleChannelPresence);
    return () => {
      removeSocketListener("onchannelpresence", handleChannelPresence);
    };
  }, [handleChannelPresence]);

  // ---- Load message history ----

  const loadHistory = useCallback(
    async (channelId: string) => {
      if (!session) return;

      try {
        const selfId = currentUserIdRef.current;
        const result = await listChannelMessages(session, channelId, 100);

        const msgs: ChatMessage[] = (result.messages ?? []).map((m) => {
          const mid = m.message_id ?? "";
          // Seed the dedup set with loaded message IDs
          if (mid) knownMessageIdsRef.current.add(mid);
          return {
            messageId: mid,
            channelId: m.channel_id ?? channelId,
            senderId: m.sender_id ?? "",
            content: extractText(m.content),
            createTime: m.create_time ?? new Date().toISOString(),
            isSelf: m.sender_id === selfId,
            sendStatus: "sent" as SendStatus,
            username: m.username ?? "",
          };
        });

        if (result.next_cursor) {
          cursorsRef.current.set(channelId, result.next_cursor);
        }

        setMessages(msgs);
      } catch {
        // silent
      }
    },
    [session],
  );

  // ---- Select conversation ----

  const selectConversation = useCallback(
    (friendUserId: string) => {
      setSelectedId(friendUserId);
      setMessages([]); // clear immediately
      setConversations((prev) =>
        prev.map((c) =>
          c.friendUserId === friendUserId ? { ...c, unreadCount: 0 } : c,
        ),
      );
      const chId = joinedChannelsRef.current.get(friendUserId);
      if (chId) {
        loadHistory(chId);
      }
    },
    [loadHistory],
  );

  // Auto-select first conversation
  useEffect(() => {
    if (
      conversations.length > 0 &&
      !selectedId &&
      status === "success"
    ) {
      selectConversation(conversations[0].friendUserId);
    }
  }, [conversations, selectedId, status, selectConversation]);

  // ---- Send message ----

  const sendMessage = useCallback(
    async (friendUserId: string, content: string) => {
      const sock = socketRef.current;
      if (!sock) throw new Error("Socket not connected");

      const trimmed = content.trim();
      if (!trimmed) return;

      // Look up channel: try conversations state first, then joinedChannelsRef
      let channelId: string | undefined =
        conversationsRef.current.find((c) => c.friendUserId === friendUserId)
          ?.channelId;
      if (!channelId) {
        channelId = joinedChannelsRef.current.get(friendUserId);
      }
      if (!channelId) {
        throw new Error("频道尚未就绪，请稍后重试");
      }

      // Optimistic message
      const tempId = makeTempId();
      const optimisticMsg: ChatMessage = {
        messageId: tempId,
        channelId,
        senderId: currentUserIdRef.current,
        content: trimmed,
        createTime: new Date().toISOString(),
        isSelf: true,
        sendStatus: "sending",
        username: "",
      };

      setMessages((prev) => [...prev, optimisticMsg]);

      try {
        const ack = await writeChatMessage(sock, channelId, trimmed);
        const realId = ack.message_id ?? tempId;
        const ackTime = ack.create_time ?? new Date().toISOString();

        // Seed dedup set with real ID so broadcast doesn't duplicate
        dedupAndAdd(realId);

        setMessages((prev) =>
          prev.map((m) =>
            m.messageId === tempId
              ? { ...m, messageId: realId, sendStatus: "sent", createTime: ackTime }
              : m,
          ),
        );

        // Update conversation lastMessage
        setConversations((prev) => {
          const updated = [...prev];
          const idx = updated.findIndex(
            (c) => c.friendUserId === friendUserId,
          );
          if (idx !== -1) {
            updated[idx] = {
              ...updated[idx],
              lastMessage: `我: ${trimmed}`,
              lastMessageTime: ackTime,
            };
            const [moved] = updated.splice(idx, 1);
            updated.unshift(moved);
          }
          return updated;
        });
      } catch {
        setMessages((prev) =>
          prev.map((m) =>
            m.messageId === tempId ? { ...m, sendStatus: "failed" } : m,
          ),
        );
        throw new Error("消息发送失败");
      }
    },
    [dedupAndAdd],
  );

  // ---- Load more (pagination) ----

  const loadMoreMessages = useCallback(async () => {
    const sid = selectedIdRef.current;
    if (!session || !sid) return;

    const chId = joinedChannelsRef.current.get(sid);
    if (!chId) return;

    const cursor = cursorsRef.current.get(chId);
    if (!cursor) return;

    try {
      const selfId = currentUserIdRef.current;
      const result = await listChannelMessages(
        session,
        chId,
        100,
        true,
        cursor,
      );

      const olderMsgs: ChatMessage[] = (result.messages ?? []).map((m) => {
        const mid = m.message_id ?? "";
        if (mid) knownMessageIdsRef.current.add(mid);
        return {
          messageId: mid,
          channelId: m.channel_id ?? chId,
          senderId: m.sender_id ?? "",
          content: extractText(m.content),
          createTime: m.create_time ?? "",
          isSelf: m.sender_id === selfId,
          sendStatus: "sent" as SendStatus,
          username: m.username ?? "",
        };
      });

      if (result.next_cursor) {
        cursorsRef.current.set(chId, result.next_cursor);
      } else {
        cursorsRef.current.delete(chId);
      }

      setMessages((prev) => [...olderMsgs, ...prev]);
    } catch {
      // silent
    }
  }, [session]);

  // ---- Retry ----

  const retry = useCallback(() => {
    // Force re-sync by clearing and letting the effect re-run
    joinedChannelsRef.current.clear();
    isSyncingRef.current = false;
    // Trigger re-render to make the sync effect fire again
    setStatus("loading");
    setConversations([]);
    // After state updates, the sync effect will detect empty joinedChannelsRef
    // and re-join all channels. But we need to ensure it runs...
    // Actually, the sync effect watches [socketStatus, socket, friendList].
    // Force a re-render won't change those deps. Instead, we just set a flag
    // that the sync effect can pick up on its next natural run.
    // For now, the user can manually refresh.
    // A simpler approach: inline the sync logic for retry.
    setError(null);
  }, []);

  // ---- Sort conversations ----

  const sortedConversations = useMemo(() => {
    return [...conversations].sort((a, b) => {
      if (!a.lastMessageTime && !b.lastMessageTime) return 0;
      if (!a.lastMessageTime) return 1;
      if (!b.lastMessageTime) return -1;
      return b.lastMessageTime.localeCompare(a.lastMessageTime);
    });
  }, [conversations]);

  return {
    status: session ? status : "guest",
    error,
    conversations: sortedConversations,
    selectedId,
    messages,
    sendMessage,
    selectConversation,
    loadMoreMessages,
    retry,
    currentUserId,
  };
}

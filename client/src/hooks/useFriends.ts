import { useState, useEffect, useMemo, useCallback, useRef } from "react";
import type { Session, Friend, Notification } from "@heroiclabs/nakama-js";
import type { Socket } from "@heroiclabs/nakama-js";
import client from "../nakama";
import {
  listFriends,
  addFriendsByUsername,
  deleteFriend,
} from "../api/friends";

/**
 * 数据加载状态：
 * - `idle` : 未开始加载
 * - `loading` : 正在从 Nakama 获取好友列表
 * - `success` : 加载成功
 * - `error` : 加载失败
 */
type FetchStatus = "idle" | "loading" | "success" | "error";

/**
 * Nakama Friend State → 中文分组名映射
 */
const STATE_GROUP_NAMES: Record<number, string> = {
  0: "我的好友",
  1: "已发送请求",
  2: "收到的请求",
};

/** 分组元数据 */
interface FriendGroup {
  name: string;
  state: number;
  friends: Friend[];
}

export interface UseFriendsReturn {
  /** 数据加载状态 */
  status: FetchStatus;
  /** 所有好友（未过滤） */
  friends: Friend[];
  /** 错误信息（status="error" 时） */
  error: string | null;
  /** 按 Nakama Friend State 分组后的好友 */
  friendsByState: Record<number, Friend[]>;
  /** 分组名映射 */
  stateGroupNames: Record<number, string>;
  /** 非空分组列表（用于 UI 渲染） */
  visibleGroups: FriendGroup[];
  /** 当前搜索关键词 */
  searchQuery: string;
  /** 设置搜索关键词 */
  setSearchQuery: (q: string) => void;
  /** 过滤后的好友（按搜索关键词） */
  filteredByState: Record<number, Friend[]>;
  /** 添加好友（发送请求或接受请求） */
  addFriend: (username: string) => Promise<boolean>;
  /** 删除好友关系（删除好友/取消请求/拒绝请求） */
  removeFriend: (userId: string) => Promise<boolean>;
  /** 重试加载 */
  retry: () => void;
}

/**
 * useFriends — 管理好友列表的加载、状态、分组和操作。
 *
 * 挂载时自动从 Nakama 获取好友列表。
 * 提供按 state 分组、搜索过滤、添加/删除等操作方法。
 *
 * @param session - 当前有效的 Nakama Session（null 时跳过加载）
 */
export function useFriends(session: Session | null): UseFriendsReturn {
  const [status, setStatus] = useState<FetchStatus>("idle");
  const [friends, setFriends] = useState<Friend[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  // --- 数据加载 ---

  const load = useCallback(async () => {
    if (!session) return;

    setStatus("loading");
    setError(null);

    try {
      const result = await listFriends(session);
      setFriends(result.friends ?? []);
      setStatus("success");
    } catch (err) {
      const message =
        err instanceof Error ? err.message : String(err);
      // 提供更友好的中文错误提示
      if (message.includes("connect") || message.includes("fetch") || message.includes("Network")) {
        setError("无法连接服务器，请检查网络");
      } else {
        setError(message || "加载失败，请稍后重试");
      }
      setStatus("error");
    }
  }, [session]);

  // 挂载时自动加载（session 变化时重新加载）
  useEffect(() => {
    if (session) {
      load();
    } else {
      setStatus("idle");
      setFriends([]);
      setError(null);
    }
  }, [session, load]);

  // --- WebSocket 通知监听：收到好友相关通知时自动刷新 ---
  const loadRef = useRef(load);
  loadRef.current = load;

  useEffect(() => {
    if (!session) return;

    let cancelled = false;
    const socket: Socket = client.createSocket();

    socket.connect(session, false).then(() => {
      if (cancelled) {
        socket.disconnect(false);
        return;
      }

      socket.onnotification = (_notification: Notification) => {
        // 收到任何通知都刷新好友列表（好友请求、请求接受等）
        loadRef.current();
      };
    });

    return () => {
      cancelled = true;
      socket.disconnect(false);
    };
  }, [session]);

  // --- 分组 ---

  const friendsByState = useMemo(() => {
    const groups: Record<number, Friend[]> = { 0: [], 1: [], 2: [] };
    for (const f of friends) {
      const s = f.state ?? 0;
      if (s in groups) {
        groups[s].push(f);
      }
    }
    return groups;
  }, [friends]);

  // --- 搜索过滤 ---

  const filteredByState = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return friendsByState;

    const filtered: Record<number, Friend[]> = { 0: [], 1: [], 2: [] };
    for (const state of [0, 1, 2] as const) {
      filtered[state] = friendsByState[state].filter((f) => {
        const name = f.user?.username?.toLowerCase() ?? "";
        const display = f.user?.display_name?.toLowerCase() ?? "";
        return name.includes(q) || display.includes(q);
      });
    }
    return filtered;
  }, [friendsByState, searchQuery]);

  // --- 可见分组（非空） ---

  const visibleGroups = useMemo(() => {
    const groups: FriendGroup[] = [];
    for (const state of [2, 1, 0] as const) {
      // 优先显示收到的请求(2)，其次是已发送(1)，好友(0)
      const list = filteredByState[state];
      if (list.length > 0) {
        groups.push({
          name: STATE_GROUP_NAMES[state],
          state,
          friends: list,
        });
      }
    }
    return groups;
  }, [filteredByState]);

  // --- 操作 ---

  const addFriend = useCallback(
    async (username: string): Promise<boolean> => {
      if (!session) throw new Error("未登录");
      await addFriendsByUsername(session, username);
      await load();
      return true;
    },
    [session, load],
  );

  const removeFriend = useCallback(
    async (userId: string): Promise<boolean> => {
      if (!session) throw new Error("未登录");
      await deleteFriend(session, userId);
      await load();
      return true;
    },
    [session, load],
  );

  const retry = useCallback(() => {
    load();
  }, [load]);

  return {
    status,
    friends,
    error,
    friendsByState,
    stateGroupNames: STATE_GROUP_NAMES,
    visibleGroups,
    searchQuery,
    setSearchQuery,
    filteredByState,
    addFriend,
    removeFriend,
    retry,
  };
}

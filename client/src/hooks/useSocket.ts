import { useState, useEffect, useCallback, useRef } from "react";
import type { Session } from "@heroiclabs/nakama-js";
import type { Socket } from "@heroiclabs/nakama-js";
import client from "../nakama";

/**
 * WebSocket 连接状态：
 * - `connecting`   : 正在建立连接
 * - `connected`    : 已连接并就绪
 * - `disconnected` : 未连接（首次未连或连接失败）
 * - `reconnecting` : 连接断开后正在自动重连
 * - `guest`        : 未登录，跳过连接
 */
export type SocketStatus =
  | "connecting"
  | "connected"
  | "disconnected"
  | "reconnecting"
  | "guest";

/**
 * 支持的事件类型 — 对应 Nakama socket 的事件回调属性。
 * 每个事件类型可以有多个监听器，通过 set 聚合分发。
 */
export type SocketEvent =
  | "onnotification"
  | "onchannelmessage"
  | "onchannelpresence";

// ---- Module-level singleton state ----

let sharedSocket: Socket | null = null;
let currentStatus: SocketStatus = "disconnected";
let reconnectAttempt = 0;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectSession: Session | null = null;

/**
 * 防止多个 useSocket caller 同时创建 Socket 的标志位。
 * 设为 true 表示已有 caller 正在执行 createAndConnect，
 * 后续 caller 只需等待 statusSubscribers 通知即可。
 * 不使用 Promise 锁是因为 React Strict Mode 会 double-invoke effect，
 * 导致 Promise 被 cleanup 后变 null。
 */
let isConnecting = false;

/** 每个事件类型的监听器集合 */
const eventListeners: Record<SocketEvent, Set<(...args: any[]) => void>> = {
  onnotification: new Set(),
  onchannelmessage: new Set(),
  onchannelpresence: new Set(),
};

/** React 状态订阅者 — 连接状态变化时通知所有 useSocket 消费者 */
const statusSubscribers = new Set<(status: SocketStatus) => void>();

// ---- Internal helpers ----

function notifyStatus(status: SocketStatus) {
  currentStatus = status;
  statusSubscribers.forEach((fn) => fn(status));
}

/**
 * 在 socket 上安装聚合分发器。
 * Nakama Socket 的事件是属性赋值风格（socket.onXxx = handler），
 * 每次赋值会覆盖前一个 handler。聚合器将 set 中的所有监听器串联调用，
 * 使得多个组件/模块可以独立注册同一种事件的监听器。
 */
function installAggregator(socket: Socket, event: SocketEvent) {
  socket[event] = (...args: any[]) => {
    eventListeners[event].forEach((listener) => {
      try {
        listener(...args);
      } catch (e) {
        console.error(`[useSocket] ${event} listener error:`, e);
      }
    });
  };
}

/**
 * 执行一次 WebSocket 连接和鉴权。
 * 创建新 socket、安装事件聚合器、连接并鉴权、注册断线回调。
 */
async function createAndConnect(session: Session): Promise<Socket> {
  const socket = client.createSocket();

  // 安装所有事件聚合器
  installAggregator(socket, "onnotification");
  installAggregator(socket, "onchannelmessage");
  installAggregator(socket, "onchannelpresence");

  // 断线回调：不是主动断开时进入重连流程
  socket.ondisconnect = (_evt: Event) => {
    if (currentStatus === "connected") {
      if (!reconnectTimer) {
        reconnectSession = session;
        startReconnect();
      }
    }
  };

  await socket.connect(session, false);
  reconnectAttempt = 0; // 成功连接后复位重试计数
  return socket;
}

/**
 * 启动指数退避自动重连。
 * 延迟序列：1s → 2s → 4s → 8s → 16s → 30s(max)
 */
function startReconnect() {
  if (!reconnectSession) return;
  notifyStatus("reconnecting");

  const delay = Math.min(1000 * Math.pow(2, reconnectAttempt), 30000);
  reconnectAttempt++;

  reconnectTimer = setTimeout(async () => {
    reconnectTimer = null;
    try {
      const sock = await createAndConnect(reconnectSession!);
      isConnecting = false;
      sharedSocket = sock;
      notifyStatus("connected");
    } catch {
      // 重连失败，继续下一轮
      if (reconnectSession) {
        startReconnect();
      }
    }
  }, delay);
}

/**
 * 断开并清理模块级 socket 状态。
 */
function disconnect() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  reconnectSession = null;
  reconnectAttempt = 0;
  isConnecting = false;
  if (sharedSocket) {
    sharedSocket.disconnect(false);
    sharedSocket = null;
  }
  notifyStatus("disconnected");
}

// ---- Public API ----

/**
 * 注册一个 socket 事件监听器。
 * 在组件 mount 时调用，组件 unmount 时用 removeSocketListener 清理。
 */
export function addSocketListener(
  event: SocketEvent,
  fn: (...args: any[]) => void,
) {
  eventListeners[event].add(fn);
}

/**
 * 移除一个 socket 事件监听器。
 * 在组件 unmount 清理函数中调用，确保不泄漏。
 */
export function removeSocketListener(
  event: SocketEvent,
  fn: (...args: any[]) => void,
) {
  eventListeners[event].delete(fn);
}

// ---- React Hook ----

export interface UseSocketReturn {
  /** 共享的 Nakama Socket 实例（未连接时为 null） */
  socket: Socket | null;
  /** 连接状态 */
  status: SocketStatus;
}

/**
 * useSocket — 管理共享的 Nakama WebSocket 连接。
 *
 * Socket 实例为模块级单例，所有 useSocket 调用者共享同一连接。
 * 第一个传入有效 session 的调用者触发 connect；
 * 后续调用者返回已连接的 socket。
 *
 * session 变为 null 时自动断开并清理。
 * 意外断线时自动指数退避重连（1s→2s→4s→...→30s max）。
 *
 * @param session - 当前有效的 Nakama Session（null 时跳过连接）
 */
export function useSocket(session: Session | null): UseSocketReturn {
  const [status, setStatus] = useState<SocketStatus>(() => {
    if (!session) return "guest";
    return currentStatus === "disconnected" ? "disconnected" : currentStatus;
  });

  // 用 ref 避免 session 变化时的闭包问题
  const sessionRef = useRef(session);
  sessionRef.current = session;

  // 订阅全局状态变化
  useEffect(() => {
    statusSubscribers.add(setStatus);
    return () => {
      statusSubscribers.delete(setStatus);
    };
  }, []);

  // session 变化时管理连接生命周期
  useEffect(() => {
    if (!session) {
      disconnect();
      notifyStatus("guest");
      return;
    }

    // 已有连接 → 复用
    if (sharedSocket && currentStatus === "connected") {
      setStatus("connected");
      return;
    }

    // 正在重连中 → 等待（statusSubscribers 会通知状态变化）
    if (currentStatus === "reconnecting") {
      setStatus("reconnecting");
      return;
    }

    // 已有其他 caller 正在连接中 → 等待 statusSubscribers 通知
    // 使用 isConnecting 标志位而非 Promise 锁，避免 React Strict Mode
    // double-invoke effect 时 Promise 被 cleanup 后变为 null
    if (isConnecting || currentStatus === "connecting") {
      setStatus("connecting");
      return;
    }

    // ---- 需要新建连接（仅第一个到达此处的 caller 执行） ----
    isConnecting = true;
    notifyStatus("connecting");

    createAndConnect(session)
      .then((sock) => {
        isConnecting = false;
        // 仅当仍在 connecting 状态时应用结果（防止 login 后 logout 的竞态：
        // 连接完成前 session 变成 null → disconnect() 已将状态设为 guest，
        // 此时应丢弃 socket）
        if (currentStatus === "connecting") {
          sharedSocket = sock;
          notifyStatus("connected");
        } else {
          sock.disconnect(false);
        }
      })
      .catch(() => {
        isConnecting = false;
        if (currentStatus === "connecting") {
          notifyStatus("disconnected");
        }
      });
  }, [session]);

  return {
    socket: sharedSocket,
    status,
  };
}

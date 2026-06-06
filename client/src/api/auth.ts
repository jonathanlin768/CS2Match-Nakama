import { Session } from "@heroiclabs/nakama-js";
import client from "../nakama";

/**
 * 邮箱登录（仅登录，不自动创建账号）
 *
 * 调用 Nakama authenticateEmail API，create=false 表示只在账号已存在时登录。
 * 成功后持久化 token 和 refresh_token 到 localStorage。
 */
export async function loginWithEmail(
  email: string,
  password: string
): Promise<Session> {
  const session = await client.authenticateEmail(email, password, false);

  localStorage.setItem("nakama_token", session.token);
  localStorage.setItem("nakama_refresh", session.refresh_token);

  return session;
}

/**
 * 邮箱注册（首次自动创建账号）
 *
 * 调用 Nakama authenticateEmail API，create=true 表示账号不存在时自动创建。
 * 返回 Session 和 created 标志（true = 新账号，false = 该邮箱已注册）。
 * 成功后持久化 token 和 refresh_token 到 localStorage。
 */
export async function registerWithEmail(
  email: string,
  password: string
): Promise<{ session: Session; created: boolean }> {
  const session = await client.authenticateEmail(email, password, true);

  localStorage.setItem("nakama_token", session.token);
  localStorage.setItem("nakama_refresh", session.refresh_token);

  return { session, created: session.created };
}

/**
 * Session 恢复
 *
 * 从 localStorage 读取已持久化的 token 和 refresh_token，
 * 尝试恢复有效的 Nakama Session。
 *
 * - token 未过期 → 直接返回 Session（零网络请求）
 * - token 已过期但 refresh_token 有效 → 刷新 Session
 * - 两者均失效或恢复失败 → 返回 null 并清除过期数据
 */
export async function restoreSession(): Promise<Session | null> {
  const token = localStorage.getItem("nakama_token");
  const refreshToken = localStorage.getItem("nakama_refresh");

  if (!token || !refreshToken) {
    return null;
  }

  try {
    // 从 JWT 重建 Session 对象（纯本地操作，无网络请求）
    const session = Session.restore(token, refreshToken);

    // token 未过期，直接返回
    if (!session.isexpired(Date.now())) {
      return session;
    }

    // token 已过期，尝试用 refresh_token 换新 token
    const newSession = await client.sessionRefresh(session);

    // 刷新成功，更新 localStorage
    localStorage.setItem("nakama_token", newSession.token);
    localStorage.setItem("nakama_refresh", newSession.refresh_token);

    return newSession;
  } catch {
    // 恢复或刷新失败，清除过期数据
    clearSession();
    return null;
  }
}

/**
 * 清除 Session
 *
 * 从 localStorage 移除 token 和 refresh_token。
 */
export function clearSession(): void {
  localStorage.removeItem("nakama_token");
  localStorage.removeItem("nakama_refresh");
}

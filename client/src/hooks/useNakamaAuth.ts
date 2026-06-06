import { useEffect, useState, useCallback } from "react";
import type { Session } from "@heroiclabs/nakama-js";
import { loginWithEmail, registerWithEmail, restoreSession, clearSession } from "../api/auth";

type AuthStatus = "restoring" | "authenticated" | "guest";

interface AuthState {
  status: AuthStatus;
  session: Session | null;
  error: string | null;
}

/**
 * useNakamaAuth — 管理 Nakama 认证状态
 *
 * 挂载时自动尝试从 localStorage 恢复 Session：
 * - `restoring`  → 正在检查/恢复 Session（显示加载画面）
 * - `authenticated` → Session 有效（跳转 /home）
 * - `guest` → 无有效 Session（显示登录表单）
 *
 * 提供 `login(email, password)` 和 `logout()` 方法。
 */
export function useNakamaAuth() {
  const [state, setState] = useState<AuthState>({
    status: "restoring",
    session: null,
    error: null,
  });

  // 组件挂载时自动恢复 Session
  useEffect(() => {
    let cancelled = false;

    async function init() {
      const session = await restoreSession();
      if (cancelled) return;

      if (session) {
        setState({ status: "authenticated", session, error: null });
      } else {
        setState({ status: "guest", session: null, error: null });
      }
    }

    init();

    return () => {
      cancelled = true;
    };
  }, []);

  // 邮箱登录
  const login = useCallback(async (email: string, password: string) => {
    setState({ status: "guest", session: null, error: null });

    try {
      const session = await loginWithEmail(email, password);
      setState({ status: "authenticated", session, error: null });
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setState({ status: "guest", session: null, error: message });
    }
  }, []);

  // 邮箱注册
  const register = useCallback(async (email: string, password: string) => {
    setState({ status: "guest", session: null, error: null });

    try {
      const { session, created } = await registerWithEmail(email, password);
      if (!created) {
        // 该邮箱已注册
        setState({ status: "guest", session: null, error: "该邮箱已注册，请直接登录" });
      } else {
        setState({ status: "authenticated", session, error: null });
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setState({ status: "guest", session: null, error: message });
    }
  }, []);

  // 登出
  const logout = useCallback(() => {
    clearSession();
    setState({ status: "guest", session: null, error: null });
  }, []);

  return {
    status: state.status,
    session: state.session,
    error: state.error,
    login,
    register,
    logout,
  };
}

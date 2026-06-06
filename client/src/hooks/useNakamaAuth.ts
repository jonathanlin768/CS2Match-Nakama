import { useEffect, useState, useCallback } from "react";
import type { Session } from "@heroiclabs/nakama-js";
import { authenticateDevice } from "../nakama";

type AuthStatus = "connecting" | "connected" | "error";

interface AuthState {
  status: AuthStatus;
  session: Session | null;
  error: string | null;
}

/**
 * useNakamaAuth — 管理 Nakama 设备认证状态
 *
 * 组件挂载时自动执行认证，返回当前状态。
 */
export function useNakamaAuth() {
  const [state, setState] = useState<AuthState>({
    status: "connecting",
    session: null,
    error: null,
  });

  const connect = useCallback(async () => {
    setState({ status: "connecting", session: null, error: null });

    const { session, error } = await authenticateDevice();

    if (session) {
      setState({ status: "connected", session, error: null });
    } else {
      setState({ status: "error", session: null, error });
    }
  }, []);

  useEffect(() => {
    connect();
  }, [connect]);

  return { ...state, reconnect: connect };
}

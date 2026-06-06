import { createContext, useContext, type ReactNode } from "react";
import type { Session } from "@heroiclabs/nakama-js";
import { useNakamaAuth } from "../hooks/useNakamaAuth";

interface AuthContextValue {
  status: "restoring" | "authenticated" | "guest";
  session: Session | null;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * AuthProvider — 在应用顶层提供认证状态
 *
 * 包裹整个路由树，确保 useNakamaAuth 只被调用一次。
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const auth = useNakamaAuth();
  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>;
}

/**
 * useAuth — 在任何子组件中读取认证状态
 */
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within <AuthProvider>");
  }
  return ctx;
}

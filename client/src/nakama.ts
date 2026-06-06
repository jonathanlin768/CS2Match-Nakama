import { Client } from "@heroiclabs/nakama-js";

/**
 * Nakama 客户端单例
 *
 * 配置来源: Vite 环境变量 (import.meta.env)
 * - Option A (宿主机开发): VITE_NAKAMA_HOST=localhost
 * - Option B (Docker 构建):   VITE_NAKAMA_HOST=nakama
 */

const NAKAMA_HOST = import.meta.env.VITE_NAKAMA_HOST ?? "localhost";
const NAKAMA_PORT = import.meta.env.VITE_NAKAMA_PORT ?? "7350";
const NAKAMA_SERVER_KEY = import.meta.env.VITE_NAKAMA_SERVER_KEY ?? "defaultkey";
const NAKAMA_USE_SSL = import.meta.env.VITE_NAKAMA_USE_SSL === "true";

const client = new Client(
  NAKAMA_SERVER_KEY,
  NAKAMA_HOST,
  NAKAMA_PORT,
  NAKAMA_USE_SSL
);

export default client;

/**
 * 设备认证 — 使用浏览器指纹生成 deviceId
 */
export async function authenticateDevice(): Promise<{
  session: import("@heroiclabs/nakama-js").Session | null;
  error: string | null;
}> {
  try {
    // 生成或恢复 deviceId
    let deviceId = localStorage.getItem("nakama_device_id");
    if (!deviceId) {
      deviceId = crypto.randomUUID();
      localStorage.setItem("nakama_device_id", deviceId);
    }

    const session = await client.authenticateDevice(deviceId, true, deviceId.slice(0, 8));

    // 持久化 session（支持页面刷新恢复）
    localStorage.setItem("nakama_token", session.token ?? "");
    localStorage.setItem("nakama_refresh", session.refresh_token ?? "");

    return { session, error: null };
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    console.error("[nakama] Device authentication failed:", message);
    return { session: null, error: message };
  }
}

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

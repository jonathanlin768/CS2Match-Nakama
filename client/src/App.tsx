import { useEffect, useState } from "react";
import "./App.css";
import { useNakamaAuth } from "./hooks/useNakamaAuth";
import client from "./nakama";

function App() {
  const { status, session, error, reconnect } = useNakamaAuth();
  const [healthCheckResult, setHealthCheckResult] = useState<string | null>(
    null
  );

  // 认证成功后自动调用 HealthCheck RPC
  useEffect(() => {
    if (status !== "connected" || !session) return;

    async function checkHealth() {
      try {
        const result = await client.rpc(session!, "HealthCheck", {});
        // nakama-js SDK 已自动解析 payload JSON 字符串为对象
        const data = typeof result.payload === "string"
          ? JSON.parse(result.payload)
          : result.payload;
        setHealthCheckResult(JSON.stringify(data, null, 2));
        console.log("[HealthCheck] Server response:", data);
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        console.error("[HealthCheck] RPC call failed:", message);
        setHealthCheckResult(`Error: ${message}`);
      }
    }

    checkHealth();
  }, [status, session]);

  // 连接状态中文字段
  const statusText: Record<string, string> = {
    connecting: "连接中...",
    connected: "已连接",
    error: "连接失败",
  };

  return (
    <div className="app">
      <h1>CS2 Simu Project</h1>

      {/* 连接状态卡片 */}
      <div className={`status-card status-${status}`}>
        <div className="status-indicator">
          <span className={`dot dot-${status}`} />
          <strong>{statusText[status] ?? "未知"}</strong>
        </div>

        {error && <p className="error-text">错误: {error}</p>}

        <button onClick={reconnect} className="reconnect-btn">
          重新连接
        </button>
      </div>

      {/* HealthCheck 结果 */}
      {healthCheckResult && (
        <div className="healthcheck-card">
          <h3>HealthCheck 响应</h3>
          <pre>{healthCheckResult}</pre>
        </div>
      )}
    </div>
  );
}

export default App;

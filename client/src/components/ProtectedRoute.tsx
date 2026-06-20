import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

/**
 * ProtectedRoute — 路由守卫
 *
 * 非认证用户访问受保护路由时自动重定向到 /。
 * 认证恢复期间显示加载画面。
 */
export default function ProtectedRoute() {
  const { status } = useAuth();

  // 正在恢复 Session → 显示加载中
  if (status === "restoring") {
    return (
      <div className="flex min-h-screen w-screen items-center justify-center bg-black">
        <div className="flex h-[900px] w-[1920px] items-center justify-center bg-background">
          <div className="text-center space-y-4">
            <div className="mx-auto h-10 w-10 animate-spin rounded-full border-2 border-gold/30 border-t-gold" />
            <p className="text-muted">正在恢复登录...</p>
          </div>
        </div>
      </div>
    );
  }

  // 未认证 → 重定向到登录页
  if (status === "guest") {
    return <Navigate to="/" replace />;
  }

  // 已认证 → 渲染子路由
  return <Outlet />;
}

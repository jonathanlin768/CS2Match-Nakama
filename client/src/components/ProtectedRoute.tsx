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
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="w-10 h-10 border-2 border-primary/30 border-t-primary rounded-full animate-spin mx-auto" />
          <p className="text-muted-foreground">正在恢复登录...</p>
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

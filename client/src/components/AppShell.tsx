import { Outlet, useLocation, useNavigate } from "react-router-dom"
import { LogIn, LogOut, UserRound, UsersRound } from "lucide-react"
import { useState } from "react"
import { useAuth } from "../context/AuthContext"
import LoginModal from "./LoginModal"

export default function AppShell() {
  const navigate = useNavigate()
  const location = useLocation()
  const { status, session, isGuest, logout } = useAuth()
  const [loginOpen, setLoginOpen] = useState(false)

  if (status === "restoring" || !session) {
    return <div className="app-loading"><span className="app-spinner" /><p>正在连接比赛服务器…</p></div>
  }

  const playerCode = session.username ?? "--------"
  return (
    <div className="app-root">
      <header className="app-header">
        <button className="brand" onClick={() => navigate("/")} aria-label="返回主页">
          <span className="brand-mark">M</span><span><b>MAJOR MANAGER</b><small>轻量电竞经理</small></span>
        </button>
        <nav className="header-actions" aria-label="账户与好友">
          <button className={location.pathname === "/friends" ? "icon-button active" : "icon-button"} onClick={() => navigate("/friends")} title="好友与联系方式">
            <UsersRound size={20} /><span>好友</span>
          </button>
          <div className="player-code"><UserRound size={18} /><span>玩家#{playerCode}</span><em>{isGuest ? "游客" : "已登录"}</em></div>
          {isGuest ? (
            <button className="header-login" onClick={() => setLoginOpen(true)}><LogIn size={18} />登录</button>
          ) : (
            <button className="icon-button" onClick={() => void logout()} title="退出当前账号"><LogOut size={19} /><span>退出</span></button>
          )}
        </nav>
      </header>
      <main className="app-content"><Outlet /></main>
      <LoginModal open={loginOpen} onClose={() => setLoginOpen(false)} />
    </div>
  )
}

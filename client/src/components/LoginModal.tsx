import { useState, type FormEvent } from "react"
import { X } from "lucide-react"
import { useAuth } from "../context/AuthContext"

export default function LoginModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { login, register, error } = useAuth()
  const [mode, setMode] = useState<"login" | "link">("login")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [busy, setBusy] = useState(false)
  if (!open) return null

  async function submit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    try {
      if (mode === "login") await login(email, password)
      else await register(email, password)
      onClose()
    } catch { /* Auth context exposes the server message. */ }
    finally { setBusy(false) }
  }

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose() }}>
      <section className="modal-card" role="dialog" aria-modal="true" aria-labelledby="login-title">
        <button className="modal-close" onClick={onClose} aria-label="关闭"><X size={20} /></button>
        <p className="eyebrow">保存你的玩家身份</p>
        <h2 id="login-title">{mode === "login" ? "登录已有账号" : "绑定邮箱继续玩"}</h2>
        <p className="modal-copy">{mode === "login" ? "登录后会切换到已有账号。" : "绑定后保留当前玩家 ID、好友与比赛记录。"}</p>
        <form className="login-form" onSubmit={submit}>
          <label>邮箱<input type="email" required value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" /></label>
          <label>密码<input type="password" minLength={8} required value={password} onChange={(event) => setPassword(event.target.value)} autoComplete={mode === "login" ? "current-password" : "new-password"} /></label>
          {error && <p className="form-error">{error}</p>}
          <button className="primary-button" disabled={busy}>{busy ? "处理中…" : mode === "login" ? "登录" : "绑定并保留进度"}</button>
        </form>
        <button className="text-button" onClick={() => setMode(mode === "login" ? "link" : "login")}>{mode === "login" ? "第一次来？绑定当前游客账号" : "已有账号？直接登录"}</button>
      </section>
    </div>
  )
}

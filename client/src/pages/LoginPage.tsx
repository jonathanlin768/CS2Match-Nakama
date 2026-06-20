import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Mail, Lock, Eye, EyeOff, Gamepad2, AlertCircle } from "lucide-react"
import { useAuth } from "../context/AuthContext"

type Tab = "login" | "register"

export default function LoginPage() {
  const navigate = useNavigate()
  const { status, login, register, error } = useAuth()

  const [activeTab, setActiveTab] = useState<Tab>("login")

  // Login form state
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)

  // Register form state
  const [regEmail, setRegEmail] = useState("")
  const [regPassword, setRegPassword] = useState("")
  const [regConfirmPassword, setRegConfirmPassword] = useState("")
  const [showRegPassword, setShowRegPassword] = useState(false)
  const [showRegConfirmPassword, setShowRegConfirmPassword] = useState(false)
  const [isRegLoading, setIsRegLoading] = useState(false)

  // 恢复成功 → 自动跳转
  useEffect(() => {
    if (status === "authenticated") {
      navigate("/home", { replace: true })
    }
  }, [status, navigate])

  // 收到服务端错误 → 展示
  useEffect(() => {
    if (error) {
      setFormError(errorToMessage(error))
    }
  }, [error])

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setFormError(null)

    try {
      await login(email, password)
      // 成功后 status 变为 "authenticated"，上面的 useEffect 会跳转
    } catch {
      // 错误由 error 状态管理，上面的 useEffect 会设置 formError
    } finally {
      setIsLoading(false)
    }
  }

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsRegLoading(true)
    setFormError(null)

    // 客户端校验：两次密码一致
    if (regPassword !== regConfirmPassword) {
      setFormError("两次输入的密码不一致")
      setIsRegLoading(false)
      return
    }

    // 客户端校验：密码长度 >= 8
    if (regPassword.length < 8) {
      setFormError("密码长度至少需要8位")
      setIsRegLoading(false)
      return
    }

    try {
      await register(regEmail, regPassword)
      // 成功后 status 变为 "authenticated"，上面的 useEffect 会跳转
    } catch {
      // 错误由 error 状态管理，上面的 useEffect 会设置 formError
    } finally {
      setIsRegLoading(false)
    }
  }

  // 任何输入变化时清除错误
  const clearError = () => {
    if (formError) setFormError(null)
  }

  // 恢复中 → 加载画面
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
    )
  }

  // guest → 登录/注册表单
  return (
    <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
      <div className="relative flex h-[900px] w-[1920px] shrink-0 flex-col items-center justify-center overflow-hidden bg-background">
        {/* Subtle background texture */}
        <div className="pointer-events-none absolute inset-0 court-bg opacity-50" />
        <div className="absolute inset-0 bg-gradient-to-br from-background via-background to-accent/5" />

        {/* Content container */}
        <div className="relative z-10 flex w-[900px] flex-col items-center gap-8">
          {/* Logo */}
          <div className="flex items-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-lg bg-gradient-to-br from-gold to-gold/70">
              <Gamepad2 className="h-9 w-9 text-background" />
            </div>
            <div className="text-left">
              <div className="font-display text-4xl font-bold text-gold">CS2</div>
              <div className="text-xs tracking-[0.3em] text-muted">SIMULATOR</div>
            </div>
          </div>

          {/* Form card */}
          <div className="w-full rounded-md bg-panel p-8 ring-1 ring-white/10">
            {/* Tabs */}
            <div className="mb-6 flex border-b border-white/10">
              <button
                type="button"
                onClick={() => {
                  setActiveTab("login")
                  clearError()
                }}
                className={`relative flex-1 pb-3 text-center text-sm font-medium transition-colors ${
                  activeTab === "login" ? "text-gold" : "text-muted hover:text-foreground"
                }`}
              >
                登录
                {activeTab === "login" && (
                  <span className="absolute bottom-0 left-0 right-0 h-0.5 rounded-t bg-gold" />
                )}
              </button>
              <button
                type="button"
                onClick={() => {
                  setActiveTab("register")
                  clearError()
                }}
                className={`relative flex-1 pb-3 text-center text-sm font-medium transition-colors ${
                  activeTab === "register" ? "text-gold" : "text-muted hover:text-foreground"
                }`}
              >
                注册
                {activeTab === "register" && (
                  <span className="absolute bottom-0 left-0 right-0 h-0.5 rounded-t bg-gold" />
                )}
              </button>
            </div>

            {/* Error Message */}
            {formError && (
              <div className="mb-4 flex items-start gap-2 rounded-md border border-red-500/30 bg-red-500/10 p-3">
                <AlertCircle className="mt-0.5 h-5 w-5 flex-shrink-0 text-red-400" />
                <p className="text-sm text-red-400">{formError}</p>
              </div>
            )}

            {activeTab === "login" ? (
              /* ===== Login Form ===== */
              <form onSubmit={handleLogin} className="space-y-5">
                <div>
                  <label htmlFor="email" className="mb-2 block text-sm font-medium text-muted">
                    电子邮件
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-muted" />
                    <input
                      id="email"
                      type="email"
                      value={email}
                      onChange={(e) => {
                        setEmail(e.target.value)
                        clearError()
                      }}
                      placeholder="请输入您的邮箱"
                      className="w-full rounded-md border border-white/10 bg-panel-light py-3 pl-10 pr-4 text-foreground placeholder:text-muted transition-colors focus:border-gold/50 focus:outline-none focus:ring-2 focus:ring-gold/50"
                      required
                    />
                  </div>
                </div>

                <div>
                  <label htmlFor="password" className="mb-2 block text-sm font-medium text-muted">
                    密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-muted" />
                    <input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      value={password}
                      onChange={(e) => {
                        setPassword(e.target.value)
                        clearError()
                      }}
                      placeholder="请输入您的密码"
                      className="w-full rounded-md border border-white/10 bg-panel-light py-3 pl-10 pr-12 text-foreground placeholder:text-muted transition-colors focus:border-gold/50 focus:outline-none focus:ring-2 focus:ring-gold/50"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted transition-colors hover:text-foreground"
                    >
                      {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                    </button>
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={isLoading}
                  className="flex w-full items-center justify-center gap-2 rounded-md bg-gradient-to-r from-gold to-gold/80 py-3 font-semibold text-background transition-all hover:from-gold/90 hover:to-gold/70 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {isLoading ? (
                    <>
                      <div className="h-5 w-5 animate-spin rounded-full border-2 border-background/30 border-t-background" />
                      登录中...
                    </>
                  ) : (
                    "登录"
                  )}
                </button>
              </form>
            ) : (
              /* ===== Register Form ===== */
              <form onSubmit={handleRegister} className="space-y-5">
                <div>
                  <label htmlFor="reg-email" className="mb-2 block text-sm font-medium text-muted">
                    电子邮件
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-muted" />
                    <input
                      id="reg-email"
                      type="email"
                      value={regEmail}
                      onChange={(e) => {
                        setRegEmail(e.target.value)
                        clearError()
                      }}
                      placeholder="请输入您的邮箱"
                      className="w-full rounded-md border border-white/10 bg-panel-light py-3 pl-10 pr-4 text-foreground placeholder:text-muted transition-colors focus:border-gold/50 focus:outline-none focus:ring-2 focus:ring-gold/50"
                      required
                    />
                  </div>
                </div>

                <div>
                  <label htmlFor="reg-password" className="mb-2 block text-sm font-medium text-muted">
                    密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-muted" />
                    <input
                      id="reg-password"
                      type={showRegPassword ? "text" : "password"}
                      value={regPassword}
                      onChange={(e) => {
                        setRegPassword(e.target.value)
                        clearError()
                      }}
                      placeholder="请设置密码"
                      className="w-full rounded-md border border-white/10 bg-panel-light py-3 pl-10 pr-12 text-foreground placeholder:text-muted transition-colors focus:border-gold/50 focus:outline-none focus:ring-2 focus:ring-gold/50"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowRegPassword(!showRegPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted transition-colors hover:text-foreground"
                    >
                      {showRegPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                    </button>
                  </div>
                </div>

                <div>
                  <label htmlFor="reg-confirm-password" className="mb-2 block text-sm font-medium text-muted">
                    确认密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-muted" />
                    <input
                      id="reg-confirm-password"
                      type={showRegConfirmPassword ? "text" : "password"}
                      value={regConfirmPassword}
                      onChange={(e) => {
                        setRegConfirmPassword(e.target.value)
                        clearError()
                      }}
                      placeholder="请再次输入密码"
                      className="w-full rounded-md border border-white/10 bg-panel-light py-3 pl-10 pr-12 text-foreground placeholder:text-muted transition-colors focus:border-gold/50 focus:outline-none focus:ring-2 focus:ring-gold/50"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowRegConfirmPassword(!showRegConfirmPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted transition-colors hover:text-foreground"
                    >
                      {showRegConfirmPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                    </button>
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={isRegLoading}
                  className="flex w-full items-center justify-center gap-2 rounded-md bg-gradient-to-r from-gold to-gold/80 py-3 font-semibold text-background transition-all hover:from-gold/90 hover:to-gold/70 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {isRegLoading ? (
                    <>
                      <div className="h-5 w-5 animate-spin rounded-full border-2 border-background/30 border-t-background" />
                      注册中...
                    </>
                  ) : (
                    "注册"
                  )}
                </button>
              </form>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

/**
 * 将 Nakama 错误信息转为用户友好的中文提示
 */
function errorToMessage(error: string): string {
  if (error.includes("Network") || error.includes("fetch") || error.includes("connect")) {
    return "无法连接服务器，请检查网络"
  }
  if (error.includes("已注册")) {
    return error // 使用 hook 层返回的中文提示，如"该邮箱已注册，请直接登录"
  }
  if (error.includes("credentials") || error.includes("password") || error.includes("email") || error.includes("invalid")) {
    return "邮箱或密码错误"
  }
  return "登录失败，请稍后重试"
}

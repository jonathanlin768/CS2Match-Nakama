import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Link } from "react-router-dom";
import { Mail, Lock, Eye, EyeOff, Gamepad2, AlertCircle } from "lucide-react";
import { useAuth } from "../context/AuthContext";

type Tab = "login" | "register";

export default function LoginPage() {
  const navigate = useNavigate();
  const { status, login, register, error } = useAuth();

  const [activeTab, setActiveTab] = useState<Tab>("login");

  // Login form state
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  // Register form state
  const [regEmail, setRegEmail] = useState("");
  const [regPassword, setRegPassword] = useState("");
  const [regConfirmPassword, setRegConfirmPassword] = useState("");
  const [showRegPassword, setShowRegPassword] = useState(false);
  const [showRegConfirmPassword, setShowRegConfirmPassword] = useState(false);
  const [isRegLoading, setIsRegLoading] = useState(false);

  // 恢复成功 → 自动跳转
  useEffect(() => {
    if (status === "authenticated") {
      navigate("/home", { replace: true });
    }
  }, [status, navigate]);

  // 收到服务端错误 → 展示
  useEffect(() => {
    if (error) {
      setFormError(errorToMessage(error));
    }
  }, [error]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setFormError(null);

    try {
      await login(email, password);
      // 成功后 status 变为 "authenticated"，上面的 useEffect 会跳转
    } catch {
      // 错误由 error 状态管理，上面的 useEffect 会设置 formError
    } finally {
      setIsLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsRegLoading(true);
    setFormError(null);

    // 客户端校验：两次密码一致
    if (regPassword !== regConfirmPassword) {
      setFormError("两次输入的密码不一致");
      setIsRegLoading(false);
      return;
    }

    // 客户端校验：密码长度 >= 8
    if (regPassword.length < 8) {
      setFormError("密码长度至少需要8位");
      setIsRegLoading(false);
      return;
    }

    try {
      await register(regEmail, regPassword);
      // 成功后 status 变为 "authenticated"，上面的 useEffect 会跳转
    } catch {
      // 错误由 error 状态管理，上面的 useEffect 会设置 formError
    } finally {
      setIsRegLoading(false);
    }
  };

  // 任何输入变化时清除错误
  const clearError = () => {
    if (formError) setFormError(null);
  };

  // 恢复中 → 加载画面
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

  // guest → 登录/注册表单
  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Background with gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-br from-background via-background to-primary/10" />
      <div className="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg%20width%3D%2260%22%20height%3D%2260%22%20viewBox%3D%220%200%2060%2060%22%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%3E%3Cg%20fill%3D%22none%22%20fill-rule%3D%22evenodd%22%3E%3Cg%20fill%3D%22%23c9a227%22%20fill-opacity%3D%220.03%22%3E%3Cpath%20d%3D%22M36%2034v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6%2034v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6%204V0H4v4H0v2h4v4h2V6h4V4H6z%22%2F%3E%3C%2Fg%3E%3C%2Fg%3E%3C%2Fsvg%3E')] opacity-50" />

      <div className="flex-1 flex items-center justify-center p-4 relative z-10">
        <div className="w-full max-w-md">
          {/* Logo */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center gap-3 mb-4">
              <div className="w-14 h-14 rounded-md bg-gradient-to-br from-primary to-primary/70 flex items-center justify-center">
                <Gamepad2 className="w-8 h-8 text-primary-foreground" />
              </div>
              <div className="text-left">
                <div className="text-3xl font-bold text-primary">CS2</div>
                <div className="text-xs text-muted-foreground tracking-widest">SIMULATOR</div>
              </div>
            </div>
            <p className="text-muted-foreground">
              {activeTab === "login" ? "登录您的账户，开始电竞之旅" : "创建新账户，加入电竞世界"}
            </p>
          </div>

          {/* Card */}
          <div className="bg-card rounded-md border border-border p-6 lg:p-8">
            {/* Tabs */}
            <div className="flex mb-6 border-b border-border">
              <button
                type="button"
                onClick={() => {
                  setActiveTab("login");
                  clearError();
                }}
                className={`flex-1 pb-3 text-sm font-medium text-center transition-colors relative ${
                  activeTab === "login"
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                登录
                {activeTab === "login" && (
                  <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t" />
                )}
              </button>
              <button
                type="button"
                onClick={() => {
                  setActiveTab("register");
                  clearError();
                }}
                className={`flex-1 pb-3 text-sm font-medium text-center transition-colors relative ${
                  activeTab === "register"
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                注册
                {activeTab === "register" && (
                  <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t" />
                )}
              </button>
            </div>

            {/* Error Message */}
            {formError && (
              <div className="mb-4 p-3 rounded-md bg-red-500/10 border border-red-500/30 flex items-start gap-2">
                <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
                <p className="text-sm text-red-400">{formError}</p>
              </div>
            )}

            {activeTab === "login" ? (
              /* ===== Login Form ===== */
              <form onSubmit={handleLogin} className="space-y-5">
                {/* Email Field */}
                <div>
                  <label htmlFor="email" className="block text-sm font-medium mb-2 text-muted-foreground">
                    电子邮件
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                    <input
                      id="email"
                      type="email"
                      value={email}
                      onChange={(e) => {
                        setEmail(e.target.value);
                        clearError();
                      }}
                      placeholder="请输入您的邮箱"
                      className="w-full pl-10 pr-4 py-3 bg-secondary border border-border rounded-md text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                      required
                    />
                  </div>
                </div>

                {/* Password Field */}
                <div>
                  <label htmlFor="password" className="block text-sm font-medium mb-2 text-muted-foreground">
                    密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                    <input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      value={password}
                      onChange={(e) => {
                        setPassword(e.target.value);
                        clearError();
                      }}
                      placeholder="请输入您的密码"
                      className="w-full pl-10 pr-12 py-3 bg-secondary border border-border rounded-md text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {showPassword ? (
                        <EyeOff className="w-5 h-5" />
                      ) : (
                        <Eye className="w-5 h-5" />
                      )}
                    </button>
                  </div>
                </div>

                {/* Remember Me & Forgot Password */}
                <div className="flex items-center justify-between">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={false}
                      readOnly
                      className="w-4 h-4 rounded border-border bg-secondary text-primary focus:ring-primary/50"
                    />
                    <span className="text-sm text-muted-foreground">记住我</span>
                  </label>
                  <Link to="#" className="text-sm text-primary hover:underline">
                    忘记密码？
                  </Link>
                </div>

                {/* Login Button */}
                <button
                  type="submit"
                  disabled={isLoading}
                  className="w-full py-3 bg-gradient-to-r from-primary to-primary/80 hover:from-primary/90 hover:to-primary/70 text-primary-foreground font-semibold rounded-md transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {isLoading ? (
                    <>
                      <div className="w-5 h-5 border-2 border-primary-foreground/30 border-t-primary-foreground rounded-full animate-spin" />
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
                {/* Email Field */}
                <div>
                  <label htmlFor="reg-email" className="block text-sm font-medium mb-2 text-muted-foreground">
                    电子邮件
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                    <input
                      id="reg-email"
                      type="email"
                      value={regEmail}
                      onChange={(e) => {
                        setRegEmail(e.target.value);
                        clearError();
                      }}
                      placeholder="请输入您的邮箱"
                      className="w-full pl-10 pr-4 py-3 bg-secondary border border-border rounded-md text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                      required
                    />
                  </div>
                </div>

                {/* Password Field */}
                <div>
                  <label htmlFor="reg-password" className="block text-sm font-medium mb-2 text-muted-foreground">
                    密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                    <input
                      id="reg-password"
                      type={showRegPassword ? "text" : "password"}
                      value={regPassword}
                      onChange={(e) => {
                        setRegPassword(e.target.value);
                        clearError();
                      }}
                      placeholder="请设置密码"
                      className="w-full pl-10 pr-12 py-3 bg-secondary border border-border rounded-md text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowRegPassword(!showRegPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {showRegPassword ? (
                        <EyeOff className="w-5 h-5" />
                      ) : (
                        <Eye className="w-5 h-5" />
                      )}
                    </button>
                  </div>
                </div>

                {/* Confirm Password Field */}
                <div>
                  <label htmlFor="reg-confirm-password" className="block text-sm font-medium mb-2 text-muted-foreground">
                    确认密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                    <input
                      id="reg-confirm-password"
                      type={showRegConfirmPassword ? "text" : "password"}
                      value={regConfirmPassword}
                      onChange={(e) => {
                        setRegConfirmPassword(e.target.value);
                        clearError();
                      }}
                      placeholder="请再次输入密码"
                      className="w-full pl-10 pr-12 py-3 bg-secondary border border-border rounded-md text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowRegConfirmPassword(!showRegConfirmPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {showRegConfirmPassword ? (
                        <EyeOff className="w-5 h-5" />
                      ) : (
                        <Eye className="w-5 h-5" />
                      )}
                    </button>
                  </div>
                </div>

                {/* Register Button */}
                <button
                  type="submit"
                  disabled={isRegLoading}
                  className="w-full py-3 bg-gradient-to-r from-primary to-primary/80 hover:from-primary/90 hover:to-primary/70 text-primary-foreground font-semibold rounded-md transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {isRegLoading ? (
                    <>
                      <div className="w-5 h-5 border-2 border-primary-foreground/30 border-t-primary-foreground rounded-full animate-spin" />
                      注册中...
                    </>
                  ) : (
                    "注册"
                  )}
                </button>
              </form>
            )}

            {/* Divider */}
            <div className="relative my-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-border" />
              </div>
              <div className="relative flex justify-center text-xs">
                <span className="px-2 bg-card text-muted-foreground">或</span>
              </div>
            </div>

            {/* Social Login */}
            <div className="grid grid-cols-2 gap-3">
              <button className="flex items-center justify-center gap-2 py-2.5 px-4 border border-border rounded-md hover:bg-secondary transition-colors">
                <svg className="w-5 h-5" viewBox="0 0 24 24">
                  <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                  <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                  <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                  <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                </svg>
                <span className="text-sm">Google</span>
              </button>
              <button className="flex items-center justify-center gap-2 py-2.5 px-4 border border-border rounded-md hover:bg-secondary transition-colors">
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.166 6.839 9.489.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.604-3.369-1.341-3.369-1.341-.454-1.155-1.11-1.462-1.11-1.462-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.163 22 16.418 22 12c0-5.523-4.477-10-10-10z"/>
                </svg>
                <span className="text-sm">GitHub</span>
              </button>
            </div>
          </div>

          {/* Footer */}
          <p className="mt-6 text-center text-xs text-muted-foreground">
            {activeTab === "login" ? "登录" : "注册"}即表示您同意我们的{" "}
            <Link to="#" className="text-primary hover:underline">服务条款</Link>
            {" "}和{" "}
            <Link to="#" className="text-primary hover:underline">隐私政策</Link>
          </p>
        </div>
      </div>
    </div>
  );
}

/**
 * 将 Nakama 错误信息转为用户友好的中文提示
 */
function errorToMessage(error: string): string {
  if (error.includes("Network") || error.includes("fetch") || error.includes("connect")) {
    return "无法连接服务器，请检查网络";
  }
  if (error.includes("已注册")) {
    return error; // 使用 hook 层返回的中文提示，如"该邮箱已注册，请直接登录"
  }
  if (error.includes("credentials") || error.includes("password") || error.includes("email") || error.includes("invalid")) {
    return "邮箱或密码错误";
  }
  return "登录失败，请稍后重试";
}

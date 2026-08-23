const socialMessages: Record<string, string> = {
  PROFILE_INCOMPLETE: "请先保存至少一种联系方式；QQ 或微信任选一项即可，不需要两项都填写。",
  PROFILE_CHANGED: "发起者的联系方式已经变化，请重新发起申请。",
  NOT_FRIENDS: "你们已不是好友，无法交换联系方式。",
  CONFLICT: "状态已发生变化，请刷新后重试。",
  INVALID_STATE: "这条申请已经处理或失效。",
  RATE_LIMITED: "操作过于频繁，请稍后再试。",
  PENDING_EXISTS: "你们之间已有一条等待处理的申请。",
  ALREADY_AUTHORIZED: "你们已经完成联系方式授权。",
  FORMAL_ACCOUNT_REQUIRED: "请先登录正式账号再保存联系方式。",
  INVALID_QQ: "QQ 号格式不正确，请输入 5–12 位数字。",
  INVALID_WECHAT: "微信号格式不正确。",
  INVALID_CHANNEL: "请选择有效的联系方式渠道。",
  FORBIDDEN: "你无权处理这条申请。",
  EXPIRED: "这条申请已经过期，请重新发起。",
  REVOCATION_FAILED: "未能安全撤销联系方式授权，好友删除已取消。",
  INVALID_REQUEST: "请求格式不正确，请刷新后重试。",
}

export class SocialApiError extends Error {
  readonly code: string
  readonly status?: number

  constructor(code: string, message?: string, status?: number) {
    super(message ?? socialMessages[code] ?? "联系方式服务暂时不可用，请稍后重试。")
    this.name = "SocialApiError"
    this.code = code
    this.status = status
  }
}

function codeFromMessage(message: string) {
  return /^([A-Z][A-Z0-9_]+):/.exec(message.trim())?.[1]
}

function safeMessage(code: string, fallback?: string) {
  return socialMessages[code] ?? (fallback && /^(NetworkError|Failed to fetch|Load failed|offline)$/i.test(fallback.trim())
    ? "网络连接失败，请检查连接后重试。"
    : "联系方式服务暂时不可用，请稍后重试。")
}

export async function toSocialApiError(error: unknown): Promise<SocialApiError> {
  if (error instanceof SocialApiError) return error
  if (typeof Response !== "undefined" && error instanceof Response) {
    let body: unknown
    try { body = await error.clone().json() } catch { body = undefined }
    const record = body && typeof body === "object" ? body as Record<string, unknown> : {}
    const rawMessage = typeof record.message === "string" ? record.message : ""
    const rawCode = typeof record.code === "string" ? record.code : codeFromMessage(rawMessage)
    const code = rawCode || `HTTP_${error.status || 0}`
    return new SocialApiError(code, safeMessage(code, rawMessage), error.status)
  }
  if (error && typeof error === "object") {
    const record = error as Record<string, unknown>
    const rawMessage = typeof record.message === "string" ? record.message : ""
    const rawCode = typeof record.code === "string" ? record.code : codeFromMessage(rawMessage)
    const code = rawCode || "NETWORK_ERROR"
    return new SocialApiError(code, safeMessage(code, rawMessage))
  }
  return new SocialApiError("UNKNOWN_ERROR")
}

export function socialErrorMessage(error: unknown) {
  if (error instanceof SocialApiError) return error.message
  return "联系方式服务暂时不可用，请稍后重试。"
}

export function decodeSocialRpcPayload<T>(payload: unknown): T {
  return (typeof payload === "string" ? JSON.parse(payload) : payload) as T
}

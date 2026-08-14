export async function rpcErrorMessage(reason: unknown): Promise<string> {
  if (reason instanceof Response) {
    try {
      const body = await reason.clone().json() as { message?: string; error?: string | { code?: string; message?: string } }
      if (typeof body.error === "object" && body.error) {
        const nestedMessage = [body.error.code, body.error.message].filter(Boolean).join(": ")
        if (nestedMessage) return nestedMessage
      }
      return body.message || (typeof body.error === "string" ? body.error : "") || `HTTP ${reason.status}`
    } catch {
      return `HTTP ${reason.status} ${reason.statusText}`.trim()
    }
  }
  return reason instanceof Error ? reason.message : String(reason)
}

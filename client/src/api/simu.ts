import type { Session } from "@heroiclabs/nakama-js"
import client from "../nakama"
import type { MatchReport } from "../types/match-report"

/**
 * 调用 DebugSimuMatch RPC，执行一回合简化模拟对战。
 *
 * Nakama RPC 返回的 payload 是 JSON 字符串，需要手动解析。
 */
export async function debugSimuMatch(session: Session, mapId = "de_dust2"): Promise<MatchReport> {
  const rpcRes = await client.rpc(session, "DebugSimuMatch", { map_id: mapId })
  const payload = typeof rpcRes.payload === "string" ? rpcRes.payload : JSON.stringify(rpcRes.payload)
  return JSON.parse(payload) as MatchReport
}

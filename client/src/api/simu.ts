import type { Session } from "@heroiclabs/nakama-js"
import client from "../nakama"
import type { MatchReport } from "../types/match-report"
import { rpcErrorMessage } from "./rpc-error"

/**
 * 调用 DebugSimuMatch RPC，执行完整模拟对战。
 *
 * Nakama RPC 返回的 payload 是 JSON 字符串，需要手动解析。
 */
export async function debugSimuMatch(session: Session, mapId = "de_dust2", seed?: number): Promise<MatchReport> {
  return callMatchRpc(session, "DebugSimuMatch", { map_id: mapId, ...(seed ? { seed } : {}) })
}

export type SimuMatchRequest =
  | { mode: "computer" }
  | { mode: "tutorial"; tutorial_config_id: string; config_version: number; player_ids: string[] }

export async function simuMatch(session: Session, request: SimuMatchRequest): Promise<MatchReport> {
  return callMatchRpc(session, "SimuMatch", request)
}

async function callMatchRpc(session: Session, id: string, input: object): Promise<MatchReport> {
  try {
    const response = await client.rpc(session, id, input)
    if (typeof response.payload === "string") return JSON.parse(response.payload) as MatchReport
    return response.payload as MatchReport
  } catch (reason) {
    throw new Error(await rpcErrorMessage(reason), { cause: reason })
  }
}

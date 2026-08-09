import assert from "node:assert/strict"
import test from "node:test"
import { authoritativeScoreAtPlayback, latestBombAtPlayback, previousTeamScore } from "./battle-playback.ts"

const report = {
  rounds: [
    { score_team_a: 1, score_team_b: 0, score_t: 99, score_ct: 98 },
    { score_team_a: 1, score_team_b: 1, score_t: 77, score_ct: 76 },
  ],
}

test("playback uses team identity scores and never side projection totals", () => {
  assert.deepEqual(previousTeamScore(report, 1), { teamA: 1, teamB: 0 })
  assert.deepEqual(authoritativeScoreAtPlayback(report, report.rounds[1], 1, []), { teamA: 1, teamB: 0 })
  assert.deepEqual(
    authoritativeScoreAtPlayback(report, report.rounds[1], 1, [{ event_type: "ROUND_END" }]),
    { teamA: 1, teamB: 1 },
  )
})

test("playback uses the latest authoritative bomb snapshot", () => {
  const round = { bomb: { status: "Carried", node_id: "T_SPAWN" } }
  const events = [
    { event_type: "BOMB_PLANT_START", bomb: { status: "Carried", node_id: "A_SITE" } },
    { event_type: "BOMB_PLANT", bomb: { status: "Planted", node_id: "A_SITE", site: "A" } },
  ]
  assert.deepEqual(latestBombAtPlayback(round, events), events[1].bomb)
})

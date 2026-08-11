import assert from "node:assert/strict"
import test from "node:test"
import {
  authoritativeScoreAtPlayback,
  cumulativePlayerStatsAtPlayback,
  formatBattleEvent,
  latestBombAtPlayback,
  playerVitalsAtPlayback,
  previousTeamScore,
  selectedRoundInitialEventCount,
} from "./battle-playback.ts"

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

test("round selection resets playback to the round start event", () => {
	assert.equal(selectedRoundInitialEventCount({ events: [{ event_type: "ROUND_START" }, { event_type: "KILL" }] }), 1)
	assert.equal(selectedRoundInitialEventCount({ events: [{ event_type: "MATCH_START" }, { event_type: "ROUND_START" }, { event_type: "KILL" }] }), 2)
	assert.equal(selectedRoundInitialEventCount({ events: [] }), 0)
})

test("playback KDA accumulates completed rounds and visible current events", () => {
  const match = {
    rounds: [
      { events: [{ event_type: "KILL", attacker_id: "a", victim_id: "b", extra: { assist_ids: ["c"] } }] },
      { events: [] },
    ],
  }
  const stats = cumulativePlayerStatsAtPlayback(match, 1, [
    { event_type: "KILL", attacker_id: "a", victim_id: "c", extra: { assist_ids: ["b"] } },
  ])
  assert.deepEqual(stats, {
    a: { kills: 2, deaths: 0, assists: 0 },
    b: { kills: 0, deaths: 1, assists: 1 },
    c: { kills: 0, deaths: 1, assists: 1 },
  })
})

test("playback health follows visible damage events instead of stale snapshots or final round state", () => {
  const round = {
    player_states: [
      { player_id: "attacker", alive: true, is_alive: true, hp: 100 },
      { player_id: "victim", alive: false, is_alive: false, hp: 0 },
    ],
  }

  assert.deepEqual(playerVitalsAtPlayback(round, []), {
    attacker: { alive: true, hp: 100 },
    victim: { alive: true, hp: 100 },
  })
  assert.deepEqual(
    playerVitalsAtPlayback(round, [
      { event_type: "DAMAGE", victim_id: "victim", extra: { damage: 59 } },
    ]).victim,
    { alive: true, hp: 41 },
  )
  assert.deepEqual(
    playerVitalsAtPlayback(round, [
      {
        event_type: "DAMAGE",
        victim_id: "victim",
        extra: { damage: 59 },
        state: { players: [{ player_id: "victim", alive: true, is_alive: true, hp: 100 }] },
      },
      { event_type: "KILL", victim_id: "victim" },
    ]).victim,
    { alive: false, hp: 0 },
  )
})

test("structured events are rendered as localized battle commentary", () => {
  const context = {
    teamAID: "a",
    teamAName: "TeamA",
    teamBID: "b",
    teamBName: "TeamB",
    teamTID: "a",
    teamCTID: "b",
    winnerTeamID: "a",
    winReason: "elimination",
    strategyTemplateID: "A_Long_Rush",
    ctSetupTemplateID: "CT_Default",
  }
  assert.equal(
    formatBattleEvent({ event_type: "DAMAGE", attacker_name: "选手A", victim_name: "选手B", location: { name: "A大" }, extra: { damage: 38 }, message: "damage applied" }, context),
    "选手A 在 A大 攻击了 选手B，造成 38 点伤害。",
  )
	assert.equal(
		formatBattleEvent({ event_type: "MATCH_START", message: "match started" }, context),
		"TeamA 对阵 TeamB 的比赛开始。",
	)
	assert.equal(
		formatBattleEvent({ event_type: "ROUND_START", message: "round started" }, context),
    "TeamA 作为 T 执行 A_Long_Rush 战术；TeamB 作为 CT 使用 CT_Default 防守。",
  )
})

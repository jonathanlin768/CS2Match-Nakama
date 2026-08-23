import type { MatchReport, PlayerState, RoundReport, Side } from "../../src/types/match-report"

const teamAID = "tutorial_players"
const teamBID = "team_vitality"

function player(playerID: string, displayName: string, teamID: string, side: Side): PlayerState {
  return {
    player_id: `${teamID}/${playerID}`,
    config_player_id: playerID,
    player_name: displayName,
    display_name: displayName,
    team_id: teamID,
    side,
    is_alive: true,
    alive: true,
    hp: 100,
    stamina: 100,
    focus: 100,
    current_node: "MID",
    kills: 0,
    deaths: 0,
    damage: 0,
    weapon: {
      primary: side === "T" ? "AK47" : "M4A1S",
      secondary: "Glock",
      armor: true,
      helmet: true,
      has_kit: side === "CT",
    },
  }
}

function players(teamASide: Side): PlayerState[] {
  const teamBSide: Side = teamASide === "T" ? "CT" : "T"
  return [
    ...["donk", "kyousuke", "karrigan", "NiKo", "TeSeS"].map((name) => player(`player_${name.toLowerCase()}`, name, teamAID, teamASide)),
    ...["apex", "flameZ", "mezii", "ropz", "ZywOo"].map((name) => player(`player_${name.toLowerCase()}`, name, teamBID, teamBSide)),
  ]
}

function roundOne(): RoundReport {
  return {
    round_number: 1, phase: "regulation", half: 1, seed: 1001, side_attacking: "T",
    team_t_id: teamAID, team_ct_id: teamBID, winner: "CT", winner_team_id: teamBID,
    win_reason: "elimination", score_team_a: 0, score_team_b: 1, score_t: 0, score_ct: 1,
    route_main: "A_LONG", strategy_template_id: "A_Long_Rush", ct_setup_template_id: "CT_Default",
    player_states: players("T"), bomb: { status: "Carried", carrier_id: `${teamAID}/player_donk`, node_id: "T_SPAWN" },
    events: [
      { timestamp: 0, event_type: "MATCH_START", message: "match started" },
      { timestamp: 1, event_type: "ROUND_START", message: "round started" },
      { timestamp: 25, event_type: "DAMAGE", attacker_id: `${teamBID}/player_apex`, attacker_name: "apex", attacker_team_id: teamBID, victim_id: `${teamAID}/player_donk`, victim_name: "donk", victim_team_id: teamAID, location: { name: "A Long", x: .32, y: .28 }, extra: { damage: 52 }, message: "damage" },
      { timestamp: 26, event_type: "KILL", attacker_id: `${teamBID}/player_apex`, attacker_name: "apex", attacker_team_id: teamBID, victim_id: `${teamAID}/player_donk`, victim_name: "donk", victim_team_id: teamAID, weapon: "M4A1-S", location: { name: "A Long", x: .32, y: .28 }, message: "kill" },
      { timestamp: 115, event_type: "ROUND_END", message: "round end" },
    ],
  }
}

function roundTwo(): RoundReport {
  return {
    round_number: 2, phase: "regulation", half: 1, seed: 1002, side_attacking: "T",
    team_t_id: teamAID, team_ct_id: teamBID, winner: "T", winner_team_id: teamAID,
    win_reason: "bomb_exploded", score_team_a: 1, score_team_b: 1, score_t: 1, score_ct: 1,
    route_main: "B_TUNNEL", strategy_template_id: "B_Execute", ct_setup_template_id: "CT_Default",
    player_states: players("T"), bomb: { status: "Exploded", node_id: "B_SITE", site: "B" },
    events: [
      { timestamp: 0, event_type: "ROUND_START", message: "round started" },
      { timestamp: 34, event_type: "KILL", attacker_id: `${teamAID}/player_niko`, attacker_name: "NiKo", attacker_team_id: teamAID, victim_id: `${teamBID}/player_mezii`, victim_name: "mezii", victim_team_id: teamBID, weapon: "AK-47", location: { name: "Mid", x: .5, y: .5 }, message: "kill" },
      { timestamp: 58, event_type: "BOMB_PLANT", attacker_id: `${teamAID}/player_karrigan`, attacker_name: "karrigan", attacker_team_id: teamAID, bomb: { status: "Planted", node_id: "B_SITE", site: "B" }, location: { name: "B Site", x: .75, y: .25 }, message: "plant" },
      { timestamp: 99, event_type: "BOMB_EXPLODE", bomb: { status: "Exploded", node_id: "B_SITE", site: "B" }, location: { name: "B Site", x: .75, y: .25 }, message: "explode" },
      { timestamp: 100, event_type: "ROUND_END", message: "round end" },
      { timestamp: 101, event_type: "MATCH_END", message: "match end" },
    ],
  }
}

const allPlayers = players("T")

export const battleReportFixture: MatchReport = {
  debug_enabled: false,
  match_info: {
    match_id: "e2e-mobile-battle", map_id: "de_dust2", map_name: "Dust2", map_version: "1",
    rule_set_id: "mr12_v1", seed: 1000, team_a_id: teamAID, team_b_id: teamBID,
    team_a_name: "你的临时阵容", team_b_name: "Team Vitality", start_time: 0, total_rounds: 2,
    final_score_team_a: 1, final_score_team_b: 1, winner_team_id: teamAID,
  },
  rounds: [roundOne(), roundTwo()],
  final_stats: {
    score_t: 1, score_ct: 1, score_team_a: 1, score_team_b: 1, winner_team_id: teamAID,
    player_stats: allPlayers.map((item) => ({
      player_id: item.player_id, config_player_id: item.config_player_id, player_name: item.player_name,
      team_id: item.team_id, side: item.side, kills: 1, deaths: 1, assists: 0, damage: 100,
      adr: 50, fk: 0, mk: 0, plants: 0, defuses: 0,
    })),
  },
  winner: "T", winner_team_id: teamAID, final_score_team_a: 1, final_score_team_b: 1, total_rounds: 2,
}

import type { BattleTeam } from "./data/battle"

function StatsTable({ team }: { team: BattleTeam }) {
  return (
    <section className="battle-stats-table">
      <h3>{team.name}</h3>
      <table>
        <thead><tr><th>选手</th><th>K</th><th>D</th><th>A</th></tr></thead>
        <tbody>
          {team.players.map((player) => (
            <tr key={player.instanceId ?? player.id} className={player.alive ? "" : "is-dead"}>
              <td>{player.id}</td><td>{player.kills}</td><td>{player.deaths}</td><td>{player.assists}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}

export default function BattleStatsTables({ teamA, teamB }: { teamA: BattleTeam; teamB: BattleTeam }) {
  return <div className="battle-stats-tables"><StatsTable team={teamA} /><StatsTable team={teamB} /></div>
}

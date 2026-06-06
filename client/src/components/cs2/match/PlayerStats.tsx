interface PlayerStat {
  name: string;
  team: "t" | "ct";
  kills: number;
  assists: number;
  deaths: number;
  kd: number;
  adr: number;
  hsPercent: number;
  kast: string;
  rating: number;
}

interface Props {
  team1Stats: PlayerStat[];
  team2Stats: PlayerStat[];
  team1Name: string;
  team2Name: string;
  team1Logo: string;
  team2Logo: string;
}

function StatRow({ stat, highlight }: { stat: PlayerStat; highlight?: boolean }) {
  return (
    <tr className={`text-xs ${highlight ? "bg-muted/60" : ""}`}>
      <td className={`px-2 py-1 font-medium ${stat.team === "t" ? "text-orange-400" : "text-blue-400"}`}>
        {stat.name}
      </td>
      <td className="px-2 py-1 text-center tabular-nums">{stat.kills}</td>
      <td className="px-2 py-1 text-center tabular-nums">{stat.assists}</td>
      <td className="px-2 py-1 text-center tabular-nums">{stat.deaths}</td>
      <td className="px-2 py-1 text-center tabular-nums">{stat.kd.toFixed(2)}</td>
      <td className="px-2 py-1 text-center tabular-nums">{stat.adr}</td>
      <td className="px-2 py-1 text-center tabular-nums">{stat.hsPercent}%</td>
      <td className="px-2 py-1 text-center tabular-nums">{stat.kast}</td>
      <td className="px-2 py-1 text-center tabular-nums font-bold">{stat.rating.toFixed(2)}</td>
    </tr>
  );
}

export function PlayerStats({ team1Stats, team2Stats, team1Name: _team1Name, team2Name: _team2Name, team1Logo: _team1Logo, team2Logo: _team2Logo }: Props) {
  return (
    <div className="bg-card border border-border rounded-lg p-3">
      <div className="font-bold text-sm mb-2">玩家数据</div>
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="text-xs text-muted-foreground border-b border-border">
              <th className="px-2 py-1">玩家</th>
              <th className="px-2 py-1 text-center">K</th>
              <th className="px-2 py-1 text-center">A</th>
              <th className="px-2 py-1 text-center">D</th>
              <th className="px-2 py-1 text-center">K/D</th>
              <th className="px-2 py-1 text-center">ADR</th>
              <th className="px-2 py-1 text-center">HS%</th>
              <th className="px-2 py-1 text-center">KAST</th>
              <th className="px-2 py-1 text-center">Rating</th>
            </tr>
          </thead>
          <tbody>
            {team1Stats.map((s, i) => (
              <StatRow key={`t-${i}`} stat={s} highlight={i % 2 === 1} />
            ))}
            {team2Stats.map((s, i) => (
              <StatRow key={`ct-${i}`} stat={s} highlight={i % 2 === 1} />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

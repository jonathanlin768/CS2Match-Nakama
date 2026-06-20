interface RosterPlayer {
  id: string;
  name: string;
  avatar: string;
  weapons: string[];
  money: number;
  health: number;
  kills: number;
  deaths: number;
}

interface Props {
  teamName: string;
  teamLogo: string;
  teamMoney: number;
  players: RosterPlayer[];
  side: "left" | "right";
  teamColor: "orange" | "blue";
}

export function TeamRoster({ teamName, teamLogo, teamMoney, players, side, teamColor }: Props) {
  const colorClass = teamColor === "orange" ? "text-orange-500" : "text-blue-500";
  const barClass = teamColor === "orange" ? "bg-orange-500" : "bg-blue-500";

  return (
    <div className="bg-card border border-border rounded-lg p-3 h-full">
      <div className={`flex items-center gap-2 mb-3 ${side === "right" ? "flex-row-reverse" : ""}`}>
        <span className="text-xl">{teamLogo}</span>
        <div className={side === "right" ? "text-right" : ""}>
          <div className={`font-bold text-sm ${colorClass}`}>{teamName}</div>
          <div className="text-xs text-muted-foreground">${teamMoney.toLocaleString()}</div>
        </div>
      </div>

      <div className="space-y-2">
        {players.map((p) => (
          <div
            key={p.id}
            className={`flex items-center gap-2 p-2 rounded bg-muted/50 ${side === "right" ? "flex-row-reverse text-right" : ""}`}
          >
            <div className="text-lg">{p.avatar}</div>
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate">{p.name}</div>
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                <span>{p.weapons.join(" ")}</span>
              </div>
              <div className="mt-1 h-1.5 w-full bg-muted rounded-full overflow-hidden">
                <div
                  className={`h-full ${barClass} transition-all`}
                  style={{ width: `${p.health}%` }}
                />
              </div>
            </div>
            <div className="text-xs font-mono tabular-nums">
              {p.kills}/{p.deaths}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

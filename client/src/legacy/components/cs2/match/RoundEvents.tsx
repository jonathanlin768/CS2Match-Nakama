interface RoundEventItem {
  time: string;
  type: "kill" | "bomb" | "tactic" | "smoke";
  player: string;
  playerTeam: "t" | "ct";
  target: string;
  targetTeam: "t" | "ct";
  weapon: string;
  description: string;
}

interface Props {
  events: RoundEventItem[];
}

export function RoundEvents({ events }: Props) {
  if (events.length === 0) {
    return (
      <div className="bg-card border border-border rounded-lg p-3 h-full">
        <div className="font-bold text-sm mb-2">回合事件</div>
        <div className="text-sm text-muted-foreground text-center py-8">暂无事件</div>
      </div>
    );
  }

  return (
    <div className="bg-card border border-border rounded-lg p-3 h-full flex flex-col">
      <div className="font-bold text-sm mb-2">回合事件</div>
      <div className="flex-1 overflow-y-auto space-y-2">
        {events.map((e, i) => (
          <div key={i} className="text-xs p-2 rounded bg-muted/50">
            <div className="flex items-center gap-1 mb-1">
              <span className="font-mono text-muted-foreground">{e.time}</span>
              {e.type === "kill" && <span className="text-red-400 font-bold">击杀</span>}
              {e.type === "bomb" && <span className="text-orange-400 font-bold">💣 C4</span>}
            </div>
            <div className="leading-relaxed">
              <span className={e.playerTeam === "t" ? "text-orange-400 font-medium" : "text-blue-400 font-medium"}>
                {e.player}
              </span>
              {" "}{e.description}{" "}
              {e.target && (
                <span className={e.targetTeam === "t" ? "text-orange-400 font-medium" : "text-blue-400 font-medium"}>
                  {e.target}
                </span>
              )}
              {e.weapon && (
                <span className="text-muted-foreground ml-1">({e.weapon})</span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

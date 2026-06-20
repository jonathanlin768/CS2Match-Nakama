interface TacticItem {
  time: string;
  type: "tactic" | "kill" | "smoke";
  title: string;
  team: "t" | "ct";
  description: string;
}

interface Props {
  tactics: TacticItem[];
  currentRound: number;
}

export function RoundTactics({ tactics, currentRound: _currentRound }: Props) {
  return (
    <div className="bg-card border border-border rounded-lg p-3 h-full flex flex-col">
      <div className="font-bold text-sm mb-2">战术部署</div>
      {tactics.length === 0 ? (
        <div className="text-sm text-muted-foreground text-center py-8">暂无战术信息</div>
      ) : (
        <div className="flex-1 overflow-y-auto space-y-2">
          {tactics.map((t, i) => (
            <div key={i} className="text-xs p-2 rounded bg-muted/50">
              <div className="flex items-center gap-1 mb-1">
                <span className="font-mono text-muted-foreground">{t.time}</span>
                <span className={t.team === "t" ? "text-orange-400 font-bold" : "text-blue-400 font-bold"}>
                  {t.title}
                </span>
              </div>
              <div className="text-muted-foreground leading-relaxed">{t.description}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

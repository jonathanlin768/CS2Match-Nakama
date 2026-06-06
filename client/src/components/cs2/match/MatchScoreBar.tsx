interface TeamInfo {
  name: string;
  logo: string;
  score?: number;
}

interface Props {
  team1: TeamInfo;
  team2: TeamInfo;
  roundTime: string;
  currentRound: number;
  totalRounds: number;
}

export function MatchScoreBar({ team1, team2, roundTime, currentRound, totalRounds }: Props) {
  return (
    <div className="flex items-center justify-between px-4 py-3 bg-card border-b border-border">
      <div className="flex items-center gap-3">
        <span className="text-2xl">{team1.logo}</span>
        <div>
          <div className="font-bold text-sm">{team1.name}</div>
          <div className="text-xs text-muted-foreground">T 方</div>
        </div>
        <div className="text-2xl font-bold ml-2">{team1.score ?? 0}</div>
      </div>

      <div className="flex flex-col items-center">
        <div className="text-2xl font-mono font-bold tabular-nums">{roundTime}</div>
        <div className="text-xs text-muted-foreground">
          Round {currentRound} / {totalRounds}
        </div>
      </div>

      <div className="flex items-center gap-3">
        <div className="text-2xl font-bold mr-2">{team2.score ?? 0}</div>
        <div className="text-right">
          <div className="font-bold text-sm">{team2.name}</div>
          <div className="text-xs text-muted-foreground">CT 方</div>
        </div>
        <span className="text-2xl">{team2.logo}</span>
      </div>
    </div>
  );
}

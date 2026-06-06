"use client";

import { ChevronRight } from "lucide-react";

const matches = [
  {
    tournament: "IEM Dallas 2024",
    type: "BO3",
    team1: { name: "FaZe", logo: "🔴", score: 2 },
    team2: { name: "NaVi", logo: "🟡", score: 1 },
    time: "2小时前",
  },
  {
    tournament: "BLAST Premier Spring",
    type: "BO3",
    team1: { name: "Vitality", logo: "🟡", score: 2 },
    team2: { name: "G2", logo: "⚫", score: 0 },
    time: "5小时前",
  },
  {
    tournament: "PGL Major Copenhagen 2024",
    type: "BO3",
    team1: { name: "ENCE", logo: "🟠", score: 1 },
    team2: { name: "FaZe", logo: "🔴", score: 2 },
    time: "昨天",
  },
];

export function RecentMatches() {
  return (
    <div className="flex-1 bg-card rounded-md border border-border p-3 lg:p-4 flex flex-col">
      <div className="flex items-center justify-between mb-3 lg:mb-4">
        <h3 className="font-semibold text-sm lg:text-base">最近比赛</h3>
        <button className="flex items-center gap-1 text-[10px] lg:text-xs text-primary hover:underline">
          查看全部 <ChevronRight className="w-3 h-3" />
        </button>
      </div>

      <div className="space-y-2 flex-1 flex flex-col justify-between">
        {matches.map((match, index) => (
          <div
            key={index}
            className="p-2 lg:p-2.5 bg-secondary/50 rounded"
          >
            <div className="flex items-center justify-between mb-1 lg:mb-1.5">
              <div className="flex items-center gap-1.5 lg:gap-2 flex-1 min-w-0">
                <span className="text-[10px] lg:text-xs text-muted-foreground truncate">{match.tournament}</span>
                <span className="text-[8px] lg:text-[10px] px-1 lg:px-1.5 py-0.5 bg-muted rounded text-muted-foreground flex-shrink-0">
                  {match.type}
                </span>
              </div>
              <span className="text-[10px] lg:text-xs text-muted-foreground flex-shrink-0 ml-2">{match.time}</span>
            </div>
            
            <div className="flex items-center">
              <div className="flex-1 flex items-center gap-1.5 lg:gap-2 min-w-0">
                <span className="text-sm lg:text-base">{match.team1.logo}</span>
                <span className="text-xs lg:text-sm truncate">{match.team1.name}</span>
              </div>
              
              <div className="flex items-center justify-center gap-1 lg:gap-1.5 w-12 lg:w-16 flex-shrink-0">
                <span className={`text-sm lg:text-base font-bold ${match.team1.score > match.team2.score ? 'text-green-400' : 'text-muted-foreground'}`}>
                  {match.team1.score}
                </span>
                <span className="text-muted-foreground text-xs lg:text-sm">:</span>
                <span className={`text-sm lg:text-base font-bold ${match.team2.score > match.team1.score ? 'text-green-400' : 'text-muted-foreground'}`}>
                  {match.team2.score}
                </span>
              </div>

              <div className="flex-1 flex items-center justify-end gap-1.5 lg:gap-2 min-w-0">
                <span className="text-xs lg:text-sm truncate">{match.team2.name}</span>
                <span className="text-sm lg:text-base">{match.team2.logo}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

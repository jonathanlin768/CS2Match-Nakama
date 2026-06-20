"use client";

import { ChevronLeft, ChevronRight, Heart } from "lucide-react";

const tabs = ["热门战队", "热门选手", "热门卡组"];

const teams = [
  { rank: 1, name: "Natus Vincere", logo: "🟡", price: "$1,245,000", rating: 87 },
  { rank: 2, name: "FaZe Clan", logo: "🔴", price: "$1,128,000", rating: 85, liked: true },
  { rank: 3, name: "G2 Esports", logo: "⚫", price: "$982,000", rating: 84 },
  { rank: 4, name: "Vitality", logo: "🟡", price: "$872,000", rating: 83 },
  { rank: 5, name: "ENCE", logo: "🟠", price: "$745,000", rating: 82 },
];

export function TeamRankings() {
  return (
    <div className="flex-1 flex flex-col">
      <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4 lg:gap-6 mb-3 lg:mb-4">
        <h3 className="font-semibold text-sm lg:text-base">推荐内容</h3>
        <div className="flex items-center gap-3 lg:gap-4 overflow-x-auto">
          {tabs.map((tab, index) => (
            <button
              key={tab}
              className={`text-xs lg:text-sm transition-colors whitespace-nowrap ${
                index === 0
                  ? "text-foreground font-medium"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      <div className="relative flex-1 flex flex-col">
        <div className="flex items-stretch gap-2 lg:gap-4 flex-1">
          <button className="hidden sm:block absolute -left-4 top-1/2 -translate-y-1/2 z-10 p-1.5 lg:p-2 bg-card border border-border rounded-full hover:bg-secondary transition-colors">
            <ChevronLeft className="w-3 lg:w-4 h-3 lg:h-4" />
          </button>

          <div className="flex gap-2 lg:gap-3 overflow-x-auto flex-1 pb-2 scrollbar-hide">
            {teams.map((team) => (
              <div
                key={team.rank}
                className="flex-shrink-0 w-24 sm:w-28 lg:w-32 bg-card rounded-md border border-border p-2 sm:p-3 lg:p-4 hover:border-primary/50 transition-colors flex flex-col"
              >
                <div className="flex items-center justify-between mb-2 lg:mb-4">
                  <span className="text-[10px] lg:text-xs text-muted-foreground font-medium">
                    {team.rank}
                  </span>
                  <button className={team.liked ? "text-red-500" : "text-muted-foreground"}>
                    <Heart className="w-3 lg:w-4 h-3 lg:h-4" fill={team.liked ? "currentColor" : "none"} />
                  </button>
                </div>

                <div className="flex-1 flex items-center justify-center">
                  <div className="w-10 sm:w-12 lg:w-16 h-10 sm:h-12 lg:h-16 rounded bg-secondary flex items-center justify-center text-xl sm:text-2xl lg:text-3xl">
                    {team.logo}
                  </div>
                </div>

                <div className="text-center mt-2 lg:mt-4">
                  <div className="text-[10px] sm:text-xs lg:text-sm font-medium truncate">{team.name}</div>
                  <div className="flex items-center justify-center gap-1 lg:gap-2 mt-1">
                    <span className="text-[10px] lg:text-xs text-green-400">{team.price}</span>
                    <span className="text-[10px] lg:text-xs text-muted-foreground">{team.rating}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <button className="hidden sm:block absolute -right-4 top-1/2 -translate-y-1/2 z-10 p-1.5 lg:p-2 bg-card border border-border rounded-full hover:bg-secondary transition-colors">
            <ChevronRight className="w-3 lg:w-4 h-3 lg:h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

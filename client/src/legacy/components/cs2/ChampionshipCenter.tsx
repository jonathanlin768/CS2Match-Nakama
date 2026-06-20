"use client";

import { Trophy } from "lucide-react";

export function ChampionshipCenter() {
  return (
    <div className="flex-1 bg-card rounded-md border border-border p-3 lg:p-4 flex flex-col">
      <div className="flex items-center gap-3 lg:gap-4 flex-1">
        <div className="flex-1">
          <h3 className="font-semibold text-sm lg:text-lg">锦标赛中心</h3>
          <p className="text-[10px] lg:text-xs text-muted-foreground mt-1">
            参加官方锦标赛赢取奖励
          </p>
          <button className="mt-2 lg:mt-3 px-3 lg:px-4 py-1 lg:py-1.5 border border-primary text-primary text-xs lg:text-sm rounded hover:bg-primary/10 transition-colors">
            查看赛事
          </button>
        </div>
        <div className="relative flex-shrink-0">
          <Trophy className="w-14 lg:w-20 h-14 lg:h-20 text-primary opacity-80" />
          <div className="absolute inset-0 bg-gradient-to-t from-primary/20 to-transparent rounded-full blur-xl" />
        </div>
      </div>
    </div>
  );
}

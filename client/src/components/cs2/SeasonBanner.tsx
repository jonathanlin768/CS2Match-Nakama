"use client";

import { ChevronRight } from "lucide-react";

export function SeasonBanner() {
  return (
    <div className="flex-1 relative bg-gradient-to-r from-[#1a2332] to-[#0f1520] rounded-md border border-border overflow-hidden">
      {/* Background Image Overlay */}
      <div className="absolute inset-0 bg-[url('https://hebbkx1anhila5yf.public.blob.vercel-storage.com/file_00000000ab607209b3b7459ba45900da-H81WxYDQ5A5LWqbOB8WMNv6LIb1stU.png')] bg-cover bg-center opacity-20" />
      
      <div className="relative p-4 lg:p-5 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex-1">
          <div className="flex items-center gap-2 lg:gap-3 mb-2 flex-wrap">
            <h2 className="text-lg sm:text-xl lg:text-2xl font-bold">
              模拟竞技赛季 <span className="text-primary">5</span>
            </h2>
            <span className="px-2 py-0.5 bg-green-600/20 text-green-400 text-xs rounded-full">
              进行中
            </span>
          </div>
          <p className="text-xs sm:text-sm text-muted-foreground mb-3 lg:mb-4">
            体验真实CS2赛场，运筹帷幄，决胜千里
          </p>
          
          <button className="px-4 lg:px-6 py-1.5 lg:py-2 bg-secondary hover:bg-secondary/80 rounded text-xs lg:text-sm font-medium transition-colors border border-border">
            进入赛季
          </button>

          {/* Season Progress */}
          <div className="mt-3 lg:mt-4 flex flex-wrap items-center gap-3 lg:gap-4">
            <div className="flex items-center gap-2">
              <div className="w-5 lg:w-6 h-5 lg:h-6 rounded bg-blue-600 flex items-center justify-center text-[10px] lg:text-xs font-bold">
                18
              </div>
              <span className="text-[10px] lg:text-xs text-muted-foreground">赛季等级 18</span>
            </div>
            <div className="flex-1 min-w-32 max-w-48">
              <div className="flex items-center justify-between text-[10px] lg:text-xs mb-1">
                <span className="text-muted-foreground">600 / 1000 XP</span>
              </div>
              <div className="h-1 lg:h-1.5 bg-secondary rounded-full overflow-hidden">
                <div className="h-full w-[60%] bg-blue-500 rounded-full" />
              </div>
            </div>
          </div>

          {/* Pagination dots */}
          <div className="flex items-center gap-1.5 mt-4">
            <div className="w-2 h-2 rounded-full bg-blue-500" />
            <div className="w-2 h-2 rounded-full bg-muted" />
            <div className="w-2 h-2 rounded-full bg-muted" />
            <div className="w-2 h-2 rounded-full bg-muted" />
          </div>
        </div>

        {/* Weapon Display - Hidden on very small screens */}
        <div className="hidden sm:flex flex-col items-end gap-2 flex-shrink-0">
          <div className="relative">
            <div className="w-40 lg:w-64 h-20 lg:h-32 bg-gradient-to-br from-purple-900/30 to-pink-900/30 rounded flex items-center justify-center border border-purple-500/30">
              <div className="text-2xl lg:text-4xl">🔫</div>
            </div>
          </div>
          <div className="text-right">
            <div className="text-[10px] lg:text-xs text-primary">赛季奖励</div>
            <div className="text-xs lg:text-sm font-medium">M4A1-S | 黑莲花</div>
          </div>
        </div>

        <button className="absolute right-4 top-1/2 -translate-y-1/2 p-2 hover:bg-secondary/50 rounded transition-colors">
          <ChevronRight className="w-5 h-5" />
        </button>
      </div>
    </div>
  );
}

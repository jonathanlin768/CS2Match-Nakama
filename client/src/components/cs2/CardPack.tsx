"use client";

import { ChevronRight, Gem } from "lucide-react";

export function CardPack() {
  return (
    <div className="bg-card rounded-md border border-border p-3 lg:p-4 flex-1">
      <div className="flex items-center justify-between mb-3 lg:mb-4">
        <h3 className="font-semibold text-sm lg:text-base">卡包推荐</h3>
        <button className="flex items-center gap-1 text-[10px] lg:text-xs text-primary hover:underline">
          查看全部 <ChevronRight className="w-3 h-3" />
        </button>
      </div>

      <div className="relative bg-gradient-to-br from-orange-900/30 to-amber-900/30 rounded-md p-3 lg:p-4 border border-orange-500/30">
        <div className="absolute top-2 right-2">
          <div className="w-14 lg:w-20 h-18 lg:h-24 bg-gradient-to-br from-orange-600 to-amber-700 rounded flex items-center justify-center shadow-lg transform rotate-6">
            <span className="text-2xl lg:text-3xl">🃏</span>
          </div>
        </div>

        <div className="pr-16 lg:pr-20">
          <h4 className="font-bold text-sm lg:text-lg">传奇之路卡包</h4>
          <p className="text-[10px] lg:text-xs text-muted-foreground mt-1">
            包含 5 张 80+ OVR 选手卡
          </p>
          <p className="text-[10px] lg:text-xs text-primary mt-0.5">
            2.00% 获得 90+ OVR
          </p>
        </div>

        <button className="mt-3 lg:mt-4 w-full py-1.5 lg:py-2 bg-gradient-to-r from-orange-600 to-amber-600 hover:from-orange-500 hover:to-amber-500 rounded text-xs lg:text-sm font-medium flex items-center justify-center gap-2 transition-colors">
          <Gem className="w-3.5 lg:w-4 h-3.5 lg:h-4" />
          <span>200</span>
        </button>
      </div>
    </div>
  );
}

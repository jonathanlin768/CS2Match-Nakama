"use client";

import { Swords, Map, History, Crown } from "lucide-react";

const gameModes = [
  {
    title: "模拟对战",
    subtitle: "选择战队，模拟比赛",
    icon: Swords,
    button: "开始对战",
    bgGradient: "from-slate-800 to-slate-900",
    isNew: false,
  },
  {
    title: "战术沙盘",
    subtitle: "自定义战术，模拟推演",
    icon: Map,
    button: "进入沙盘",
    bgGradient: "from-amber-900/50 to-slate-900",
    isNew: false,
  },
  {
    title: "经典战役",
    subtitle: "重现历史经典赛事",
    icon: History,
    button: "选择战役",
    bgGradient: "from-emerald-900/50 to-slate-900",
    isNew: true,
  },
  {
    title: "自定义联赛",
    subtitle: "创建你的专属联赛",
    icon: Crown,
    button: "创建联赛",
    bgGradient: "from-purple-900/50 to-slate-900",
    isNew: false,
  },
];

export function QuickStart() {
  return (
    <div className="flex-1 rounded-md border border-border/50 bg-card/30 p-3 lg:p-4 flex flex-col">
      <h3 className="font-semibold mb-3 lg:mb-4 text-muted-foreground text-sm lg:text-base">快速开始</h3>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4 flex-1">
        {gameModes.map((mode) => (
          <div
            key={mode.title}
            className={`relative bg-gradient-to-br ${mode.bgGradient} rounded-md border border-border overflow-hidden group`}
          >
            {mode.isNew && (
              <div className="absolute top-2 right-2 px-2 py-0.5 bg-red-500 text-white text-[10px] font-bold rounded">
                NEW
              </div>
            )}
            
            <div className="p-3 lg:p-4 flex flex-col flex-1">
              <div className="flex-1">
                <h4 className="font-semibold text-sm sm:text-base lg:text-lg">{mode.title}</h4>
                <p className="text-[10px] lg:text-xs text-muted-foreground mt-1 line-clamp-2">
                  {mode.subtitle}
                </p>
              </div>
              
              <div className="flex items-end justify-between mt-2">
                <button className="px-2 sm:px-3 lg:px-4 py-1.5 lg:py-2 bg-secondary hover:bg-secondary/80 border border-border rounded text-[10px] sm:text-xs lg:text-sm transition-colors">
                  {mode.button}
                </button>
                <mode.icon className="w-10 sm:w-12 lg:w-16 h-10 sm:h-12 lg:h-16 text-muted-foreground/30 -mr-1 lg:-mr-2 -mb-1 lg:-mb-2" />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

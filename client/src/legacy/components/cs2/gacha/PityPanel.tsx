"use client";

import { FileText, History, ChevronRight } from "lucide-react";

const dropRates = [
  { tier: "传奇", range: "90+ OVR", rate: "2.00%", color: "text-yellow-500" },
  { tier: "史诗", range: "85-89 OVR", rate: "8.00%", color: "text-purple-500" },
  { tier: "精英", range: "80-84 OVR", rate: "30.00%", color: "text-blue-500" },
  { tier: "优秀", range: "75-79 OVR", rate: "40.00%", color: "text-green-500" },
  { tier: "普通", range: "70-74 OVR", rate: "20.00%", color: "text-muted-foreground" },
];

export function PityPanel() {
  const luckyValue = 23;
  const maxLucky = 100;
  const pityRemaining = 77;

  return (
    <div className="flex flex-col gap-4">
      {/* Top Actions */}
      <div className="flex items-center justify-end gap-3">
        <button className="flex items-center gap-1.5 px-3 py-1.5 text-xs lg:text-sm text-muted-foreground hover:text-foreground border border-border rounded transition-colors">
          <FileText className="w-3.5 h-3.5" />
          <span>卡池概率</span>
        </button>
        <button className="flex items-center gap-1.5 px-3 py-1.5 text-xs lg:text-sm text-muted-foreground hover:text-foreground border border-border rounded transition-colors">
          <History className="w-3.5 h-3.5" />
          <span>抽卡记录</span>
        </button>
      </div>

      {/* Lucky Value Progress */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <span className="text-sm lg:text-base">幸运值</span>
            <span className="text-lg">🍀</span>
          </div>
          <div className="text-sm lg:text-base font-medium">
            <span className="text-primary">{luckyValue}</span>
            <span className="text-muted-foreground"> / {maxLucky}</span>
          </div>
        </div>
        
        <div className="h-2 bg-secondary rounded-full overflow-hidden mb-3">
          <div 
            className="h-full bg-gradient-to-r from-primary to-yellow-500 rounded-full transition-all"
            style={{ width: `${(luckyValue / maxLucky) * 100}%` }}
          />
        </div>

        <div className="flex items-center gap-3">
          <div className="flex-shrink-0 w-12 h-12 lg:w-14 lg:h-14 bg-gradient-to-br from-yellow-600 to-amber-700 rounded flex items-center justify-center">
            <span className="text-lg lg:text-xl font-bold text-white">90+</span>
          </div>
          <div className="text-xs lg:text-sm text-muted-foreground">
            幸运值满必得 <span className="text-primary font-medium">90+ OVR</span> 选手
          </div>
        </div>
      </div>

      {/* Pity System */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <h3 className="text-sm lg:text-base font-semibold mb-3">保底机制</h3>
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-xs lg:text-sm">
            <span className="text-primary">🎯</span>
            <span className="text-muted-foreground">每 100 抽必得</span>
            <span className="text-yellow-500 font-medium">90+ OVR</span>
            <span className="text-muted-foreground">选手</span>
          </div>
          <div className="flex items-center gap-2 text-xs lg:text-sm">
            <span className="text-primary">🎯</span>
            <span className="text-muted-foreground">每 10 抽必得</span>
            <span className="text-purple-500 font-medium">80+ OVR</span>
            <span className="text-muted-foreground">选手</span>
          </div>
        </div>
      </div>

      {/* Drop Rates */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <h3 className="text-sm lg:text-base font-semibold mb-3">卡池概率</h3>
        <div className="space-y-2">
          {dropRates.map((rate) => (
            <div key={rate.tier} className="flex items-center justify-between text-xs lg:text-sm">
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${rate.color.replace("text-", "bg-")}`} />
                <span className={rate.color}>{rate.tier}</span>
                <span className="text-muted-foreground">({rate.range})</span>
              </div>
              <span className={rate.color}>{rate.rate}</span>
            </div>
          ))}
        </div>
        
        <button className="flex items-center gap-1 mt-3 text-xs text-muted-foreground hover:text-foreground transition-colors">
          完整概率说明 <ChevronRight className="w-3 h-3" />
        </button>
      </div>

      {/* Pity Counter */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <div className="flex items-center gap-3">
          <div className="w-12 h-14 lg:w-14 lg:h-16 bg-gradient-to-br from-purple-600 to-purple-800 rounded flex items-center justify-center">
            <span className="text-2xl">🃏</span>
          </div>
          <div>
            <div className="text-xs lg:text-sm text-muted-foreground">
              再抽 <span className="text-primary font-bold text-base lg:text-lg">{pityRemaining}</span> 次
            </div>
            <div className="text-xs lg:text-sm text-muted-foreground">
              必得 <span className="text-yellow-500 font-medium">90+ OVR</span> 选手
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

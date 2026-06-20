"use client";

import { Gift, ExternalLink } from "lucide-react";

export function RankingInfoPanel() {
  return (
    <div className="w-full lg:w-64 flex-shrink-0 flex flex-col gap-4">
      {/* Leaderboard Description */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <h3 className="font-semibold text-sm lg:text-base mb-3">排行榜说明</h3>
        <p className="text-[10px] lg:text-xs text-muted-foreground leading-relaxed">
          综合实力评分基于 ELO 等级、模拟胜场、胜率、K/D、资产总值等多维度数据计算。
        </p>
        <button className="w-full mt-3 py-2 text-xs lg:text-sm text-muted-foreground hover:text-foreground transition-colors border border-border rounded">
          查看完整规则
        </button>
      </div>

      {/* Top Player Rewards */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <h3 className="font-semibold text-sm lg:text-base mb-3">顶级玩家奖励</h3>
        <div className="flex items-start gap-3">
          <div className="w-12 h-12 lg:w-16 lg:h-16 rounded bg-gradient-to-br from-purple-600 to-pink-600 flex items-center justify-center flex-shrink-0">
            <Gift className="w-6 lg:w-8 h-6 lg:h-8 text-white" />
          </div>
          <p className="text-[10px] lg:text-xs text-muted-foreground leading-relaxed">
            赛季结算时，排名前 100 的玩家将获得丰厚奖励！
          </p>
        </div>
        <button className="w-full mt-3 py-2 text-xs lg:text-sm text-muted-foreground hover:text-foreground transition-colors border border-border rounded">
          查看奖励详情
        </button>
      </div>

      {/* My Details */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <h3 className="font-semibold text-sm lg:text-base mb-3">我的详情</h3>
        
        {/* Player Info */}
        <div className="flex items-center gap-3 mb-4">
          <div className="relative">
            <div className="w-12 h-12 lg:w-14 lg:h-14 rounded-md bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center overflow-hidden">
              <span className="text-xl lg:text-2xl font-bold">P</span>
            </div>
            <div className="absolute -bottom-1 -right-1 px-1.5 py-0.5 bg-blue-600 text-[10px] rounded font-bold">
              52
            </div>
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-medium">PlayerOne</span>
              <ExternalLink className="w-3 h-3 text-muted-foreground" />
            </div>
            <div className="text-xs text-muted-foreground">无畏先锋</div>
            <div className="mt-1">
              <div className="flex items-center justify-between text-[10px] mb-0.5">
                <span className="text-muted-foreground">12800 / 20000</span>
              </div>
              <div className="h-1.5 bg-secondary rounded-full overflow-hidden">
                <div className="h-full w-[64%] bg-blue-500 rounded-full" />
              </div>
            </div>
          </div>
        </div>

        {/* Season Stats */}
        <div className="border-t border-border pt-3">
          <div className="text-xs text-muted-foreground mb-2">本赛季数据</div>
          <div className="space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">综合分</span>
              <span className="text-primary font-bold">1,254</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">ELO 等级</span>
              <span>1,256 <span className="text-muted-foreground">(全球前 4.5%)</span></span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">模拟胜场</span>
              <span>312</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">胜率</span>
              <span>54.3%</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">K/D</span>
              <span>1.02</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">资产总值</span>
              <span className="text-green-400">$23,568</span>
            </div>
          </div>
        </div>

        <button className="w-full mt-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground text-xs lg:text-sm font-medium rounded transition-colors">
          查看个人主页
        </button>
      </div>
    </div>
  );
}

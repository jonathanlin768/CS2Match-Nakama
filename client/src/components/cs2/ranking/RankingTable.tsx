"use client";

import { useState } from "react";
import { ChevronLeft, ChevronRight, Search, HelpCircle } from "lucide-react";

const tabs = ["综合实力", "ELO 等级", "赛事积分", "模拟胜场", "胜率", "K/D", "资产总值"];

const players = [
  { rank: 1, name: "Ares", title: "无畏先锋", globalRank: "0.1%", elo: 3256, score: 2852, wins: 1248, winRate: "68.7%", kd: 1.42, assets: "$128,560", avatar: "A" },
  { rank: 2, name: "Knight", title: "战术大师", globalRank: "0.2%", elo: 3192, score: 2750, wins: 1102, winRate: "67.1%", kd: 1.37, assets: "$112,350", avatar: "K" },
  { rank: 3, name: "Nitro", title: "传奇教练", globalRank: "0.3%", elo: 3128, score: 2618, wins: 987, winRate: "66.3%", kd: 1.31, assets: "$98,760", avatar: "N" },
  { rank: 4, name: "Vortex", title: "战术大师", globalRank: "0.4%", elo: 3052, score: 2548, wins: 1023, winRate: "65.8%", kd: 1.29, assets: "$105,420", avatar: "V" },
  { rank: 5, name: "Echo", title: "无畏先锋", globalRank: "0.5%", elo: 3984, score: 2482, wins: 935, winRate: "64.2%", kd: 1.26, assets: "$92,180", avatar: "E" },
  { rank: 6, name: "Phoenix", title: "战术大师", globalRank: "0.6%", elo: 2912, score: 2415, wins: 892, winRate: "63.7%", kd: 1.24, assets: "$89,540", avatar: "P" },
  { rank: 7, name: "Storm", title: "精英指挥", globalRank: "0.7%", elo: 2856, score: 2356, wins: 845, winRate: "62.9%", kd: 1.22, assets: "$85,210", avatar: "S" },
  { rank: 8, name: "Blaze", title: "无畏先锋", globalRank: "0.8%", elo: 2798, score: 2298, wins: 812, winRate: "61.3%", kd: 1.18, assets: "$78,630", avatar: "B" },
  { rank: 9, name: "Shadow", title: "战术大师", globalRank: "1.0%", elo: 2742, score: 2245, wins: 756, winRate: "60.8%", kd: 1.16, assets: "$72,490", avatar: "S" },
  { rank: 10, name: "Comet", title: "精英指挥", globalRank: "1.2%", elo: 2688, score: 2198, wins: 732, winRate: "59.7%", kd: 1.13, assets: "$66,880", avatar: "C" },
  { rank: 52, name: "PlayerOne", title: "无畏先锋", globalRank: "4.5%", elo: 1256, score: 1254, wins: 312, winRate: "54.3%", kd: 1.02, assets: "$23,568", avatar: "P", isCurrentUser: true },
];

const getRankStyle = (rank: number) => {
  if (rank === 1) return "bg-gradient-to-r from-yellow-600 to-yellow-400 text-black";
  if (rank === 2) return "bg-gradient-to-r from-gray-400 to-gray-300 text-black";
  if (rank === 3) return "bg-gradient-to-r from-amber-700 to-amber-500 text-black";
  return "bg-secondary text-muted-foreground";
};

export function RankingTable() {
  const [activeTab, setActiveTab] = useState("综合实力");
  const [currentPage, setCurrentPage] = useState(1);

  return (
    <div className="flex-1 flex flex-col gap-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-lg lg:text-xl font-bold">玩家排行榜</h2>
            <HelpCircle className="w-4 h-4 text-muted-foreground" />
          </div>
          <p className="text-xs lg:text-sm text-muted-foreground mt-1">
            基于玩家的综合表现进行排名，每小时更新一次
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select className="bg-secondary border border-border rounded px-3 py-1.5 text-xs lg:text-sm">
            <option>第 5 赛季 (2024.05.01 - 2024.07.31)</option>
            <option>第 4 赛季</option>
            <option>第 3 赛季</option>
          </select>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 lg:gap-4 overflow-x-auto pb-2 scrollbar-hide">
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-3 lg:px-4 py-1.5 lg:py-2 rounded text-xs lg:text-sm font-medium whitespace-nowrap transition-colors ${
              activeTab === tab
                ? "bg-secondary text-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="bg-card rounded-md border border-border overflow-hidden">
        {/* Table Header */}
        <div className="hidden lg:grid grid-cols-12 gap-2 px-4 py-3 bg-secondary/50 text-xs text-muted-foreground font-medium">
          <div className="col-span-1">排名</div>
          <div className="col-span-2">玩家</div>
          <div className="col-span-1">ELO 等级</div>
          <div className="col-span-1 text-primary">综合分</div>
          <div className="col-span-1">模拟胜场</div>
          <div className="col-span-1">胜率</div>
          <div className="col-span-1">K/D</div>
          <div className="col-span-1">资产总值</div>
          <div className="col-span-3">常用阵容</div>
        </div>

        {/* Table Body */}
        <div className="divide-y divide-border">
          {players.map((player) => (
            <div
              key={player.rank}
              className={`grid grid-cols-2 lg:grid-cols-12 gap-2 px-3 lg:px-4 py-3 items-center hover:bg-secondary/30 transition-colors ${
                player.isCurrentUser ? "bg-primary/5 border-l-2 border-l-primary" : ""
              }`}
            >
              {/* Rank */}
              <div className="col-span-1 flex items-center">
                <span className={`w-7 h-7 lg:w-8 lg:h-8 rounded-full flex items-center justify-center text-xs lg:text-sm font-bold ${getRankStyle(player.rank)}`}>
                  {player.rank}
                </span>
              </div>

              {/* Player Info */}
              <div className="col-span-1 lg:col-span-2 flex items-center gap-2 lg:gap-3">
                <div className="w-8 h-8 lg:w-10 lg:h-10 rounded-md bg-gradient-to-br from-orange-600 to-orange-800 flex items-center justify-center text-sm lg:text-base font-bold flex-shrink-0">
                  {player.avatar}
                </div>
                <div className="min-w-0">
                  <div className="text-sm font-medium truncate">{player.name}</div>
                  <div className="text-[10px] lg:text-xs text-muted-foreground truncate">{player.title}</div>
                </div>
              </div>

              {/* ELO - Desktop */}
              <div className="hidden lg:flex col-span-1 flex-col">
                <span className="text-[10px] text-muted-foreground">全球前 {player.globalRank}</span>
                <span className="text-sm">{player.elo.toLocaleString()}</span>
              </div>

              {/* Score - Desktop */}
              <div className="hidden lg:block col-span-1 text-primary font-bold">
                {player.score.toLocaleString()}
              </div>

              {/* Wins - Desktop */}
              <div className="hidden lg:block col-span-1 text-sm">
                {player.wins.toLocaleString()}
              </div>

              {/* Win Rate - Desktop */}
              <div className="hidden lg:block col-span-1 text-sm">
                {player.winRate}
              </div>

              {/* K/D - Desktop */}
              <div className="hidden lg:block col-span-1 text-sm">
                {player.kd.toFixed(2)}
              </div>

              {/* Assets - Desktop */}
              <div className="hidden lg:block col-span-1 text-sm text-green-400">
                {player.assets}
              </div>

              {/* Team Lineup - Desktop */}
              <div className="hidden lg:flex col-span-3 items-center gap-1">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div
                    key={i}
                    className="w-7 h-7 rounded bg-secondary flex items-center justify-center text-[10px] text-muted-foreground"
                  >
                    {i}
                  </div>
                ))}
              </div>

              {/* Mobile Stats */}
              <div className="col-span-2 lg:hidden flex items-center justify-end gap-4 text-xs">
                <div className="text-primary font-bold">{player.score.toLocaleString()}</div>
                <div className="text-muted-foreground">{player.winRate}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Pagination */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-muted-foreground">
        <div>排行榜数据更新时间：2024-05-24 14:00:00</div>
        <div className="flex items-center gap-2">
          <button className="p-1.5 rounded border border-border hover:bg-secondary transition-colors">
            <ChevronLeft className="w-4 h-4" />
          </button>
          {[1, 2, 3].map((page) => (
            <button
              key={page}
              onClick={() => setCurrentPage(page)}
              className={`w-8 h-8 rounded flex items-center justify-center transition-colors ${
                currentPage === page
                  ? "bg-primary text-primary-foreground"
                  : "border border-border hover:bg-secondary"
              }`}
            >
              {page}
            </button>
          ))}
          <span>...</span>
          <button className="w-8 h-8 rounded border border-border hover:bg-secondary transition-colors">
            100
          </button>
          <button className="p-1.5 rounded border border-border hover:bg-secondary transition-colors">
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="text"
            placeholder="搜索玩家名称"
            className="bg-secondary border border-border rounded px-3 py-1.5 text-sm w-40"
          />
          <button className="p-1.5 rounded bg-secondary border border-border hover:bg-secondary/80 transition-colors">
            <Search className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

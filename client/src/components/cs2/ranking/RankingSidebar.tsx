"use client";

import { Trophy, Users, Award, Target, Layers, DollarSign, BarChart3 } from "lucide-react";

const categories = [
  { id: "overview", label: "排行榜总览", subtitle: "", icon: BarChart3 },
  { id: "player", label: "玩家排行榜", subtitle: "个人综合实力排行", icon: Trophy },
  { id: "club", label: "俱乐部排行榜", subtitle: "俱乐部总积分排行", icon: Users },
  { id: "tournament", label: "赛事排行榜", subtitle: "赛事冠军次数排行", icon: Award },
  { id: "tactics", label: "战术大师榜", subtitle: "战术评分排行", icon: Target },
  { id: "collection", label: "卡牌收集榜", subtitle: "卡牌收集进度排行", icon: Layers },
  { id: "wealth", label: "财富排行榜", subtitle: "资产总值排行", icon: DollarSign },
];

interface RankingSidebarProps {
  selectedCategory: string;
  onSelectCategory: (id: string) => void;
}

export function RankingSidebar({ selectedCategory, onSelectCategory }: RankingSidebarProps) {
  return (
    <div className="w-full lg:w-56 flex-shrink-0 flex flex-col gap-3">
      {/* Category List */}
      <div className="bg-card rounded-md border border-border overflow-hidden">
        {categories.map((category) => (
          <button
            key={category.id}
            onClick={() => onSelectCategory(category.id)}
            className={`w-full flex items-center gap-3 p-3 lg:p-4 text-left transition-colors border-l-2 ${
              selectedCategory === category.id
                ? "bg-primary/10 border-l-primary text-foreground"
                : "border-l-transparent text-muted-foreground hover:bg-secondary/50 hover:text-foreground"
            }`}
          >
            <category.icon className={`w-5 h-5 flex-shrink-0 ${
              selectedCategory === category.id ? "text-primary" : ""
            }`} />
            <div className="min-w-0">
              <div className="text-sm font-medium truncate">{category.label}</div>
              {category.subtitle && (
                <div className="text-[10px] lg:text-xs text-muted-foreground truncate">
                  {category.subtitle}
                </div>
              )}
            </div>
          </button>
        ))}
      </div>

      {/* My Ranking */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <div className="text-xs text-muted-foreground mb-3">我的排名</div>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 lg:w-12 lg:h-12 rounded-md bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center overflow-hidden flex-shrink-0">
            <span className="text-lg lg:text-xl">P</span>
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">PlayerOne</span>
              <span className="px-1.5 py-0.5 bg-blue-600 text-[10px] rounded">52</span>
            </div>
          </div>
          <div className="text-primary font-bold">1,254</div>
        </div>
      </div>
    </div>
  );
}

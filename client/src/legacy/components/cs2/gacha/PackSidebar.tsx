"use client";

import { Crown, Trophy, Sticker, Package, Gift } from "lucide-react";

const packCategories = [
  { id: "legendary", name: "传奇之路", subtitle: "最高可得 95+ OVR 选手卡", icon: Crown, active: true },
  { id: "champion", name: "冠军典藏", subtitle: "纪念 Major 冠军选手", icon: Trophy },
  { id: "sticker", name: "印花胶囊", subtitle: "收集精美印花", icon: Sticker },
  { id: "normal", name: "普通补给", subtitle: "基础道具补给", icon: Package },
  { id: "daily", name: "每日特惠", subtitle: "每日限购折扣卡包", icon: Gift },
];

const myPacks = [
  { name: "传奇之路卡包", count: 12, icon: "🎴" },
  { name: "冠军典藏卡包", count: 8, icon: "🏆" },
  { name: "印花胶囊", count: 45, icon: "🎨" },
  { name: "普通补给卡包", count: 93, icon: "📦" },
  { name: "纪念包", count: 0, icon: "🎁" },
];

interface PackSidebarProps {
  selectedPack: string;
  onSelectPack: (packId: string) => void;
}

export function PackSidebar({ selectedPack, onSelectPack }: PackSidebarProps) {
  return (
    <div className="flex flex-col gap-4">
      {/* Pack Categories */}
      <div className="bg-card rounded-md border border-border">
        {packCategories.map((pack) => (
          <button
            key={pack.id}
            onClick={() => onSelectPack(pack.id)}
            className={`w-full flex items-center gap-3 p-3 lg:p-4 transition-colors border-l-2 ${
              selectedPack === pack.id
                ? "bg-primary/10 border-l-primary text-foreground"
                : "border-l-transparent text-muted-foreground hover:bg-secondary/50 hover:text-foreground"
            }`}
          >
            <div className={`w-8 lg:w-10 h-8 lg:h-10 rounded flex items-center justify-center ${
              selectedPack === pack.id ? "bg-primary/20" : "bg-secondary"
            }`}>
              <pack.icon className={`w-4 lg:w-5 h-4 lg:h-5 ${
                selectedPack === pack.id ? "text-primary" : "text-muted-foreground"
              }`} />
            </div>
            <div className="text-left flex-1 min-w-0">
              <div className="text-xs lg:text-sm font-medium truncate">{pack.name}</div>
              <div className="text-[10px] lg:text-xs text-muted-foreground truncate">{pack.subtitle}</div>
            </div>
          </button>
        ))}
      </div>

      {/* My Packs */}
      <div className="bg-card rounded-md border border-border p-3 lg:p-4">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm lg:text-base font-semibold">我的卡包</h3>
          <span className="text-xs lg:text-sm text-primary font-medium">158</span>
        </div>
        
        <div className="space-y-2">
          {myPacks.map((pack, index) => (
            <div key={index} className="flex items-center justify-between py-1.5">
              <div className="flex items-center gap-2">
                <span className="text-sm">{pack.icon}</span>
                <span className="text-xs lg:text-sm text-muted-foreground">{pack.name}</span>
              </div>
              <span className="text-xs lg:text-sm font-medium">{pack.count}</span>
            </div>
          ))}
        </div>

        <button className="w-full mt-3 py-2 text-xs lg:text-sm text-muted-foreground hover:text-foreground transition-colors border border-border rounded">
          前往我的卡包
        </button>
      </div>
    </div>
  );
}

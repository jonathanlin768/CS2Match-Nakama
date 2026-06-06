"use client";

import { ChevronRight } from "lucide-react";

const poolCards = [
  { id: 1, name: "s1mple", ovr: 93, position: "AWP", rarity: "legendary" },
  { id: 2, name: "dev1ce", ovr: 92, position: "AWP", rarity: "legendary" },
  { id: 3, name: "NiKo", ovr: 91, position: "RIF", rarity: "legendary" },
  { id: 4, name: "ZywOo", ovr: 89, position: "AWP", rarity: "epic" },
  { id: 5, name: "m0NESY", ovr: 89, position: "AWP", rarity: "epic" },
  { id: 6, name: "ropz", ovr: 87, position: "RIF", rarity: "epic" },
  { id: 7, name: "jks", ovr: 85, position: "RIF", rarity: "epic" },
  { id: 8, name: "Twistzz", ovr: 84, position: "RIF", rarity: "elite" },
  { id: 9, name: "HooXi", ovr: 82, position: "IGL", rarity: "elite" },
  { id: 10, name: "MAGISK", ovr: 81, position: "RIF", rarity: "elite" },
  { id: 11, name: "electroNic", ovr: 79, position: "RIF", rarity: "rare" },
  { id: 12, name: "chopper", ovr: 77, position: "RIF", rarity: "rare" },
  { id: 13, name: "FL1T", ovr: 74, position: "RIF", rarity: "common" },
  { id: 14, name: "n0rb3r7", ovr: 72, position: "RIF", rarity: "common" },
];

function MiniCard({ card }: { card: typeof poolCards[0] }) {
  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case "legendary": return "from-yellow-600 to-amber-700 border-yellow-500/50";
      case "epic": return "from-purple-600 to-purple-800 border-purple-500/50";
      case "elite": return "from-blue-600 to-blue-800 border-blue-500/50";
      case "rare": return "from-green-600 to-green-800 border-green-500/50";
      default: return "from-gray-600 to-gray-800 border-gray-500/50";
    }
  };

  const getOvrColor = (rarity: string) => {
    switch (rarity) {
      case "legendary": return "bg-yellow-600";
      case "epic": return "bg-purple-600";
      case "elite": return "bg-blue-600";
      case "rare": return "bg-green-600";
      default: return "bg-gray-600";
    }
  };

  return (
    <div className="flex-shrink-0 w-16 sm:w-18 lg:w-20">
      <div className={`relative aspect-[3/4] rounded border bg-gradient-to-b ${getRarityColor(card.rarity)} overflow-hidden`}>
        {/* OVR Badge */}
        <div className={`absolute top-1 left-1 px-1 py-0.5 ${getOvrColor(card.rarity)} rounded text-[10px] font-bold text-white`}>
          {card.ovr}
        </div>
        <div className="absolute top-1 right-1 text-[8px] text-muted-foreground bg-black/50 px-1 rounded">
          {card.position}
        </div>

        {/* Player Image Placeholder */}
        <div className="w-full h-full flex items-center justify-center">
          <span className="text-xl lg:text-2xl">👤</span>
        </div>
      </div>
      <div className="text-center mt-1">
        <div className="text-[10px] lg:text-xs font-medium truncate">{card.name}</div>
      </div>
    </div>
  );
}

export function PoolPreview() {
  return (
    <div className="bg-card/50 rounded-md border border-border p-3 lg:p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm lg:text-base font-semibold">卡池预览</h3>
        <div className="text-[10px] lg:text-xs text-muted-foreground">
          所有卡牌均可在市场交易
        </div>
      </div>

      <div className="relative">
        <div className="flex gap-2 lg:gap-3 overflow-x-auto pb-2 scrollbar-hide">
          {poolCards.map((card) => (
            <MiniCard key={card.id} card={card} />
          ))}
          
          {/* View More Button */}
          <div className="flex-shrink-0 w-16 sm:w-18 lg:w-20 flex items-center justify-center">
            <button className="flex flex-col items-center gap-1 text-muted-foreground hover:text-foreground transition-colors">
              <div className="w-10 h-10 rounded-full border border-border flex items-center justify-center">
                <ChevronRight className="w-5 h-5" />
              </div>
              <span className="text-[10px]">查看更多</span>
            </button>
          </div>
        </div>
      </div>

      {/* Pagination Dots */}
      <div className="flex items-center justify-center gap-2 mt-3">
        <div className="w-6 h-1.5 rounded-full bg-primary" />
        <div className="w-1.5 h-1.5 rounded-full bg-muted-foreground/30" />
        <div className="w-1.5 h-1.5 rounded-full bg-muted-foreground/30" />
        <div className="w-1.5 h-1.5 rounded-full bg-muted-foreground/30" />
        <div className="w-1.5 h-1.5 rounded-full bg-muted-foreground/30" />
      </div>
    </div>
  );
}

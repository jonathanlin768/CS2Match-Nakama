"use client";

import { useState } from "react";
import { ChevronLeft, ChevronRight, HelpCircle } from "lucide-react";

const featuredCards = [
  { id: 1, name: "ropz", team: "FaZe", ovr: 87, position: "RIF", rarity: "epic", teamLogo: "🔴" },
  { id: 2, name: "ZywOo", team: "Vitality", ovr: 89, position: "AWP", rarity: "epic", teamLogo: "🐝" },
  { id: 3, name: "s1mple", team: "NAVI", ovr: 93, position: "AWP", rarity: "legendary", teamLogo: "🟡" },
  { id: 4, name: "donk", team: "Spirit", ovr: 88, position: "RIF", rarity: "epic", teamLogo: "🔵" },
  { id: 5, name: "rain", team: "FaZe", ovr: 85, position: "RIF", rarity: "elite", teamLogo: "🔴" },
];

const tags = ["90+ OVR 保底", "可交易", "可出售", "可用于阵容"];

function PlayerCard({ card, isCenter }: { card: typeof featuredCards[0]; isCenter: boolean }) {
  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case "legendary": return "from-yellow-600 via-amber-500 to-yellow-600";
      case "epic": return "from-purple-600 via-purple-500 to-purple-600";
      case "elite": return "from-blue-600 via-blue-500 to-blue-600";
      default: return "from-gray-600 via-gray-500 to-gray-600";
    }
  };

  const getBorderColor = (rarity: string) => {
    switch (rarity) {
      case "legendary": return "border-yellow-500/50 shadow-yellow-500/20";
      case "epic": return "border-purple-500/50 shadow-purple-500/20";
      case "elite": return "border-blue-500/50 shadow-blue-500/20";
      default: return "border-gray-500/50";
    }
  };

  return (
    <div className={`relative flex-shrink-0 transition-all duration-300 ${
      isCenter 
        ? "w-32 sm:w-40 lg:w-48 z-10 scale-110" 
        : "w-24 sm:w-32 lg:w-36 opacity-80 scale-90"
    }`}>
      <div className={`relative rounded-md overflow-hidden border-2 ${getBorderColor(card.rarity)} ${
        isCenter ? "shadow-lg shadow-yellow-500/30" : "shadow-md"
      }`}>
        {/* Card Background */}
        <div className={`bg-gradient-to-b ${getRarityColor(card.rarity)} p-0.5`}>
          <div className="bg-card/95 rounded-sm">
            {/* OVR Badge */}
            <div className="absolute top-2 left-2 z-10">
              <div className={`px-1.5 py-0.5 rounded text-xs lg:text-sm font-bold bg-gradient-to-r ${getRarityColor(card.rarity)} text-white`}>
                {card.ovr}
              </div>
              <div className="text-[8px] lg:text-[10px] text-muted-foreground mt-0.5">{card.position}</div>
            </div>

            {/* Player Image Placeholder */}
            <div className="aspect-[3/4] bg-gradient-to-b from-secondary/50 to-secondary flex items-center justify-center">
              <div className="text-3xl sm:text-4xl lg:text-5xl">👤</div>
            </div>

            {/* Player Info */}
            <div className="p-2 lg:p-3 text-center bg-gradient-to-t from-black/80 to-transparent -mt-8 relative z-10 pt-10">
              <div className="text-xs sm:text-sm lg:text-base font-bold truncate">{card.name}</div>
              <div className="flex items-center justify-center gap-1 mt-1">
                <span className="text-sm">{card.teamLogo}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Glow effect for center card */}
        {isCenter && card.rarity === "legendary" && (
          <div className="absolute inset-0 bg-gradient-to-t from-yellow-500/20 via-transparent to-transparent pointer-events-none" />
        )}
      </div>
    </div>
  );
}

export function CardCarousel() {
  const [currentIndex, setCurrentIndex] = useState(2); // Start with s1mple in center

  const handlePrev = () => {
    setCurrentIndex((prev) => (prev === 0 ? featuredCards.length - 1 : prev - 1));
  };

  const handleNext = () => {
    setCurrentIndex((prev) => (prev === featuredCards.length - 1 ? 0 : prev + 1));
  };

  // Reorder cards to show current in center
  const getVisibleCards = () => {
    const result = [];
    for (let i = -2; i <= 2; i++) {
      const index = (currentIndex + i + featuredCards.length) % featuredCards.length;
      result.push({ ...featuredCards[index], isCenter: i === 0 });
    }
    return result;
  };

  return (
    <div className="flex-1">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-4">
        <div className="flex items-center gap-2">
          <h2 className="text-lg lg:text-2xl font-bold">传奇之路</h2>
          <button className="text-muted-foreground hover:text-foreground">
            <HelpCircle className="w-4 h-4" />
          </button>
        </div>
        <div className="text-xs lg:text-sm text-muted-foreground">
          活动结束倒计时：<span className="text-primary font-medium">19天 23:59:59</span>
        </div>
      </div>

      {/* Tags */}
      <div className="flex flex-wrap gap-2 mb-6">
        {tags.map((tag) => (
          <span
            key={tag}
            className="px-2 lg:px-3 py-1 text-[10px] lg:text-xs border border-border rounded bg-secondary/50 text-muted-foreground"
          >
            {tag}
          </span>
        ))}
      </div>

      {/* Card Carousel */}
      <div className="relative flex items-center justify-center py-4 lg:py-8">
        {/* Left Arrow */}
        <button
          onClick={handlePrev}
          className="absolute left-0 z-20 p-2 bg-card/80 border border-border rounded-full hover:bg-secondary transition-colors"
        >
          <ChevronLeft className="w-5 h-5" />
        </button>

        {/* Cards */}
        <div className="flex items-center justify-center gap-2 sm:gap-3 lg:gap-4 px-10 lg:px-16">
          {getVisibleCards().map((card, index) => (
            <PlayerCard key={`${card.id}-${index}`} card={card} isCenter={card.isCenter} />
          ))}
        </div>

        {/* Right Arrow */}
        <button
          onClick={handleNext}
          className="absolute right-0 z-20 p-2 bg-card/80 border border-border rounded-full hover:bg-secondary transition-colors"
        >
          <ChevronRight className="w-5 h-5" />
        </button>
      </div>

      {/* Carousel Dots */}
      <div className="flex items-center justify-center gap-2 mb-6">
        {featuredCards.map((_, index) => (
          <button
            key={index}
            onClick={() => setCurrentIndex(index)}
            className={`w-2 h-2 rounded-full transition-colors ${
              index === currentIndex ? "bg-primary" : "bg-muted-foreground/30"
            }`}
          />
        ))}
      </div>

      {/* Draw Buttons */}
      <div className="flex flex-col sm:flex-row items-center justify-center gap-3 lg:gap-4 mb-4">
        <button className="w-full sm:w-auto px-8 lg:px-12 py-3 bg-secondary hover:bg-secondary/80 border border-border rounded transition-colors">
          <div className="text-sm lg:text-base font-medium">抽 1 次</div>
          <div className="flex items-center justify-center gap-1 text-primary text-xs lg:text-sm">
            <span>💎</span> 200
          </div>
        </button>
        
        <div className="relative">
          <div className="absolute -top-2 -right-2 px-1.5 py-0.5 bg-red-500 text-white text-[10px] font-bold rounded">
            -10%
          </div>
          <button className="w-full sm:w-auto px-8 lg:px-12 py-3 bg-gradient-to-r from-primary to-yellow-600 hover:from-primary/90 hover:to-yellow-600/90 text-primary-foreground rounded transition-colors">
            <div className="text-sm lg:text-base font-medium">抽 10 次</div>
            <div className="flex items-center justify-center gap-1 text-xs lg:text-sm">
              <span>💎</span> 1800
            </div>
          </button>
        </div>
      </div>

      {/* Skip Animation */}
      <div className="flex items-center justify-center gap-2 text-xs lg:text-sm text-muted-foreground">
        <input type="checkbox" id="skipAnimation" className="rounded border-border" />
        <label htmlFor="skipAnimation">跳过抽卡动画</label>
      </div>
    </div>
  );
}

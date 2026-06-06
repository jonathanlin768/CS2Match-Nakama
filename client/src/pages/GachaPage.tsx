"use client";

import { useState } from "react";
import { PackSidebar } from "@/components/cs2/gacha/PackSidebar";
import { CardCarousel } from "@/components/cs2/gacha/CardCarousel";
import { PityPanel } from "@/components/cs2/gacha/PityPanel";
import { PoolPreview } from "@/components/cs2/gacha/PoolPreview";

export default function GachaPage() {
  const [selectedPack, setSelectedPack] = useState("legendary");

  return (
    <div className="flex-1 p-4 lg:p-6">
      <div className="max-w-[1400px] mx-auto flex flex-col gap-4 lg:gap-6">
        {/* Main Content */}
        <div className="flex flex-col lg:flex-row gap-4 lg:gap-6">
          {/* Left Sidebar - Pack Selection */}
          <div className="w-full lg:w-56 xl:w-64 flex-shrink-0">
            <PackSidebar selectedPack={selectedPack} onSelectPack={setSelectedPack} />
          </div>

          {/* Center - Card Carousel */}
          <div className="flex-1 flex flex-col">
            <CardCarousel />
          </div>

          {/* Right Panel - Pity System & Rates */}
          <div className="w-full lg:w-64 xl:w-72 flex-shrink-0">
            <PityPanel />
          </div>
        </div>

        {/* Bottom - Pool Preview */}
        <PoolPreview />
      </div>
    </div>
  );
}

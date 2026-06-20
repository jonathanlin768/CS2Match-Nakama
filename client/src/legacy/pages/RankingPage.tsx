"use client";

import { useState } from "react";
import { RankingSidebar } from "@/components/cs2/ranking/RankingSidebar";
import { RankingTable } from "@/components/cs2/ranking/RankingTable";
import { RankingInfoPanel } from "@/components/cs2/ranking/RankingInfoPanel";

export default function RankingPage() {
  const [selectedCategory, setSelectedCategory] = useState("player");

  return (
    <div className="flex-1 p-4 lg:p-6">
      <div className="max-w-[1400px] mx-auto flex flex-col lg:flex-row gap-4 lg:gap-6">
        {/* Left Sidebar */}
        <RankingSidebar
          selectedCategory={selectedCategory}
          onSelectCategory={setSelectedCategory}
        />

        {/* Main Content */}
        <RankingTable />

        {/* Right Panel */}
        <RankingInfoPanel />
      </div>
    </div>
  );
}

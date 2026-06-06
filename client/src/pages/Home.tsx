import { PlayerSidebar } from "@/components/cs2/PlayerSidebar";
import { SeasonBanner } from "@/components/cs2/SeasonBanner";
import { DailyTasks } from "@/components/cs2/DailyTasks";
import { ChampionshipCenter } from "@/components/cs2/ChampionshipCenter";
import { QuickStart } from "@/components/cs2/QuickStart";
import { ActivityCenter } from "@/components/cs2/ActivityCenter";
import { TeamRankings } from "@/components/cs2/TeamRankings";
import { RecentMatches } from "@/components/cs2/RecentMatches";
import { CardPack } from "@/components/cs2/CardPack";

export default function Home() {
  return (
    <div className="flex-1 p-4 lg:p-6">
      <div className="max-w-[1400px] mx-auto flex flex-col gap-4 lg:gap-6">
        {/* Top Row - Player Sidebar + Season Banner + Daily Tasks + Championship */}
        {/* Mobile: Stack vertically, Desktop: Horizontal */}
        <div className="flex flex-col lg:flex-row gap-4 lg:gap-6 lg:items-stretch">
          {/* Left Sidebar - Player Info */}
          <div className="w-full lg:w-48 flex">
            <PlayerSidebar />
          </div>

          {/* Season Banner */}
          <div className="flex-1 flex">
            <SeasonBanner />
          </div>
          
          {/* Daily Tasks & Championship - Side by side on tablet, separate on desktop */}
          <div className="flex flex-col sm:flex-row lg:flex-col xl:flex-row gap-4 lg:gap-6">
            <div className="flex-1 lg:w-72 xl:w-72 flex">
              <DailyTasks />
            </div>
            <div className="flex-1 lg:w-64 xl:w-64 flex">
              <ChampionshipCenter />
            </div>
          </div>
        </div>

        {/* Quick Start Section */}
        <div className="flex flex-col lg:flex-row gap-4 lg:gap-6 lg:items-stretch">
          <div className="flex-1 flex">
            <QuickStart />
          </div>
          <div className="w-full lg:w-64 flex">
            <ActivityCenter />
          </div>
        </div>

        {/* Bottom Row - Team Rankings + Recent Matches + Card Pack */}
        <div className="flex flex-col xl:flex-row gap-4 lg:gap-6 xl:items-stretch">
          <div className="w-full xl:w-[700px] xl:flex-shrink-0 flex">
            <TeamRankings />
          </div>
          <div className="flex flex-col sm:flex-row xl:flex-row gap-4 lg:gap-6 flex-1">
            <div className="flex-1 flex">
              <RecentMatches />
            </div>
            <div className="w-full sm:w-64 flex">
              <CardPack />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

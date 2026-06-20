import TopBar from "../components/lobby/TopBar"
import PromoBar from "../components/lobby/PromoBar"
import LeftPanel from "../components/lobby/LeftPanel"
import CenterStage from "../components/lobby/CenterStage"
import RightPanel from "../components/lobby/RightPanel"
import BottomNav from "../components/lobby/BottomNav"

export default function HomePage() {
  return (
    // Full viewport wrapper that centers the fixed 1920x900 game frame
    <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
      {/* Fixed game frame: 1920 x 900, with 20px inner padding */}
      <main className="flex h-[900px] w-[1920px] shrink-0 flex-col gap-5 overflow-hidden bg-background px-[80px] py-[15px]">
        {/* Top bar: profile, currencies, system icons */}
        <TopBar />

        {/* Promo / event shortcut icons */}
        <PromoBar />

        {/* Main three-column content area */}
        <div className="flex h-[500px] gap-5">
          <LeftPanel />
          <div className="h-[500px] w-[790px] py-3">
            <CenterStage />
          </div>
          <RightPanel />
        </div>

        {/* Bottom navigation */}
        <BottomNav />
      </main>
    </div>
  )
}

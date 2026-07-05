import { useNavigate } from "react-router-dom"
import SubPageHeader from "../components/SubPageHeader"

export default function MatchPage() {
  const navigate = useNavigate()

  return (
    // Full viewport wrapper that centers the fixed 1920x900 game frame
    <div className="flex min-h-screen w-screen items-center justify-center overflow-hidden bg-black">
      {/* Fixed game frame: 1920 x 900 — blank canvas with a single bottom-right CTA */}
      <main className="relative flex h-[900px] w-[1920px] shrink-0 flex-col overflow-hidden bg-background">
        {/* 二级界面头部：← 匹配比赛 */}
        <SubPageHeader title="匹配比赛" />

        {/* 开始匹配 — pinned to the bottom-right corner */}
        <button
          type="button"
          onClick={() => navigate("/battle")}
          className="absolute bottom-[60px] right-[80px] rounded-md bg-gradient-to-r from-gold to-gold/80 px-12 py-4 text-xl font-bold text-background shadow-lg transition-all hover:from-gold/90 hover:to-gold/70 active:scale-[0.98]"
        >
          开始匹配
        </button>
      </main>
    </div>
  )
}

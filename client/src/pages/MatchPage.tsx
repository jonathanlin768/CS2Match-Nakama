import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import SubPageHeader from "../components/SubPageHeader"
import { useAuth } from "../context/AuthContext"
import { debugSimuMatch } from "../api/simu"
import type { MatchReport } from "../types/match-report"

export default function MatchPage() {
  const navigate = useNavigate()
  const { session } = useAuth()
  const [loading, setLoading] = useState(false)

  const handleStartMatch = async () => {
    if (!session) {
      toast.error("请先登录")
      return
    }

    setLoading(true)
    try {
      const report: MatchReport = await debugSimuMatch(session, "de_dust2")
      // 把战报通过 router state 传递到 BattlePage
      navigate("/battle", { state: { report } })
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      toast.error(`模拟战斗失败，请重试：${message}`)
    } finally {
      setLoading(false)
    }
  }

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
          onClick={handleStartMatch}
          disabled={loading}
          className="absolute bottom-[60px] right-[80px] rounded-md bg-gradient-to-r from-gold to-gold/80 px-12 py-4 text-xl font-bold text-background shadow-lg transition-all hover:from-gold/90 hover:to-gold/70 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {loading ? "匹配中..." : "开始匹配"}
        </button>
      </main>
    </div>
  )
}

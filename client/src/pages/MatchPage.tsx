import { Bot, ChevronLeft, Radio, Users } from "lucide-react"
import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import { simuMatch } from "../api/simu"
import { useAuth } from "../context/AuthContext"

export default function MatchPage() {
  const navigate = useNavigate()
  const { session } = useAuth()
  const [loading, setLoading] = useState(false)

  async function startComputerBattle() {
    if (!session) return
    setLoading(true)
    try {
      const report = await simuMatch(session, { mode: "computer" })
      navigate("/battle", { state: { report } })
    } catch (error) {
      toast.error(`生成比赛失败：${error instanceof Error ? error.message : String(error)}`)
    } finally { setLoading(false) }
  }

  return <div className="sub-page match-page">
    <button className="back-button" onClick={() => navigate("/")}><ChevronLeft size={18} />返回主页</button>
    <header className="page-heading"><p className="eyebrow">MATCH</p><h1>选择对战方式</h1><p>本期电脑对战会在一次请求内生成完整比赛战报；路人匹配仍在开发。</p></header>
    <div className="mode-grid">
      <button className="mode-card available" onClick={() => void startComputerBattle()} disabled={loading}>
        <span><Bot size={34} /></span><div><small>随时可玩</small><h2>电脑对战</h2><p>Falcons 对阵 Vitality，使用服务器真实配置和完整比赛引擎。</p></div><strong>{loading ? "正在准备比赛…" : "开始模拟"}</strong>
      </button>
      <button className="mode-card" disabled><span><Users size={34} /></span><div><small>真人队列</small><h2>路人匹配</h2><p>需要权威房间、断线恢复与实时同步，将在独立版本开放。</p></div><strong><Radio size={15} />未开放</strong></button>
    </div>
    {loading && <div className="request-status"><span className="app-spinner" /><div><b>正在请求电脑模拟</b><p>服务器正在构建阵容并生成完整比赛战报，请稍候。</p></div></div>}
  </div>
}

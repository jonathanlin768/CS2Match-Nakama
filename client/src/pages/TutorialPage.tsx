import { ChevronLeft, Coins, ShieldCheck, Swords } from "lucide-react"
import { useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"
import { simuMatch } from "../api/simu"
import { useAuth } from "../context/AuthContext"
import { useGameConfig } from "../hooks/useGameConfig"
import type { Player } from "../config"
import { toggleTutorialPlayer, tutorialSelectionCost, tutorialSelectionReady } from "../game/tutorial-selection"
import PlayerFullCard from "../components/player/PlayerFullCard"

export default function TutorialPage() {
  const navigate = useNavigate()
  const { session } = useAuth()
  const { tables, loading, error } = useGameConfig()
  const [selected, setSelected] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)
  const config = tables?.TbTutorialBattle.getDataList().find((row) => row.enabled)
  const tiers = useMemo(() => config ? [
    { price: 5, ids: config.tier5PlayerIds }, { price: 4, ids: config.tier4PlayerIds }, { price: 3, ids: config.tier3PlayerIds },
    { price: 2, ids: config.tier2PlayerIds }, { price: 1, ids: config.tier1PlayerIds },
  ] : [], [config])
  const total = config ? tutorialSelectionCost(config, selected) : 0

  function toggle(player: Player) {
    if (!config) return
    const result = toggleTutorialPlayer(config, selected, player.id)
    if (result.error === "full") toast.info(`阵容只能选择 ${config.rosterSize} 人`)
    else if (result.error === "budget") toast.info(`预算不足，还剩 ${config.budget - total} 元`)
    else if (result.error === "unknown") toast.error("选手不在当前教学配置中")
    else setSelected(result.selected)
  }

  async function start() {
    if (!session || !config || !tutorialSelectionReady(config, selected)) return
    setSubmitting(true)
    try {
      const report = await simuMatch(session, { mode: "tutorial", tutorial_config_id: config.id, config_version: config.version, player_ids: selected })
      navigate("/battle", { state: { report } })
    } catch (reason) { toast.error(`教学战启动失败：${reason instanceof Error ? reason.message : String(reason)}`) }
    finally { setSubmitting(false) }
  }

  if (loading) return <div className="page-center"><span className="app-spinner" /><p>读取教学阵容配置…</p></div>
  if (error || !tables || !config) return <div className="page-center"><p>教学配置暂不可用：{error ?? "没有启用的方案"}</p><button className="secondary-button" onClick={() => navigate("/")}>返回主页</button></div>

  return <div className="sub-page tutorial-page">
    <button className="back-button" onClick={() => navigate("/")}><ChevronLeft size={18} />退出教学</button>
    <header className="tutorial-heading"><div><p className="eyebrow">15 元组队挑战</p><h1>选出你的五人阵容</h1><p>从每个价格档自由选择。服务端会再次校验价格、人数与选手 ID。</p></div><div className="budget-panel"><small>剩余预算</small><strong><Coins size={22} />{config.budget - total}</strong><span>{selected.length} / {config.rosterSize} 人</span></div></header>
    <div className="tutorial-player-grid">
      {tiers.flatMap((tier) => tier.ids.map((id) => {
        const player = tables.TbPlayer.get(id)
        if (!player) return []
        const active = selected.includes(id)
        return <button key={id} className={active ? "player-pick selected" : "player-pick"} onClick={() => toggle(player)} aria-pressed={active}>
          <span className="player-pick-price"><strong>{tier.price}</strong> 元</span>
          <PlayerFullCard cardImage={player.cardImage} portrait={player.portrait} alt={`${player.name} 完整卡面`} className="tutorial-player-card" />
          <span className="player-pick-info"><b>{player.name}</b><small>{player.teamId_ref?.shortName ?? player.teamId} · {player.positions.join(" / ")}</small></span>
          {active && <ShieldCheck className="player-pick-check" size={22} />}
        </button>
      }))}
    </div>
    <footer className="tutorial-footer"><div><span>已选阵容</span><b>{selected.map((id) => tables.TbPlayer.get(id)?.name).join(" · ") || "还没有选择选手"}</b></div><button className="primary-button battle-button" disabled={!tutorialSelectionReady(config, selected) || submitting} onClick={() => void start()}><Swords size={19} />{submitting ? "正在生成战报…" : "迎战机器人"}</button></footer>
  </div>
}

import { Bot, ChevronRight, Settings2, Swords, UsersRound, X } from "lucide-react"
import { useState } from "react"
import { useNavigate } from "react-router-dom"

const FIRST_VISIT_KEY = "major_manager_onboarding_prompt_v1"

export default function HomePage() {
  const navigate = useNavigate()
  const [unavailable, setUnavailable] = useState<string | null>(null)
  const [introOpen, setIntroOpen] = useState(() => !localStorage.getItem(FIRST_VISIT_KEY))

  function dismissIntro(play: boolean) {
    localStorage.setItem(FIRST_VISIT_KEY, "seen")
    setIntroOpen(false)
    if (play) navigate("/tutorial")
  }

  return (
    <div className="home-page">
      <section className="hero-copy">
        <p className="eyebrow">BUILD · BATTLE · REPLAY</p>
        <h1>选一支队，<br /><span>打一场真正的模拟。</span></h1>
        <p>没有活动红点，没有复杂养成。选择入口，几秒后进入完整比赛回放。</p>
      </section>
      <section className="action-grid" aria-label="游戏入口">
        <button className="action-card muted-action" onClick={() => setUnavailable("调整阵容")}>
          <span className="action-icon"><Settings2 /></span><small>阵容管理</small><h2>调整阵容</h2><p>打造你的五人首发</p><em>未开放</em>
        </button>
        <button className="action-card muted-action" onClick={() => setUnavailable("好友对战")}>
          <span className="action-icon"><UsersRound /></span><small>私人房间</small><h2>好友对战</h2><p>向已添加好友发起挑战</p><em>未开放</em>
        </button>
        <button className="action-card primary-action" onClick={() => navigate("/match")}>
          <span className="action-icon"><Swords /></span><small>立即开始</small><h2>匹配对战</h2><p><Bot size={16} /> 当前支持电脑模拟</p><strong>选择模式 <ChevronRight size={20} /></strong>
        </button>
      </section>
      <button className="tutorial-link" onClick={() => navigate("/tutorial")}>第一次玩？体验 15 元组队教学 <ChevronRight size={16} /></button>

      {introOpen && <div className="modal-backdrop"><section className="modal-card onboarding-modal" role="dialog" aria-modal="true">
        <button className="modal-close" onClick={() => dismissIntro(false)}><X size={20} /></button>
        <span className="modal-illustration"><Swords size={38} /></span><p className="eyebrow">新玩家 · 约 3 分钟</p><h2>要先体验一场战斗吗？</h2>
        <p className="modal-copy">你有 15 元，从五个价格档里选出五名选手。我们会为你安排一支机器人战队，并用真实选手数据跑完整场比赛。</p>
        <div className="modal-actions"><button className="secondary-button" onClick={() => dismissIntro(false)}>先看看主页</button><button className="primary-button" onClick={() => dismissIntro(true)}>开始体验</button></div>
      </section></div>}
      {unavailable && <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setUnavailable(null) }}><section className="modal-card compact-modal"><p className="eyebrow">COMING SOON</p><h2>{unavailable}</h2><p className="modal-copy">这个功能还没有开放。本期先把“选阵容—触发比赛—完整回放”的核心体验做好。</p><button className="primary-button" onClick={() => setUnavailable(null)}>知道了</button></section></div>}
    </div>
  )
}

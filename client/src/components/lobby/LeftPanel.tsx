import { Gift, Lock, MessageSquare } from "lucide-react"

function Card({
  children,
  className = "",
  onClick,
}: {
  children: React.ReactNode
  className?: string
  onClick?: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`relative overflow-hidden rounded-md ring-1 ring-white/10 transition hover:ring-gold/40 active:scale-[0.98] ${className}`}
    >
      {children}
    </button>
  )
}

export default function LeftPanel() {
  return (
    <div className="relative flex h-[500px] w-[480px] shrink-0 flex-col gap-3 p-3">
      {/* Monthly card promo */}
      <Card className="h-[142px] bg-gradient-to-br from-[#3a2f1a] to-[#1d1812]">
        <div className="absolute inset-0 flex flex-col justify-center px-4 text-left">
          <span className="font-display text-[10px] tracking-[0.3em] text-white/30">ARENA</span>
          <div className="flex items-center gap-2">
            <span className="text-2xl font-black text-gold drop-shadow">超值月卡</span>
          </div>
          <span className="mt-1 font-display text-xl font-bold text-foreground">X3000万</span>
          <div className="mt-1 flex gap-2 text-xs font-bold">
            <span className="rounded bg-accent px-1.5 py-0.5 text-accent-foreground">UP X40</span>
            <span className="rounded bg-purple-600 px-1.5 py-0.5 text-white">SSR X80</span>
          </div>
        </div>
      </Card>

      {/* Activity center + Hot list */}
      <div className="grid grid-cols-2 gap-3">
        <Card className="relative h-[142px] bg-gradient-to-br from-panel-light to-panel">
          <span className="absolute right-1.5 top-1.5 grid h-5 w-5 place-items-center rounded-full bg-accent">
            <Gift size={11} className="text-accent-foreground" />
          </span>
          <div className="absolute inset-0 flex flex-col justify-end p-2 text-left">
            <span className="font-display text-[9px] leading-tight tracking-widest text-white/25">
              ACTIVITY CENTER
            </span>
            <span className="text-lg font-black text-foreground">活动中心</span>
          </div>
        </Card>

        <Card className="relative h-[142px] bg-gradient-to-br from-[#2a2230] to-panel">
          <span className="absolute right-1.5 top-1.5 flex items-center gap-1 rounded-full bg-black/60 px-1.5 py-0.5 text-[9px] text-gold">
            <Lock size={9} />
            S1-60
          </span>
          <div className="absolute inset-0 flex flex-col justify-end p-2 text-left">
            <span className="text-lg font-black text-foreground">热门榜</span>
          </div>
        </Card>
      </div>

      {/* Recruitment banner */}
      <Card className="relative h-[168px] bg-gradient-to-br from-[#22272e] to-[#10131a]">
        <span className="absolute right-2 top-2 h-2.5 w-2.5 rounded-full bg-gold shadow-[0_0_8px] shadow-gold" />
        <div className="absolute inset-0 flex flex-col justify-end p-3 text-left">
          <span className="font-display text-[10px] tracking-[0.3em] text-white/20">RECRUITMENT</span>
          <span className="text-2xl font-black text-foreground">招募</span>
        </div>
      </Card>

      {/* Chat bubble */}
      <button
        type="button"
        aria-label="聊天"
        className="absolute -bottom-2 -left-2 grid h-9 w-9 place-items-center rounded-md bg-panel ring-1 ring-white/10 transition hover:ring-gold/40 active:scale-95"
      >
        <MessageSquare size={18} className="text-muted" />
      </button>
    </div>
  )
}

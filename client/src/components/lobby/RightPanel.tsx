import { Lock, Gamepad2, ChevronRight } from "lucide-react"
import { rightModes } from "./data/lobby"

function ModeCard({
  title,
  subtitle,
  hint,
  className = "",
}: {
  title: string
  subtitle: string
  hint?: string
  className?: string
}) {
  return (
    <button
      type="button"
      className={`group relative overflow-hidden rounded-md bg-gradient-to-br from-panel-light to-panel ring-1 ring-white/10 transition hover:ring-gold/40 active:scale-[0.98] ${className}`}
    >
      {hint && (
        <span className="absolute right-2 top-2 z-10 flex items-center gap-1 text-[10px] text-muted">
          {hint}
          <Lock size={9} className="text-gold" />
        </span>
      )}
      <div className="absolute inset-0 flex flex-col justify-center px-3 text-left">
        <span className="font-display text-[9px] tracking-[0.2em] text-white/15">{subtitle}</span>
        <span className="text-xl font-black text-foreground/80">{title}</span>
      </div>
    </button>
  )
}

export default function RightPanel() {
  return (
    <div className="flex h-[500px] w-[450px] shrink-0 flex-col gap-3 p-3">
      {/* Main mission + daily task */}
      <div className="grid h-[70px] grid-cols-[1fr_auto] gap-3">
        <button
          type="button"
          className="relative overflow-hidden rounded-md bg-panel px-3 py-2 text-left ring-1 ring-white/10 transition hover:ring-gold/40 active:scale-[0.98]"
        >
          <div className="flex items-baseline gap-1 text-sm">
            <span className="font-bold text-emerald-400">[主线]</span>
            <span className="font-bold text-gold">董事会目标</span>
          </div>
          <p className="mt-1 text-sm text-foreground/80">继续赢得常规赛 1/2 场</p>
        </button>
        <button
          type="button"
          className="relative grid w-24 place-items-center rounded-md bg-panel px-2 ring-1 ring-white/10 transition hover:ring-gold/40 active:scale-[0.98]"
        >
          <span className="absolute right-1.5 top-1.5 flex items-center gap-0.5 text-[9px] text-muted">
            S1-25
            <Lock size={8} className="text-gold" />
          </span>
          <span className="text-sm font-bold text-foreground/80">日常任务</span>
        </button>
      </div>

      {/* Competitive (full width) */}
      <ModeCard title={rightModes[0].title} subtitle={rightModes[0].subtitle} hint={rightModes[0].hint} className="h-[128px]" />

      {/* Commercial + Season */}
      <div className="grid grid-cols-2 gap-3">
        <ModeCard title={rightModes[1].title} subtitle={rightModes[1].subtitle} hint={rightModes[1].hint} className="h-[100px]" />
        <ModeCard title={rightModes[2].title} subtitle={rightModes[2].subtitle} hint={rightModes[2].hint} className="h-[100px]" />
      </div>

      {/* START button */}
      <button
        type="button"
        className="group relative flex h-[140px] items-center justify-between overflow-hidden rounded-md bg-gradient-to-r from-[#1d428a] via-accent to-accent px-5 ring-1 ring-white/10 transition hover:brightness-110 active:scale-[0.98]"
      >
        <Gamepad2 size={48} className="text-white/90" />
        <div className="text-right">
          <span className="block text-3xl font-black text-white drop-shadow">比赛</span>
          <span className="flex items-center justify-end gap-1 font-display text-2xl font-bold italic tracking-wider text-white">
            START
            <ChevronRight size={20} strokeWidth={3} className="-ml-1 animate-pulse" />
          </span>
        </div>
      </button>
    </div>
  )
}

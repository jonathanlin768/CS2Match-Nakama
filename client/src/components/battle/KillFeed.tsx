import { Crosshair, Skull, Radio } from "lucide-react"
import type { KillEvent, Side } from "./data/battle"

/** 阵营文字配色：CT 蓝、T 金 */
const SIDE_TEXT: Record<Side, string> = {
  ct: "text-sky-300",
  t: "text-amber-300",
}

/**
 * 实时播报区：渲染最近的击杀流水（最新在前）。
 * 格式：击杀者 [武器/爆头] 被击杀者，双方按阵营着色。
 */
export default function KillFeed({ events }: { events: KillEvent[] }) {
  return (
    <div className="flex h-[156px] shrink-0 flex-col overflow-hidden rounded-md bg-panel/80 ring-1 ring-white/10">
      {/* 标题栏 */}
      <div className="flex items-center gap-2 border-b border-white/10 px-3 py-1.5">
        <Radio size={15} className="text-accent" />
        <span className="text-xs font-bold tracking-wide text-foreground/80">实时播报</span>
        <span className="ml-1 flex h-1.5 w-1.5 animate-pulse rounded-full bg-accent" />
      </div>

      {/* 击杀列表 */}
      <ul className="scrollbar-hide flex-1 space-y-0.5 overflow-y-auto px-3 py-1.5">
        {events.map((e) => (
          <li key={e.id} className="flex items-center gap-2 text-sm leading-6">
            <span className={`font-semibold ${SIDE_TEXT[e.killerSide]}`}>{e.killer}</span>
            <span className="flex items-center gap-1 text-muted">
              <Crosshair size={13} />
              <span className="text-xs">{e.weapon}</span>
              {e.headshot && <Skull size={13} className="text-red-400" />}
            </span>
            <span className={`opacity-60 ${SIDE_TEXT[e.victimSide]}`}>{e.victim}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

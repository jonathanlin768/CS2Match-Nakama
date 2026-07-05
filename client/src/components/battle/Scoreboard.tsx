import type { BattleTeam, Side } from "./data/battle"

/** 阵营文字配色：CT 蓝、T 金 */
const SIDE_TEXT: Record<Side, string> = {
  ct: "text-sky-300",
  t: "text-amber-300",
}
const SIDE_BADGE: Record<Side, string> = {
  ct: "bg-sky-500/15 text-sky-200 ring-sky-400/40",
  t: "bg-amber-500/15 text-amber-200 ring-amber-400/40",
}

function TeamBadge({ tag, side }: { tag: string; side: Side }) {
  return (
    <div
      className={`grid h-12 min-w-12 place-items-center rounded-md px-2 font-display text-base font-black tracking-wide ring-1 ${SIDE_BADGE[side]}`}
    >
      {tag}
    </div>
  )
}

export default function Scoreboard({
  teamA,
  teamB,
  round,
  maxRounds,
}: {
  teamA: BattleTeam
  teamB: BattleTeam
  round: number
  maxRounds: number
}) {
  return (
    <div className="flex shrink-0 items-center justify-center gap-6 px-[40px] py-2">
      {/* 左队：队名 + 队标 */}
      <div className="flex flex-1 items-center justify-end gap-3">
        <span className="truncate font-display text-2xl font-bold text-foreground">{teamA.name}</span>
        <TeamBadge tag={teamA.tag} side={teamA.side} />
      </div>

      {/* 比分 */}
      <div className="flex flex-col items-center">
        <div className="flex items-center gap-4 font-display text-4xl font-black tabular-nums leading-none">
          <span className={SIDE_TEXT[teamA.side]}>{teamA.score}</span>
          <span className="text-white/30">:</span>
          <span className={SIDE_TEXT[teamB.side]}>{teamB.score}</span>
        </div>
        <span className="mt-1 text-xs tracking-wide text-muted">
          第 {round} / {maxRounds} 回合
        </span>
      </div>

      {/* 右队：队标 + 队名 */}
      <div className="flex flex-1 items-center gap-3">
        <TeamBadge tag={teamB.tag} side={teamB.side} />
        <span className="truncate font-display text-2xl font-bold text-foreground">{teamB.name}</span>
      </div>
    </div>
  )
}

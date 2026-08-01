import { useState } from "react"
import { Radio } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../../legacy/components/ui/dialog"
import { Button } from "../../legacy/components/ui/button"
import type { GameEvent } from "../../types/match-report"
import type { FinalStats } from "../../types/match-report"
import type { BattleTeam } from "./data/battle"

interface EventFeedProps {
  events: GameEvent[]
  teamA: BattleTeam
  teamB: BattleTeam
  finalStats?: FinalStats
}

function StatsTable({ team, finalStats }: { team: BattleTeam; finalStats?: FinalStats }) {
  const rows = finalStats?.player_stats.filter((p) => team.players.some((player) => player.id === p.player_name)) ?? []
  return (
    <div className="rounded-md bg-panel/60 ring-1 ring-white/10">
      <div className="border-b border-white/10 px-3 py-2 text-sm font-semibold text-foreground">
        {team.name}
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr className="text-left text-xs text-muted-foreground">
            <th className="px-3 py-1.5 font-medium">选手</th>
            <th className="w-12 px-2 py-1.5 text-center font-medium">K</th>
            <th className="w-12 px-2 py-1.5 text-center font-medium">D</th>
            <th className="w-12 px-2 py-1.5 text-center font-medium">A</th>
          </tr>
        </thead>
        <tbody>
          {(rows.length > 0 ? rows : team.players).map((p) => {
            const playerName = "player_name" in p ? p.player_name : p.id
            const alive = "alive" in p ? p.alive : true
            return (
            <tr
              key={playerName}
              className={`border-t border-white/5 ${alive ? "text-foreground" : "text-foreground/40"}`}
            >
              <td className="px-3 py-1.5">{playerName}</td>
              <td className="px-2 py-1.5 text-center tabular-nums">{p.kills}</td>
              <td className="px-2 py-1.5 text-center tabular-nums">{p.deaths}</td>
              <td className="px-2 py-1.5 text-center tabular-nums">{"adr" in p ? p.adr.toFixed(1) : p.assists}</td>
            </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

/**
 * 服务器战报事件流：按时间顺序逐条展示 HLTV 风格事件。
 * 父组件通过控制传入的 events 长度实现“每 1 秒显示一条”的播放效果。
 */
export default function EventFeed({ events, teamA, teamB, finalStats }: EventFeedProps) {
  const [open, setOpen] = useState(false)

  return (
    <div className="flex h-[200px] shrink-0 flex-col overflow-hidden rounded-md bg-panel/80 ring-1 ring-white/10">
      {/* 标题栏 */}
      <div className="flex items-center justify-between border-b border-white/10 px-3 py-1.5">
        <div className="flex items-center gap-2">
          <Radio size={15} className="text-accent" />
          <span className="text-xs font-bold tracking-wide text-foreground/80">实时播报</span>
          {events.length > 0 && <span className="ml-1 flex h-1.5 w-1.5 animate-pulse rounded-full bg-accent" />}
        </div>

        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button type="button" size="sm" className="text-xs">
              数据统计
            </Button>
          </DialogTrigger>
          <DialogContent className="bg-background p-0 sm:max-w-2xl">
            <DialogHeader className="px-5 pt-5 pb-2">
              <DialogTitle className="text-base">数据统计</DialogTitle>
            </DialogHeader>
            <div className="max-h-[70vh] space-y-4 overflow-y-auto px-5 pb-5">
              <StatsTable team={teamA} finalStats={finalStats} />
              <StatsTable team={teamB} finalStats={finalStats} />
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* 事件列表：时间顺序，新事件追加在底部 */}
      <div className="scrollbar-hide flex-1 space-y-1 overflow-y-auto px-3 py-2 text-sm">
        {events.map((ev, idx) => (
          <div
            key={idx}
            className={`leading-5 ${
              ev.event_type === "ROUND_START" || ev.event_type === "ROUND_END"
                ? "font-semibold text-primary"
                : "text-foreground/90"
            }`}
          >
            {ev.message}
          </div>
        ))}
      </div>
    </div>
  )
}

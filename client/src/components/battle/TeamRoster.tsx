import PlayerCard from "./PlayerCard"
import type { BattleTeam } from "./data/battle"

/**
 * 单侧战队阵容：纵向排列 5 名选手卡，等分列高。
 * align=left 头像在左，align=right 头像在右（镜像，用于右侧战队）。
 */
export default function TeamRoster({
  team,
  align,
}: {
  team: BattleTeam
  align: "left" | "right"
}) {
  return (
    <div className="battle-roster flex h-full w-[clamp(260px,24vw,470px)] shrink-0 flex-col gap-2">
      {team.players.map((player) => (
        <div key={player.id} className="min-h-0 flex-1">
          <PlayerCard player={player} side={team.side} align={align} />
        </div>
      ))}
    </div>
  )
}

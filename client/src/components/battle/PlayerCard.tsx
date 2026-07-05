import {
  Shield,
  HardHat,
  Wrench,
  Crosshair,
  Skull,
  Heart,
  Zap,
  Cloud,
  Bomb,
  Flame,
  CircleDashed,
  type LucideIcon,
} from "lucide-react"
import type { BattlePlayer, GrenadeType, Side } from "./data/battle"

/** 阵营配色：CT 蓝、T 金 */
const SIDE_STYLE: Record<Side, { text: string; bar: string }> = {
  ct: { text: "text-sky-300", bar: "bg-sky-500" },
  t: { text: "text-amber-300", bar: "bg-amber-500" },
}

const GRENADE_ICON: Record<GrenadeType, LucideIcon> = {
  flash: Zap,
  smoke: Cloud,
  he: Bomb,
  molotov: Flame,
  decoy: CircleDashed,
}

const GRENADE_COLOR: Record<GrenadeType, string> = {
  flash: "text-yellow-300",
  smoke: "text-slate-300",
  he: "text-orange-400",
  molotov: "text-red-400",
  decoy: "text-slate-400",
}

/** 根据血量返回数字颜色 */
function hpColor(health: number) {
  if (health > 50) return "text-emerald-400"
  if (health > 25) return "text-amber-400"
  return "text-red-400"
}

export default function PlayerCard({
  player,
  side,
  align = "left",
}: {
  player: BattlePlayer
  side: Side
  /** 卡片朝向：左侧战队 left（头像在左），右侧战队 right（头像在右，镜像） */
  align?: "left" | "right"
}) {
  const sideStyle = SIDE_STYLE[side]
  const dead = !player.alive
  const mirror = align === "right"

  return (
    <div
      className={`relative flex h-full overflow-hidden rounded-md bg-panel ring-1 ring-white/10 ${
        mirror ? "flex-row-reverse" : ""
      } ${dead ? "opacity-55 saturate-50" : ""}`}
    >
      {/* 阵营色边条 */}
      <span className={`w-1 shrink-0 ${sideStyle.bar} ${dead ? "opacity-40" : ""}`} />

      {/* 头像 */}
      <div className="relative h-full w-[84px] shrink-0 overflow-hidden bg-black/40">
        <img
          src={player.avatar}
          alt={player.id}
          className={`h-full w-full scale-125 object-cover object-top ${dead ? "grayscale" : ""}`}
        />
        {dead && (
          <div className="absolute inset-0 grid place-items-center bg-black/50">
            <Skull size={28} className="text-muted" />
          </div>
        )}
      </div>

      {/* 信息主体 */}
      <div className={`flex flex-1 flex-col justify-between p-2 ${mirror ? "items-end text-right" : ""}`}>
        {/* 第一行：ID + K/D */}
        <div className={`flex w-full items-center ${mirror ? "flex-row-reverse" : ""} justify-between gap-2`}>
          <span className={`truncate font-display text-lg font-bold leading-none ${dead ? "text-muted" : "text-foreground"}`}>
            {player.id}
          </span>
          <div className={`flex items-center gap-2 text-xs ${mirror ? "flex-row-reverse" : ""}`}>
            <span className="flex items-center gap-0.5 text-foreground/80" title="击杀">
              <Crosshair size={12} className="text-emerald-400" />
              <span className="tabular-nums">{player.kills}</span>
            </span>
            <span className="flex items-center gap-0.5 text-foreground/80" title="死亡">
              <Skull size={12} className="text-red-400" />
              <span className="tabular-nums">{player.deaths}</span>
            </span>
          </div>
        </div>

        {/* 第二行：护甲 / 头甲 / 拆弹器 / 手雷 */}
        <div className={`flex items-center gap-1.5 ${mirror ? "flex-row-reverse" : ""}`}>
          <Shield
            size={15}
            className={player.armor > 0 ? sideStyle.text : "text-white/15"}
            title={`护甲 ${player.armor}`}
          />
          <HardHat size={15} className={player.helmet ? sideStyle.text : "text-white/15"} title="头甲" />
          {player.defuseKit && <Wrench size={15} className="text-sky-300" title="拆弹器" />}
          <span className="mx-0.5 h-3 w-px bg-white/10" />
          {player.grenades.length > 0 ? (
            player.grenades.map((g, i) => {
              const Icon = GRENADE_ICON[g]
              return <Icon key={`${g}-${i}`} size={15} className={GRENADE_COLOR[g]} />
            })
          ) : (
            <span className="text-[10px] text-white/20">无道具</span>
          )}
        </div>

        {/* 第三行：金钱 + 武器 */}
        <div className={`flex w-full items-center ${mirror ? "flex-row-reverse" : ""} justify-between gap-2`}>
          <span className="font-display text-sm font-semibold tabular-nums text-emerald-400">
            ${player.money.toLocaleString()}
          </span>
          <span className="truncate text-xs font-medium text-foreground/70">{player.weapon}</span>
        </div>
      </div>

      {/* 血量 */}
      <div className="flex w-[58px] shrink-0 flex-col items-center justify-center gap-0.5 bg-black/20">
        {dead ? (
          <span className="font-display text-2xl font-black text-muted">0</span>
        ) : (
          <>
            <Heart size={13} className="text-red-500" fill="currentColor" />
            <span className={`font-display text-2xl font-black leading-none tabular-nums ${hpColor(player.health)}`}>
              {player.health}
            </span>
          </>
        )}
      </div>
    </div>
  )
}

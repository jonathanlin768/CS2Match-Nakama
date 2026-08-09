import type { GameEvent } from "../../types/match-report"

interface MapViewProps {
  events?: GameEvent[]
}

/**
 * 地图区域：渲染 Dust2 雷达图，并在击杀/炸弹事件的位置显示标记。
 * 击杀坐标使用 location.x / location.y（0.0 ~ 1.0），按容器比例定位。
 */
export default function MapView({ events = [] }: MapViewProps) {
  const markerEvents = events.filter((ev) =>
    ["KILL", "BOMB_PLANT", "BOMB_DEFUSE", "BOMB_EXPLODE"].includes(ev.event_type) && ev.location
  )

  return (
    <div className="relative flex min-h-0 flex-1 overflow-hidden rounded-md bg-panel/50 ring-1 ring-white/10">
      {/* 雷达地图：object-contain 保证整张地图可见 */}
      <img
        src="/csmaps/de_dust2_radar_trans.webp"
        alt="de_dust2 radar"
        className="absolute inset-0 h-full w-full object-contain"
      />

      {/* 击杀标记层 */}
      <div className="pointer-events-none absolute inset-0">
        {markerEvents.map((ev, idx) => {
          const loc = ev.location!
          const isBomb = ev.event_type.startsWith("BOMB")
          return (
            <div
              key={`${ev.event_type}-${ev.victim_id ?? ev.timestamp}-${idx}`}
              className={`absolute flex h-7 w-7 items-center justify-center rounded-full text-sm font-black ring-1 ${
                isBomb ? "bg-amber-400/20 text-amber-200 ring-amber-300/60" : "bg-red-500/15 text-red-400 ring-red-400/50"
              }`}
              style={{
                left: `${loc.x * 100}%`,
                top: `${loc.y * 100}%`,
                transform: "translate(-50%, -50%)",
              }}
            >
              <span className="leading-none drop-shadow-[0_1px_2px_rgba(0,0,0,0.9)]">{isBomb ? "B" : "x"}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

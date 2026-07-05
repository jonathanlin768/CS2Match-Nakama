import type { GameEvent } from "../../types/match-report"

interface MapViewProps {
  events?: GameEvent[]
}

/**
 * 地图区域：渲染 Dust2 雷达图，并在击杀事件的位置显示红色 ×。
 * 击杀坐标使用 location.x / location.y（0.0 ~ 1.0），按容器比例定位。
 */
export default function MapView({ events = [] }: MapViewProps) {
  const killEvents = events.filter((ev) => ev.event_type === "KILL" && ev.location)

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
        {killEvents.map((ev, idx) => {
          const loc = ev.location!
          return (
            <div
              key={`${ev.victim_id}-${idx}`}
              className="absolute flex items-center justify-center text-red-500"
              style={{
                left: `${loc.x * 100}%`,
                top: `${loc.y * 100}%`,
                transform: "translate(-50%, -50%)",
              }}
            >
              <span className="text-2xl font-black leading-none drop-shadow-[0_1px_2px_rgba(0,0,0,0.9)]">
                ×
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

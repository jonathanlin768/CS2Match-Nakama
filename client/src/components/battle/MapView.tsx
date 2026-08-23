import { useEffect } from "react"
import type { GameEvent } from "../../types/match-report"
import { invalidMapMarkerEvents, normalizedToRadarPoint, RADAR_HEIGHT, RADAR_WIDTH, visibleMapMarkerEvents } from "./map-projection"

interface MapViewProps {
  events?: GameEvent[]
  className?: string
}

/**
 * 地图区域：渲染 Dust2 雷达图，并在击杀/炸弹事件的位置显示标记。
 * 击杀坐标使用 location.x / location.y（0.0 ~ 1.0），按容器比例定位。
 */
export default function MapView({ events = [], className = "" }: MapViewProps) {
  const markerEvents = visibleMapMarkerEvents(events)
  const invalidEvents = invalidMapMarkerEvents(events)

  useEffect(() => {
    if (!import.meta.env.DEV || invalidEvents.length === 0) return
    console.warn("[BattleMap] 已跳过非法归一化坐标事件", invalidEvents)
  }, [invalidEvents])

  return (
    <div className={`battle-map ${className}`} data-testid="battle-map">
      <svg viewBox={`0 0 ${RADAR_WIDTH} ${RADAR_HEIGHT}`} preserveAspectRatio="xMidYMid meet" role="img" aria-label="de_dust2 雷达图">
        <image href="/csmaps/de_dust2_radar_trans.webp" x="0" y="0" width={RADAR_WIDTH} height={RADAR_HEIGHT} preserveAspectRatio="none" />
        <g className="battle-map-markers" aria-label="当前回合事件标记">
          {markerEvents.map((event, index) => {
            const location = event.location!
            const point = normalizedToRadarPoint(location.x, location.y)!
            const isBomb = event.event_type.startsWith("BOMB")
            return (
              <g
                key={event.event_id ?? `${event.event_type}-${event.victim_id ?? event.timestamp}-${index}`}
                className={isBomb ? "battle-map-marker is-bomb" : "battle-map-marker is-kill"}
                transform={`translate(${point.x} ${point.y})`}
                data-testid="battle-map-marker"
                data-map-x={location.x}
                data-map-y={location.y}
              >
                <title>{isBomb ? `炸弹事件：${location.name}` : `击杀事件：${location.name}`}</title>
                <circle r="31" />
                <text textAnchor="middle" dominantBaseline="central">{isBomb ? "B" : "×"}</text>
              </g>
            )
          })}
        </g>
      </svg>
    </div>
  )
}

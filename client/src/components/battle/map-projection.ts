import type { GameEvent } from "../../types/match-report"

export const RADAR_WIDTH = 1024
export const RADAR_HEIGHT = 984

const MARKER_EVENT_TYPES = new Set(["KILL", "BOMB_PLANT", "BOMB_DEFUSE", "BOMB_EXPLODE"])

export interface RadarPoint {
  x: number
  y: number
}

export function normalizedToRadarPoint(x: number, y: number): RadarPoint | null {
  if (!Number.isFinite(x) || !Number.isFinite(y) || x < 0 || x > 1 || y < 0 || y > 1) return null
  return { x: x * RADAR_WIDTH, y: y * RADAR_HEIGHT }
}

export function visibleMapMarkerEvents(events: GameEvent[]): GameEvent[] {
  return events.filter((event) => {
    if (!MARKER_EVENT_TYPES.has(event.event_type) || !event.location) return false
    return normalizedToRadarPoint(event.location.x, event.location.y) !== null
  })
}

export function invalidMapMarkerEvents(events: GameEvent[]): GameEvent[] {
  return events.filter((event) => (
    MARKER_EVENT_TYPES.has(event.event_type)
    && Boolean(event.location)
    && normalizedToRadarPoint(event.location!.x, event.location!.y) === null
  ))
}

import type { BombPublicState } from "../../types/match-report"

const LABELS: Record<string, string> = {
  Carried: "携带中",
  Dropped: "已掉落",
  Planted: "已下包",
  Defused: "已拆除",
  Exploded: "已爆炸",
}

export default function BombStatusStrip({ bomb, compact = false }: { bomb?: BombPublicState; compact?: boolean }) {
  const status = bomb?.status ?? "Carried"
  return (
    <div className={`battle-bomb-strip ${compact ? "is-compact" : ""}`} data-testid="bomb-status">
      <span className="battle-bomb-label">炸弹</span>
      <strong>{LABELS[status] ?? status}</strong>
      <span className="battle-bomb-location">{bomb?.site ? `${bomb.site} 区` : bomb?.node_id ?? "未下包"}</span>
    </div>
  )
}

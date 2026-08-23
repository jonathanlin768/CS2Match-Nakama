import { Pause, Play, SkipForward } from "lucide-react"
import type { BattleViewModel } from "./battle-view-model"

export default function BattleReplayControls({ model, compact = false }: { model: BattleViewModel; compact?: boolean }) {
  return (
    <div className={`battle-replay-controls ${compact ? "is-compact" : ""}`} data-testid="battle-replay-controls">
      <button type="button" onClick={model.onTogglePlaying} className="battle-icon-control" title={model.playing ? "暂停" : "继续"} aria-label={model.playing ? "暂停" : "继续"}>
        {model.playing ? <Pause size={18} /> : <Play size={18} />}
      </button>
      {!model.fullMatchRevealed && (
        <button type="button" onClick={model.onSkipMatch} className="battle-skip-control" aria-label="跳过比赛" title="跳过比赛">
          <SkipForward size={17} /><span>跳过比赛</span>
        </button>
      )}
      <div className="battle-round-scroll" aria-label="回合导航">
        <div className="battle-round-list">
          {model.visibleRounds.map((round, index) => (
            <button
              key={round.round_number}
              type="button"
              onClick={() => model.onSelectRound(index)}
              className={`battle-round-button ${index === model.roundIndex ? "is-current" : ""} ${round.phase === "overtime" ? "is-overtime" : ""}`}
              aria-current={index === model.roundIndex ? "step" : undefined}
            >
              {round.round_number}
            </button>
          ))}
        </div>
      </div>
      {!compact && <div className="battle-seed">Seed {model.report.match_info.seed} · {model.report.match_info.rule_set_id}</div>}
    </div>
  )
}

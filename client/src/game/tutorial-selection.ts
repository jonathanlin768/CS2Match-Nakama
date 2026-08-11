import type { TutorialBattle } from "../config"

export function tutorialPlayerPrice(config: TutorialBattle, playerId: string) {
  const pools = [config.tier1PlayerIds, config.tier2PlayerIds, config.tier3PlayerIds, config.tier4PlayerIds, config.tier5PlayerIds]
  return pools.findIndex((pool) => pool.includes(playerId)) + 1
}

export function tutorialSelectionCost(config: TutorialBattle, selected: string[]) {
  return selected.reduce((sum, id) => sum + tutorialPlayerPrice(config, id), 0)
}

export function toggleTutorialPlayer(config: TutorialBattle, selected: string[], playerId: string): { selected: string[]; error?: "unknown" | "duplicate" | "full" | "budget" } {
  if (selected.includes(playerId)) return { selected: selected.filter((id) => id !== playerId) }
  const price = tutorialPlayerPrice(config, playerId)
  if (price <= 0) return { selected, error: "unknown" }
  if (selected.length >= config.rosterSize) return { selected, error: "full" }
  if (tutorialSelectionCost(config, selected) + price > config.budget) return { selected, error: "budget" }
  return { selected: [...selected, playerId] }
}

export function tutorialSelectionReady(config: TutorialBattle, selected: string[]) {
  return new Set(selected).size === config.rosterSize && selected.length === config.rosterSize && tutorialSelectionCost(config, selected) <= config.budget
}

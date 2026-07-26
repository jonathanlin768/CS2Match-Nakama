import type { SelectedObject } from './model'

export type LineSelection = { kind: 'edge' | 'route'; id: string }

export function nextLineSelection(candidates: LineSelection[], selected: SelectedObject) {
  if (candidates.length === 0) return null
  const currentIndex = selected ? candidates.findIndex((candidate) => candidate.kind === selected.kind && candidate.id === selected.id) : -1
  return candidates[(currentIndex + 1) % candidates.length]
}

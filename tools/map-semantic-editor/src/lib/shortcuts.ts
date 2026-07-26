import type { ToolMode } from './model'

const toolShortcutMap: Record<string, ToolMode> = {
  s: 'select',
  n: 'node',
  e: 'edge',
  v: 'visibility',
  l: 'route',
  r: 'risk',
}

export function toolForShortcutKey(key: string): ToolMode | null {
  return toolShortcutMap[key.toLowerCase()] ?? null
}

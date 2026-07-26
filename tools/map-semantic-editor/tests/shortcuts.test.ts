import { describe, expect, it } from 'vitest'
import { toolForShortcutKey } from '../src/lib/shortcuts'

describe('tool keyboard shortcuts', () => {
  it('maps single letter keys to toolbar tools', () => {
    expect(toolForShortcutKey('s')).toBe('select')
    expect(toolForShortcutKey('N')).toBe('node')
    expect(toolForShortcutKey('e')).toBe('edge')
    expect(toolForShortcutKey('V')).toBe('visibility')
    expect(toolForShortcutKey('l')).toBe('route')
    expect(toolForShortcutKey('R')).toBe('risk')
  })

  it('ignores unrelated keys', () => {
    expect(toolForShortcutKey('x')).toBeNull()
    expect(toolForShortcutKey('Enter')).toBeNull()
  })
})

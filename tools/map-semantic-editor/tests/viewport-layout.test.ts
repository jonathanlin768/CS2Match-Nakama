import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

describe('viewport layout constraints', () => {
  const root = resolve(process.cwd())
  const appSource = readFileSync(resolve(root, 'src/pages/MapConfigPage.tsx'), 'utf8')
  const styles = readFileSync(resolve(root, 'src/styles.css'), 'utf8')

  it('keeps the default workspace columns within the editor minimum viewport', () => {
    const leftPanel = numberFrom(appSource, /useState\((\d+)\)/, 0)
    const propertyPanel = numberFrom(appSource, /useState\((\d+)\)/, 1)
    const canvasMin = numberFrom(appSource, /minmax\((\d+)px,\s*1fr\)/, 0)
    const handleWidth = 6
    const workspaceMin = leftPanel + handleWidth + canvasMin + handleWidth + propertyPanel
    const bodyMin = numberFrom(styles, /min-width:\s*(\d+)px/, 0)

    expect(workspaceMin).toBeLessThanOrEqual(bodyMin)
    expect(bodyMin).toBeLessThanOrEqual(1280)
  })

  it('keeps toolbar labels from overlapping in narrow viewports', () => {
    expect(styles).toMatch(/white-space:\s*nowrap/)
    expect(styles).toMatch(/\.titleStrip,\s*\n\.toolStrip\s*{[^}]*overflow-x:\s*auto/s)
    expect(styles).toMatch(/\.commandButton,\s*\n\.toolButton,/)
  })
})

function numberFrom(source: string, pattern: RegExp, matchIndex: number): number {
  const matches = [...source.matchAll(new RegExp(pattern.source, `${pattern.flags}g`))]
  const value = matches[matchIndex]?.[1]
  if (!value) throw new Error(`Pattern not found: ${pattern}`)
  return Number(value)
}

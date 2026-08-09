import path from 'node:path'
import fs from 'node:fs/promises'
import os from 'node:os'
import { fileURLToPath } from 'node:url'
import ExcelJS from 'exceljs'
import { afterEach, describe, expect, it } from 'vitest'
import { toExportTables } from '../src/lib/exportTables'
import type { MapNode, MapProject } from '../src/lib/model'
import { createSampleProject } from '../src/lib/sampleProject'
import { writeExportTables } from '../server/excel'
import { importLubanProject, readLubanWorkbook } from '../server/importTables'

const testDir = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(testDir, '..', '..', '..')

let tempRoots: string[] = []

afterEach(async () => {
  await Promise.all(tempRoots.map((root) => fs.rm(root, { recursive: true, force: true })))
  tempRoots = []
})

async function tempDir(): Promise<string> {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'map-import-'))
  tempRoots.push(dir)
  return dir
}

describe('importLubanProject', () => {
  it('round-trips every table through xlsx', async () => {
    const project = createSampleProject()
    const root = await tempDir()
    await writeExportTables(toExportTables(project), root)

    const result = await importLubanProject(project, root)

    expect(result.project.nodes).toEqual(project.nodes)
    expect(result.project.edges).toEqual(project.edges)
    expect(result.project.visibility).toEqual(project.visibility)
    expect(result.project.routes).toEqual(project.routes)
    expect(result.project.route_templates).toEqual(project.route_templates)
    expect(result.project.scenarios).toEqual(project.scenarios)
    expect(result.project.map_tags).toEqual(project.map_tags)
    expect(result.project.encounter_modifiers).toEqual(project.encounter_modifiers)
    expect(result.project.combat_consts).toEqual(project.combat_consts)
    expect(result.warnings).toEqual([])
    expect(result.summary.find((entry) => entry.file === '#MapNode.xlsx')?.rows).toBe(project.nodes.length)
    expect(result.summary.find((entry) => entry.file === '#Route.xlsx')?.rows).toBe(project.routes.length)
  })

  it('keeps project metadata and replaces data arrays', async () => {
    const project = createSampleProject()
    const root = await tempDir()
    await writeExportTables(toExportTables(project), root)

    const result = await importLubanProject(project, root)

    expect(result.project.map_id).toBe(project.map_id)
    expect(result.project.name).toBe(project.name)
    expect(result.project.radar_image).toBe(project.radar_image)
    expect(result.project.layers).toEqual(project.layers)
    expect(result.project.viewport).toEqual(project.viewport)
  })

  it('filters rows by current map_id', async () => {
    const project = createSampleProject()
    const source = structuredClone(project) as MapProject
    source.nodes.push({
      ...(project.nodes[0] as MapNode),
      id: 'MIRAGE_NODE',
      name: 'Mirage Node',
      map_id: 'de_mirage',
    })
    const root = await tempDir()
    await writeExportTables(toExportTables(source), root)

    const result = await importLubanProject(project, root)

    expect(result.project.nodes.find((node) => node.id === 'MIRAGE_NODE')).toBeUndefined()
    expect(result.project.nodes).toHaveLength(project.nodes.length)
  })

  it('coerces unknown enum values with a warning', async () => {
    const project = createSampleProject()
    const source = structuredClone(project) as MapProject
    ;(source.nodes[0] as unknown as { site: string }).site = 'Z'
    const root = await tempDir()
    await writeExportTables(toExportTables(source), root)

    const result = await importLubanProject(project, root)

    expect(result.project.nodes[0].site).toBe('None')
    expect(result.warnings.some((warning) => warning.includes('site'))).toBe(true)
  })

  it('accepts string booleans in bool columns', async () => {
    const project = createSampleProject()
    const root = await tempDir()
    await writeExportTables(toExportTables(project), root)

    const file = path.join(root, '#MapEdge.xlsx')
    const workbook = new ExcelJS.Workbook()
    await workbook.xlsx.readFile(file)
    const sheet = workbook.worksheets[0]
    let bidirectionalColumn = 0
    sheet.getRow(1).eachCell((cell, column) => {
      if (cell.value === 'bidirectional') bidirectionalColumn = column
    })
    expect(bidirectionalColumn).toBeGreaterThan(0)
    sheet.getRow(5).getCell(bidirectionalColumn).value = 'true'
    await workbook.xlsx.writeFile(file)

    const result = await importLubanProject(project, root)

    expect(result.project.edges[0].bidirectional).toBe(true)
    expect(result.warnings).toEqual([])
  })

  it('covers every real Luban table column so write-back does not drop data', async () => {
    const tables = toExportTables(createSampleProject())
    for (const table of tables) {
      const workbook = await readLubanWorkbook(path.join(repoRoot, 'configs', 'Datas', table.fileName))
      const sheet = workbook.worksheets[0]
      const headers: string[] = []
      sheet.getRow(1).eachCell((cell, column) => {
        if (column > 1 && cell.value !== null && cell.value !== undefined && String(cell.value).trim() !== '') {
          headers.push(String(cell.value))
        }
      })
      const exported = new Set(table.fields.map((field) => field.key))
      expect(headers.filter((header) => !exported.has(header))).toEqual([])
    }
  })

  it('imports the real Luban tables without warnings', async () => {
    const result = await importLubanProject(createSampleProject())
    expect(result.warnings).toEqual([])
    expect(result.project.nodes.length).toBeGreaterThan(0)
    expect(result.project.edges.length).toBeGreaterThan(0)
    expect(result.project.routes.length).toBeGreaterThan(0)
  })
})

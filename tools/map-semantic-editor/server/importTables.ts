import fs from 'node:fs/promises'
import ExcelJS from 'exceljs'
import JSZip from 'jszip'
import { LUBAN_TABLE_SPECS, type ImportTableKey } from '../src/lib/exportTables'
import {
  cellToPoints,
  floors,
  nodeShapes,
  nodeTypes,
  nodeUsages,
  parseProject,
  sides,
  sites,
  type CombatConst,
  type EncounterModifier,
  type MapEdge,
  type MapNode,
  type MapProject,
  type MapTag,
  type Point,
  type Route,
  type RouteTemplate,
  type Scenario,
  type Visibility,
} from '../src/lib/model'
import { dataFilePath, dataRoot } from './paths'

const ranges = ['Close', 'Mid', 'Long'] as const
const angleAdvantages = ['T', 'CT', 'None'] as const
const elevations = ['HighToLow', 'LowToHigh', 'SameLevel', 'Same', 'HeightBlocked'] as const
const tempos = ['Fast', 'Default', 'Slow', 'Late'] as const
const valueTypes = ['Int', 'Float', 'Bool', 'String'] as const
const routeSides = ['T', 'CT'] as const

export interface ImportSummaryEntry {
  tableName: string
  file: string
  rows: number
  warnings: string[]
}

export interface ImportResult {
  project: MapProject
  summary: ImportSummaryEntry[]
  warnings: string[]
}

export async function importLubanProject(project: MapProject, root = dataRoot): Promise<ImportResult> {
  const next = structuredClone(project)
  const warnings: string[] = []
  const summary: ImportSummaryEntry[] = []

  for (const spec of LUBAN_TABLE_SPECS) {
    const { rows, warnings: fileWarnings } = await readTableRows(spec.fileName, root)
    const { items, warnings: tableWarnings } = convertRows(spec.key, rows, next.map_id)
    const tableWarningsAll = [
      ...fileWarnings.map((item) => `${spec.fileName}: ${item}`),
      ...tableWarnings.map((item) => `${spec.fileName}: ${item}`),
    ]
    warnings.push(...tableWarningsAll)
    assignProjectArray(next, spec.key, items)
    summary.push({ tableName: spec.tableName, file: spec.fileName, rows: items.length, warnings: tableWarningsAll })
  }

  let parsed: MapProject
  try {
    parsed = parseProject(next)
  } catch (error) {
    throw new Error(`读取 Excel 配置后工程校验失败：${error instanceof Error ? error.message : String(error)}`, { cause: error })
  }
  return { project: parsed, summary, warnings }
}

async function readTableRows(
  fileName: string,
  root: string,
): Promise<{ fields: string[]; types: string[]; rows: Record<string, unknown>[]; warnings: string[] }> {
  const workbook = await readLubanWorkbook(dataFilePath(fileName, root))
  const sheet = workbook.worksheets[0]

  const fields: string[] = []
  const types: string[] = []
  const varRow = sheet.getRow(1)
  const typeRow = sheet.getRow(2)
  const columnCount = sheet.columnCount
  for (let column = 2; column <= columnCount; column += 1) {
    const key = cellText(varRow.getCell(column).value)
    if (!key) continue
    fields.push(key)
    types.push(cellText(typeRow.getCell(column).value))
  }

  const rows: Record<string, unknown>[] = []
  const warnings: string[] = []
  for (let rowNumber = 5; rowNumber <= sheet.rowCount; rowNumber += 1) {
    const row = sheet.getRow(rowNumber)
    const marker = cellText(row.getCell(1).value)
    if (marker.startsWith('##')) continue
    const record: Record<string, unknown> = {}
    let empty = true
    for (let index = 0; index < fields.length; index += 1) {
      const raw = row.getCell(index + 2).value
      if (!isBlank(raw)) empty = false
      record[fields[index]] = convertCell(raw, types[index], warnings, `${fields[index]}`)
    }
    if (!empty) rows.push(record)
  }
  return { fields, types, rows, warnings }
}

export async function readLubanWorkbook(filePath: string): Promise<ExcelJS.Workbook> {
  const data = await fs.readFile(filePath)
  const zip = await JSZip.loadAsync(data)
  for (const name of Object.keys(zip.files)) {
    const entry = zip.files[name]
    if (entry.dir || (!name.endsWith('.xml') && !name.endsWith('.rels'))) continue
    const content = await entry.async('string')
    if (!content.includes('<')) continue
    const normalized = content
      .replace(/<(\/?)x:/g, '<$1')
      .replace(/\sxmlns:x="[^"]*"/g, ' ')
    zip.file(name, normalized)
  }
  const buffer = await zip.generateAsync({ type: 'nodebuffer' })
  const workbook = new ExcelJS.Workbook()
  // jszip 与 exceljs 各自的 Buffer 泛型声明不一致，运行时类型相同
  await workbook.xlsx.load(buffer as never)
  return workbook
}

function convertCell(raw: unknown, type: string, warnings: string[], label: string): unknown {
  if (isBlank(raw)) return null
  if (typeof raw === 'number' || typeof raw === 'boolean') return raw
  const text = String(raw).trim()
  if (type === 'int') {
    const number = Number(text)
    if (Number.isFinite(number)) return Math.trunc(number)
    warnings.push(`${label}: 无法解析整数 "${text}"`)
    return text
  }
  if (type === 'float') {
    const number = Number(text)
    if (Number.isFinite(number)) return number
    warnings.push(`${label}: 无法解析数字 "${text}"`)
    return text
  }
  if (type === 'bool') {
    if (text === 'true' || text === '1') return true
    if (text === 'false' || text === '0') return false
    warnings.push(`${label}: 无法解析布尔值 "${text}"`)
    return text
  }
  if (type.startsWith('(list')) {
    return text.split(',').map((item) => item.trim()).filter(Boolean)
  }
  return text
}

function convertRows(
  key: ImportTableKey,
  rows: Record<string, unknown>[],
  mapId: string,
): { items: unknown[]; warnings: string[] } {
  const warnings: string[] = []
  let items: unknown[] = []
  switch (key) {
    case 'route_templates':
      items = rows.filter(byMapId(mapId)).map((record) => toRouteTemplate(record, mapId, warnings)).filter(nonNull)
      break
    case 'scenarios':
      items = rows.map((record) => toScenario(record, warnings)).filter(nonNull)
      break
    case 'map_tags':
      items = rows.filter(byMapId(mapId)).map((record) => toMapTag(record, mapId, warnings)).filter(nonNull)
      break
    case 'encounter_modifiers':
      items = rows.map((record) => toEncounterModifier(record, warnings)).filter(nonNull)
      break
    case 'nodes':
      items = rows.filter(byMapId(mapId)).map((record) => toMapNode(record, mapId, warnings)).filter(nonNull)
      break
    case 'edges':
      items = rows.map((record) => toMapEdge(record, warnings)).filter(nonNull)
      break
    case 'visibility':
      items = rows.map((record) => toVisibility(record, warnings)).filter(nonNull)
      break
    case 'routes':
      items = rows.map((record) => toRoute(record, warnings)).filter(nonNull)
      break
    case 'combat_consts':
      items = rows.map((record) => toCombatConst(record, warnings)).filter(nonNull)
      break
  }
  return { items, warnings }
}

function assignProjectArray(project: MapProject, key: ImportTableKey, items: unknown[]): void {
  switch (key) {
    case 'route_templates':
      project.route_templates = items as RouteTemplate[]
      break
    case 'scenarios':
      project.scenarios = items as Scenario[]
      break
    case 'map_tags':
      project.map_tags = items as MapTag[]
      break
    case 'encounter_modifiers':
      project.encounter_modifiers = items as EncounterModifier[]
      break
    case 'nodes':
      project.nodes = items as MapNode[]
      break
    case 'edges':
      project.edges = items as MapEdge[]
      break
    case 'visibility':
      project.visibility = items as Visibility[]
      break
    case 'routes':
      project.routes = items as Route[]
      break
    case 'combat_consts':
      project.combat_consts = items as CombatConst[]
      break
  }
}

function toMapNode(record: Record<string, unknown>, mapId: string, warnings: string[]): MapNode | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  const x = clamp01(num(record.x, 0, warnings, `${id}.x`), warnings, `${id}.x`)
  const y = clamp01(num(record.y, 0, warnings, `${id}.y`), warnings, `${id}.y`)
  const radius = nullableNum(record.radius, warnings, `${id}.radius`)
  return {
    id,
    map_id: str(record.map_id, mapId),
    name: str(record.name, id),
    zone: str(record.zone),
    site: enumOf(record.site, sites, 'None', warnings, `${id}.site`),
    node_type: enumOf(record.node_type, nodeTypes, 'Lane', warnings, `${id}.node_type`),
    default_side: enumOf(record.default_side, sides, 'None', warnings, `${id}.default_side`),
    x,
    y,
    floor: enumOf(record.floor, floors, 'Ground', warnings, `${id}.floor`),
    area_usages: usagesOf(record.area_usages, warnings, `${id}.area_usages`),
    shape: enumOf(record.shape, nodeShapes, 'None', warnings, `${id}.shape`),
    radius: radius === null ? null : clamp01(radius, warnings, `${id}.radius`),
    points: safePoints(record.points, warnings, `${id}.points`),
  }
}

function toMapEdge(record: Record<string, unknown>, warnings: string[]): MapEdge | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  return {
    id,
    from: str(record.from_node),
    to: str(record.to_node),
    base_time: intNum(record.base_time, 8, warnings, `${id}.base_time`),
    stamina_cost: intNum(record.stamina_cost, 5, warnings, `${id}.stamina_cost`),
    risk: intNum(record.risk, 10, warnings, `${id}.risk`),
    noise: intNum(record.noise, 10, warnings, `${id}.noise`),
    risk_points: listOf(record.risk_points),
    intercept_nodes: listOf(record.intercept_nodes),
    bidirectional: boolOf(record.bidirectional, true, warnings, `${id}.bidirectional`),
  }
}

function toVisibility(record: Record<string, unknown>, warnings: string[]): Visibility | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  return {
    id,
    from: str(record.from_node),
    to: str(record.to_node),
    visible: boolOf(record.visible, true, warnings, `${id}.visible`),
    range: enumOf(record.range, ranges, 'Mid', warnings, `${id}.range`),
    angle_advantage: enumOf(record.angle_advantage, angleAdvantages, 'None', warnings, `${id}.angle_advantage`),
    elevation: enumOf(record.elevation, elevations, 'SameLevel', warnings, `${id}.elevation`),
    cover_modifier: intNum(record.cover_modifier, 0, warnings, `${id}.cover_modifier`),
    exposure_modifier: intNum(record.exposure_modifier, 0, warnings, `${id}.exposure_modifier`),
  }
}

function toRoute(record: Record<string, unknown>, warnings: string[]): Route | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  return {
    id,
    name: str(record.name, id),
    side: enumOf(record.side, routeSides, 'T', warnings, `${id}.side`),
    target_site: enumOf(record.target_site, sites, 'None', warnings, `${id}.target_site`),
    nodes: listOf(record.nodes),
    min_players: intNum(record.min_players, 1, warnings, `${id}.min_players`),
    max_players: intNum(record.max_players, 5, warnings, `${id}.max_players`),
    style_tags: listOf(record.style_tags),
  }
}

function toRouteTemplate(record: Record<string, unknown>, mapId: string, warnings: string[]): RouteTemplate | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  return {
    id,
    map_id: str(record.map_id, mapId),
    side: enumOf(record.side, routeSides, 'T', warnings, `${id}.side`),
    target_site: enumOf(record.target_site, sites, 'None', warnings, `${id}.target_site`),
    tempo: enumOf(record.tempo, tempos, 'Default', warnings, `${id}.tempo`),
    recommended_min: intNum(record.recommended_min, 1, warnings, `${id}.recommended_min`),
    recommended_max: intNum(record.recommended_max, 5, warnings, `${id}.recommended_max`),
    required_roles: listOf(record.required_roles),
    key_attributes: str(record.key_attributes),
    route_ids: listOf(record.route_ids),
    route_allocations: allocationsOf(record.route_allocations, warnings, `${id}.route_allocations`),
    scenario_ids: listOf(record.scenario_ids),
    map_tag_ids: listOf(record.map_tag_ids),
    common_ct_setup_ids: listOf(record.common_ct_setup_ids),
    success_next_phase: str(record.success_next_phase),
    failure_fallbacks: listOf(record.failure_fallbacks),
  }
}

function toScenario(record: Record<string, unknown>, warnings: string[]): Scenario | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  return {
    id,
    route: str(record.route, 'A_Long'),
    phase: str(record.phase, 'OpeningDuel'),
    range: enumOf(record.range, ranges, 'Mid', warnings, `${id}.range`),
    site: enumOf(record.site, sites, 'None', warnings, `${id}.site`),
    tempo: str(record.tempo, 'SlowDefault'),
    posture: str(record.posture, 'Even'),
    utility_context: str(record.utility_context, 'Even'),
    map_tag_ids: listOf(record.map_tag_ids),
    base_time_cost: intNum(record.base_time_cost, 0, warnings, `${id}.base_time_cost`),
    base_weight: intNum(record.base_weight, 0, warnings, `${id}.base_weight`),
  }
}

function toMapTag(record: Record<string, unknown>, mapId: string, warnings: string[]): MapTag | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  return {
    id,
    map_id: str(record.map_id, mapId),
    category: str(record.category, 'Range'),
    value: str(record.value),
    side: enumOf(record.side, sides, 'Both', warnings, `${id}.side`),
    weight: intNum(record.weight, 0, warnings, `${id}.weight`),
    reason_code: str(record.reason_code),
    description: str(record.description),
  }
}

function toEncounterModifier(record: Record<string, unknown>, warnings: string[]): EncounterModifier | null {
  const id = str(record.id)
  if (!id) {
    warnings.push('缺少 id，已跳过该行')
    return null
  }
  return {
    id,
    scenario_id: str(record.scenario_id),
    factor: str(record.factor),
    side: enumOf(record.side, sides, 'Both', warnings, `${id}.side`),
    attribute: str(record.attribute),
    weight: intNum(record.weight, 0, warnings, `${id}.weight`),
    reason_code: str(record.reason_code),
  }
}

function toCombatConst(record: Record<string, unknown>, warnings: string[]): CombatConst | null {
  const key = str(record.key)
  if (!key) {
    warnings.push('缺少 key，已跳过该行')
    return null
  }
  return {
    key,
    category: str(record.category, 'Decision'),
    value_type: enumOf(record.value_type, valueTypes, 'Int', warnings, `${key}.value_type`),
    value: str(record.value, '0'),
    min_value: str(record.min_value),
    max_value: str(record.max_value),
    unit: str(record.unit, 'none'),
    description: str(record.description),
  }
}

function byMapId(mapId: string): (record: Record<string, unknown>) => boolean {
  return (record) => isBlank(record.map_id) || cellText(record.map_id) === mapId
}

function nonNull<T>(value: T | null): value is T {
  return value !== null
}

function cellText(value: unknown): string {
  if (value === null || value === undefined) return ''
  return String(value).trim()
}

function isBlank(value: unknown): boolean {
  return value === null || value === undefined || String(value).trim() === ''
}

function str(value: unknown, fallback = ''): string {
  return cellText(value) || fallback
}

function num(value: unknown, fallback: number, warnings: string[], label: string): number {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  const text = cellText(value)
  if (text !== '') {
    const number = Number(text)
    if (Number.isFinite(number)) return number
  }
  warnings.push(`${label}: 无法解析数字，使用 ${fallback}`)
  return fallback
}

function intNum(value: unknown, fallback: number, warnings: string[], label: string): number {
  return Math.trunc(num(value, fallback, warnings, label))
}

function nullableNum(value: unknown, warnings: string[], label: string): number | null {
  if (isBlank(value)) return null
  return num(value, 0, warnings, label)
}

function boolOf(value: unknown, fallback: boolean, warnings: string[], label: string): boolean {
  if (typeof value === 'boolean') return value
  const text = cellText(value).toLowerCase()
  if (text === 'true' || text === '1') return true
  if (text === 'false' || text === '0') return false
  warnings.push(`${label}: 无法解析布尔值，使用 ${fallback}`)
  return fallback
}

function listOf(value: unknown): string[] {
  if (Array.isArray(value)) return value.map((item) => cellText(item)).filter(Boolean)
  const text = cellText(value)
  if (!text) return []
  return text.split(',').map((item) => item.trim()).filter(Boolean)
}

function allocationsOf(value: unknown, warnings: string[], label: string): Record<string, number> {
  const text = cellText(value)
  if (!text) return {}
  const items = text.split(',')
  const out: Record<string, number> = {}
  for (let index = 0; index + 1 < items.length; index += 2) {
    const key = items[index].trim()
    const count = Number(items[index + 1].trim())
    if (key && Number.isFinite(count) && count >= 0) {
      out[key] = Math.trunc(count)
    } else {
      warnings.push(`${label}: 无法解析分配 "${items[index]},${items[index + 1]}"`)
    }
  }
  if (items.length % 2 !== 0) {
    warnings.push(`${label}: 分配条目数量为奇数，已忽略 "${items[items.length - 1]}"`)
  }
  return out
}

function usagesOf(value: unknown, warnings: string[], label: string): Array<(typeof nodeUsages)[number]> {
  const out: Array<(typeof nodeUsages)[number]> = []
  for (const item of listOf(value)) {
    if ((nodeUsages as readonly string[]).includes(item)) {
      out.push(item as (typeof nodeUsages)[number])
    } else {
      warnings.push(`${label}: 无法识别的用途 "${item}"，已忽略`)
    }
  }
  return out
}

function safePoints(value: unknown, warnings: string[], label: string): Point[] {
  const text = cellText(value)
  if (!text) return []
  try {
    return cellToPoints(text)
  } catch {
    warnings.push(`${label}: 非法点位格式 "${text}"，已置空`)
    return []
  }
}

function clamp01(value: number, warnings: string[], label: string): number {
  if (value < 0 || value > 1) {
    warnings.push(`${label}: 超出 0..1，已截断为 ${Math.min(1, Math.max(0, value))}`)
  }
  return Math.min(1, Math.max(0, Math.round(value * 10000) / 10000))
}

function enumOf<T extends readonly string[]>(
  value: unknown,
  allowed: T,
  fallback: T[number],
  warnings: string[],
  label: string,
): T[number] {
  const text = cellText(value)
  if ((allowed as readonly string[]).includes(text)) return text as T[number]
  if (text) warnings.push(`${label}: 无法识别的值 "${text}"，已回退为 "${fallback}"`)
  return fallback
}

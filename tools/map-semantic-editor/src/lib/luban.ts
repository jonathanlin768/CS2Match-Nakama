export type LubanCellValue = string | number | boolean | null | string[]

export type LubanFieldKind = 'string' | 'int' | 'float' | 'bool' | 'list' | 'ref' | 'unknown'

export type LubanTableOwner = 'map' | 'player' | 'team' | 'tutorial' | 'generic'

export interface WorkbookVersion {
  mtimeMs: number
  size: number
  hash: string
}

export interface WorkbookMeta {
  sheetName: string
  version: WorkbookVersion
}

export interface LubanField {
  key: string
  type: string
  group: string
  comment: string
  kind: LubanFieldKind
  refTable?: string
  list: boolean
  column: number
}

export type LubanRow = Record<string, LubanCellValue>

export interface LubanTableDocument {
  fileName: string
  tableName: string
  owner: LubanTableOwner
  editable: boolean
  fields: LubanField[]
  rows: LubanRow[]
  originalMeta: WorkbookMeta
  warnings: string[]
  dirty: boolean
}

export interface LubanTableSummary {
  fileName: string
  tableName: string
  owner: LubanTableOwner
  editable: boolean
  rows: number
  columns: number
  status: 'ready' | 'error'
  warnings: string[]
}

export interface ConfigValidationIssue {
  severity: 'ERROR' | 'WARNING'
  fileName: string
  rowIndex: number
  rowId: string
  field: string
  message: string
}

export interface ConfigReference {
  sourceFile: string
  sourceTable: string
  sourceRowId: string
  field: string
  targetTable: string
  targetId: string
}

export const mapOwnedFiles = new Set([
  '#RouteTemplate.xlsx',
  '#Scenario.xlsx',
  '#MapTag.xlsx',
  '#EncounterModifier.xlsx',
  '#MapNode.xlsx',
  '#MapEdge.xlsx',
  '#Visibility.xlsx',
  '#Route.xlsx',
  '#CombatConst.xlsx',
])

export function ownerForFile(fileName: string): LubanTableOwner {
  if (mapOwnedFiles.has(fileName)) return 'map'
  if (fileName === '#Player.xlsx') return 'player'
  if (fileName === '#Team.xlsx') return 'team'
  if (fileName === '#TutorialBattle.xlsx') return 'tutorial'
  return 'generic'
}

export function tableNameFromFile(fileName: string): string {
  const base = fileName.replace(/^#/, '').replace(/\.xlsx$/i, '')
  return `Tb${base}`
}

export function parseLubanFieldType(type: string): Pick<LubanField, 'kind' | 'list' | 'refTable'> {
  const normalized = type.trim()
  const list = /(^|\W)list(?:#|,|$)/i.test(normalized)
  const refTable = normalized.match(/#ref=([A-Za-z_][A-Za-z0-9_]*)/i)?.[1]
  const scalar = normalized.replace(/^\(list[^)]*\),?/i, '').replace(/^list,?/i, '').split('#')[0].trim().toLowerCase()

  if (refTable) return { kind: 'ref', list, refTable }
  if (list) return { kind: 'list', list: true }
  if (scalar === 'string') return { kind: 'string', list: false }
  if (scalar === 'int' || scalar === 'long') return { kind: 'int', list: false }
  if (scalar === 'float' || scalar === 'double') return { kind: 'float', list: false }
  if (scalar === 'bool' || scalar === 'boolean') return { kind: 'bool', list: false }
  return { kind: 'unknown', list: false }
}

export function parseCellValue(value: unknown, field: LubanField): LubanCellValue {
  if (value === undefined || value === null || value === '') return field.list ? [] : null
  if (field.list) {
    if (Array.isArray(value)) return value.map(String).map((item) => item.trim()).filter(Boolean)
    return String(value).split(',').map((item) => item.trim()).filter(Boolean)
  }
  if (field.kind === 'bool') {
    if (typeof value === 'boolean') return value
    if (String(value).toLowerCase() === 'true' || String(value) === '1') return true
    if (String(value).toLowerCase() === 'false' || String(value) === '0') return false
  }
  if (field.kind === 'int' || field.kind === 'float') {
    if (typeof value === 'number') return value
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : String(value)
  }
  if (typeof value === 'number' || typeof value === 'boolean') return value
  return String(value)
}

export function serializeCellValue(value: LubanCellValue, field: LubanField): string | number | boolean | null {
  if (field.list) return Array.isArray(value) ? value.join(',') : value == null ? '' : String(value)
  if (value === undefined || value === null) return null
  if (Array.isArray(value)) return value.join(',')
  return value
}

export function rowId(row: LubanRow): string {
  const value = row.id
  return value == null || Array.isArray(value) ? '' : String(value)
}

export function displayLabel(row: LubanRow): string {
  for (const key of ['name', 'shortName', 'nickname', 'id']) {
    const value = row[key]
    if (value !== undefined && value !== null && !Array.isArray(value) && String(value).trim()) return String(value)
  }
  return '未命名记录'
}

export function cloneDocument(document: LubanTableDocument): LubanTableDocument {
  return structuredClone(document)
}

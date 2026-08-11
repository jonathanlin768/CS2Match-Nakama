import crypto from 'node:crypto'
import fs from 'node:fs/promises'
import path from 'node:path'
import ExcelJS from 'exceljs'
import {
  mapOwnedFiles,
  ownerForFile,
  parseCellValue,
  parseLubanFieldType,
  rowId,
  serializeCellValue,
  tableNameFromFile,
  type ConfigReference,
  type LubanCellValue,
  type LubanField,
  type LubanRow,
  type LubanTableDocument,
  type LubanTableSummary,
  type WorkbookVersion,
} from '../src/lib/luban'
import { dataFilePath, dataRoot, playerCardRoot, portraitRoot, resolveInside, teamLogoRoot } from './paths'
import { readLubanWorkbook } from './importTables'
import { validateDocuments } from '../src/lib/lubanValidation'

const maxImageBytes = 8 * 1024 * 1024
const imageExtensions = new Set(['.png', '.jpg', '.jpeg', '.webp'])

export interface SaveDocumentResult {
  document: LubanTableDocument
  backup: string
}

export async function listLubanTables(root = dataRoot): Promise<LubanTableSummary[]> {
  const entries = await fs.readdir(root, { withFileTypes: true })
  const files = entries.filter((entry) => entry.isFile() && /^#[A-Za-z0-9_]+\.xlsx$/.test(entry.name)).map((entry) => entry.name)
  const summaries = await Promise.all(files.map(async (fileName): Promise<LubanTableSummary> => {
    try {
      const document = await readLubanTable(fileName, root)
      return {
        fileName,
        tableName: document.tableName,
        owner: document.owner,
        editable: document.editable,
        rows: document.rows.length,
        columns: document.fields.length,
        status: 'ready',
        warnings: document.warnings,
      }
    } catch (error) {
      const owner = ownerForFile(fileName)
      return {
        fileName,
        tableName: tableNameFromFile(fileName),
        owner,
        editable: owner !== 'map',
        rows: 0,
        columns: 0,
        status: 'error',
        warnings: [errorMessage(error)],
      }
    }
  }))
  return summaries.sort((a, b) => a.fileName.localeCompare(b.fileName))
}

export async function readLubanTable(fileName: string, root = dataRoot): Promise<LubanTableDocument> {
  const target = dataFilePath(fileName, root)
  try {
    const workbook = await readLubanWorkbook(target)
    return await documentFromWorkbook(workbook, fileName, target)
  } catch (error) {
    throw new Error(`无法读取 ${fileName}: ${errorMessage(error)}`, { cause: error })
  }
}

async function documentFromWorkbook(workbook: ExcelJS.Workbook, fileName: string, target: string): Promise<LubanTableDocument> {
  const sheet = workbook.worksheets.find((candidate) => textValue(candidate.getRow(1).getCell(1).value) === '##var')
  if (!sheet) throw new Error(`${fileName} 缺少标准 ##var 表头工作表`)

  const fields: LubanField[] = []
  for (let column = 2; column <= sheet.columnCount; column += 1) {
    const key = textValue(sheet.getRow(1).getCell(column).value)
    if (!key) continue
    const type = textValue(sheet.getRow(2).getCell(column).value)
    const parsed = parseLubanFieldType(type)
    fields.push({
      key,
      type,
      group: textValue(sheet.getRow(3).getCell(column).value),
      comment: textValue(sheet.getRow(4).getCell(column).value),
      kind: parsed.kind,
      refTable: parsed.refTable,
      list: parsed.list,
      column,
    })
  }
  if (fields.length === 0) throw new Error(`${fileName} 没有可解析字段`)

  const rows: LubanRow[] = []
  for (let rowNumber = 5; rowNumber <= sheet.rowCount; rowNumber += 1) {
    const values: LubanRow = {}
    let populated = false
    for (const field of fields) {
      const raw = plainCellValue(sheet.getRow(rowNumber).getCell(field.column).value)
      const value = parseCellValue(raw, field)
      values[field.key] = value
      if (Array.isArray(value) ? value.length > 0 : value !== null && value !== '') populated = true
    }
    if (populated) rows.push(values)
  }
  const owner = ownerForFile(fileName)
  const warnings = fields.filter((field) => field.kind === 'unknown').map((field) => `${field.key} 使用未知类型 ${field.type}，按文本编辑`)
  return {
    fileName,
    tableName: tableNameFromFile(fileName),
    owner,
    editable: owner !== 'map',
    fields,
    rows,
    originalMeta: { sheetName: sheet.name, version: await fileVersion(target) },
    warnings,
    dirty: false,
  }
}

export async function readEditableDocuments(root = dataRoot): Promise<LubanTableDocument[]> {
  const summaries = await listLubanTables(root)
  return Promise.all(summaries.filter((summary) => summary.editable && summary.status === 'ready').map((summary) => readLubanTable(summary.fileName, root)))
}

export async function hydrateIncomingDocument(value: unknown, root = dataRoot): Promise<LubanTableDocument> {
  if (!value || typeof value !== 'object') throw new Error('表文档格式不正确')
  const input = value as Partial<LubanTableDocument>
  if (typeof input.fileName !== 'string') throw new Error('表文档缺少 fileName')
  assertEditable(input.fileName)
  const base = await readLubanTable(input.fileName, root)
  if (!Array.isArray(input.rows)) throw new Error(`${input.fileName} 缺少 rows`)
  const version = input.originalMeta?.version
  if (!version || typeof version.hash !== 'string' || typeof version.size !== 'number' || typeof version.mtimeMs !== 'number') {
    throw new Error(`${input.fileName} 缺少有效的文件版本`)
  }
  base.originalMeta.version = version
  base.rows = input.rows.map((candidate) => {
    if (!candidate || typeof candidate !== 'object' || Array.isArray(candidate)) throw new Error(`${input.fileName} 包含非法数据行`)
    const raw = candidate as Record<string, unknown>
    return Object.fromEntries(base.fields.map((field) => [field.key, parseCellValue(raw[field.key], field)]))
  })
  base.dirty = true
  return base
}

export async function saveLubanTable(document: LubanTableDocument, root = dataRoot): Promise<SaveDocumentResult> {
  const results = await saveLubanTables([document], root)
  return results[0]
}

export async function saveLubanTables(documents: LubanTableDocument[], root = dataRoot): Promise<SaveDocumentResult[]> {
  if (documents.length === 0) return []
  const unique = new Set<string>()
  for (const document of documents) {
    if (unique.has(document.fileName)) throw new Error(`重复保存表 ${document.fileName}`)
    unique.add(document.fileName)
    assertEditable(document.fileName)
  }

  const stamp = timestamp()
  const prepared: PreparedWrite[] = []
  try {
    for (const document of documents) prepared.push(await prepareWrite(document, root, stamp))
    for (const item of prepared) await swapPrepared(item)
  } catch (error) {
    await rollbackPrepared(prepared)
    throw error
  }
  await Promise.all(prepared.map(cleanPrepared))

  return Promise.all(prepared.map(async (item) => ({
    document: await readLubanTable(item.document.fileName, root),
    backup: path.relative(root, item.backup),
  })))
}

export async function findReferences(targetTable: string, targetId: string, root = dataRoot): Promise<ConfigReference[]> {
  const summaries = await listLubanTables(root)
  const documents = await Promise.all(summaries.filter((summary) => summary.status === 'ready').map((summary) => readLubanTable(summary.fileName, root)))
  return referencesInDocuments(documents, targetTable, targetId)
}

export async function syncRecordId(fileName: string, oldId: string, newId: string, root = dataRoot): Promise<{ documents: LubanTableDocument[]; references: ConfigReference[] }> {
  if (!oldId || !newId) throw new Error('旧 ID 和新 ID 均不能为空')
  assertEditable(fileName)
  const summaries = await listLubanTables(root)
  const documents = await Promise.all(summaries.filter((summary) => summary.status === 'ready').map((summary) => readLubanTable(summary.fileName, root)))
  const source = documents.find((document) => document.fileName === fileName)
  if (!source) throw new Error(`找不到 ${fileName}`)
  if (source.rows.some((row) => rowId(row) === newId)) throw new Error(`新 ID ${newId} 已存在`)
  const sourceRow = source.rows.find((row) => rowId(row) === oldId)
  if (!sourceRow) throw new Error(`找不到 ID ${oldId}`)
  const references = referencesInDocuments(documents, source.tableName, oldId)
  const mapReferences = references.filter((reference) => mapOwnedFiles.has(reference.sourceFile))
  if (mapReferences.length > 0) throw new Error('该 ID 被地图语义表引用，不能通过非地图配置页同步修改')

  sourceRow.id = newId
  const changed = new Set<LubanTableDocument>([source])
  for (const reference of references) {
    const document = documents.find((candidate) => candidate.fileName === reference.sourceFile)
    const row = document?.rows.find((candidate) => rowId(candidate) === reference.sourceRowId)
    const field = document?.fields.find((candidate) => candidate.key === reference.field)
    if (!document || !row || !field) continue
    if (field.list) row[field.key] = listValue(row[field.key]).map((value) => value === oldId ? newId : value)
    else row[field.key] = newId
    changed.add(document)
  }
  const issues = validateDocuments(documents)
  const blocking = issues.filter((issue) => issue.severity === 'ERROR')
  if (blocking.length > 0) throw new Error(`同步 ID 后校验失败：${blocking[0].message}`)
  const results = await saveLubanTables([...changed], root)
  return { documents: results.map((result) => result.document), references }
}

export async function saveConfigImage(
  kind: 'portrait' | 'team' | 'player-card',
  fileName: string,
  dataBase64: string,
  overwrite = false,
  roots: { portrait: string; team: string; playerCard?: string } = { portrait: portraitRoot, team: teamLogoRoot, playerCard: playerCardRoot },
): Promise<{ path: string; url: string }> {
  if (path.basename(fileName) !== fileName || !/^[A-Za-z0-9_.-]+$/.test(fileName)) throw new Error('图片文件名不合法')
  const extension = path.extname(fileName).toLowerCase()
  if (!imageExtensions.has(extension)) throw new Error('只支持 PNG、JPG、JPEG 和 WEBP 图片')
  const data = Buffer.from(dataBase64, 'base64')
  if (data.length === 0 || data.length > maxImageBytes) throw new Error('图片为空或超过 8MB')
  if (!matchesImageSignature(data, extension)) throw new Error('文件内容与图片扩展名不匹配')

  const root = kind === 'portrait' ? roots.portrait : kind === 'team' ? roots.team : roots.playerCard ?? playerCardRoot
  await fs.mkdir(root, { recursive: true })
  const target = resolveInside(root, fileName)
  if (!overwrite && await exists(target)) throw new Error('目标图片已存在，请确认覆盖或重新命名')
  await fs.writeFile(target, data, { flag: overwrite ? 'w' : 'wx' })
  const directory = kind === 'portrait' ? 'portraits' : kind === 'team' ? 'teams' : 'player-cards'
  const relative = `${directory}/${fileName}`
  return { path: relative, url: `/${relative}` }
}

export async function configAssetExists(relativePath: string): Promise<boolean> {
  const target = configAssetFilePath(relativePath)
  return target ? exists(target) : false
}

export function configAssetFilePath(relativePath: string): string | null {
  const normalized = relativePath.replace(/\\/g, '/')
  if (normalized.startsWith('portraits/')) return resolveInside(portraitRoot, normalized.slice('portraits/'.length))
  if (normalized.startsWith('teams/')) return resolveInside(teamLogoRoot, normalized.slice('teams/'.length))
  if (normalized.startsWith('player-cards/')) return resolveInside(playerCardRoot, normalized.slice('player-cards/'.length))
  return null
}

function referencesInDocuments(documents: LubanTableDocument[], targetTable: string, targetId: string): ConfigReference[] {
  const references: ConfigReference[] = []
  for (const document of documents) {
    for (const field of document.fields.filter((candidate) => candidate.refTable === targetTable)) {
      for (const row of document.rows) {
        const values = field.list ? listValue(row[field.key]) : [scalarValue(row[field.key])]
        if (values.includes(targetId)) {
          references.push({ sourceFile: document.fileName, sourceTable: document.tableName, sourceRowId: rowId(row), field: field.key, targetTable, targetId })
        }
      }
    }
  }
  return references
}

interface PreparedWrite {
  document: LubanTableDocument
  target: string
  temp: string
  rollback: string
  backup: string
  swapped: boolean
}

async function prepareWrite(document: LubanTableDocument, root: string, stamp: string): Promise<PreparedWrite> {
  const target = dataFilePath(document.fileName, root)
  const currentVersion = await fileVersion(target)
  if (!sameVersion(currentVersion, document.originalMeta.version)) throw new Error(`${document.fileName} 已被外部修改，请从项目重新读取`)

  let workbook: ExcelJS.Workbook
  try {
    workbook = await readLubanWorkbook(target)
  } catch (error) {
    throw new Error(`无法保存 ${document.fileName}: ${errorMessage(error)}`, { cause: error })
  }
  const sheet = workbook.getWorksheet(document.originalMeta.sheetName)
  if (!sheet || textValue(sheet.getRow(1).getCell(1).value) !== '##var') throw new Error(`${document.fileName} 原工作表结构已损坏`)
  const headerKeys = document.fields.map((field) => textValue(sheet.getRow(1).getCell(field.column).value))
  if (headerKeys.some((key, index) => key !== document.fields[index].key)) throw new Error(`${document.fileName} 表头已变化，请重新读取`)

  const oldDataCount = Math.max(0, sheet.rowCount - 4)
  for (let index = 0; index < document.rows.length; index += 1) {
    const row = sheet.getRow(5 + index)
    row.getCell(1).value = null
    for (const field of document.fields) row.getCell(field.column).value = serializeCellValue(document.rows[index][field.key], field)
    row.commit()
  }
  if (oldDataCount > document.rows.length) sheet.spliceRows(5 + document.rows.length, oldDataCount - document.rows.length)

  const temp = `${target}.config-editor-${process.pid}-${Date.now()}.tmp`
  const rollback = `${target}.config-editor-${process.pid}-${Date.now()}.rollback`
  const backupDir = resolveInside(root, '.bak', 'config-editor', stamp)
  await fs.mkdir(backupDir, { recursive: true })
  const backup = resolveInside(backupDir, path.basename(target))
  await fs.copyFile(target, backup)
  await workbook.xlsx.writeFile(temp)
  return { document, target, temp, rollback, backup, swapped: false }
}

async function swapPrepared(item: PreparedWrite): Promise<void> {
  await fs.rename(item.target, item.rollback)
  try {
    await fs.rename(item.temp, item.target)
    item.swapped = true
  } catch (error) {
    await fs.rename(item.rollback, item.target)
    throw error
  }
}

async function rollbackPrepared(items: PreparedWrite[]): Promise<void> {
  for (const item of [...items].reverse()) {
    if (item.swapped) {
      await fs.rm(item.target, { force: true })
      await fs.rename(item.rollback, item.target).catch(() => undefined)
    }
    await fs.rm(item.temp, { force: true }).catch(() => undefined)
  }
}

async function cleanPrepared(item: PreparedWrite): Promise<void> {
  await fs.rm(item.rollback, { force: true }).catch(() => undefined)
  await fs.rm(item.temp, { force: true }).catch(() => undefined)
}

function assertEditable(fileName: string): void {
  dataFilePath(fileName)
  if (mapOwnedFiles.has(fileName)) throw new Error(`${fileName} 只能通过地图配置页编辑`)
}

async function fileVersion(file: string): Promise<WorkbookVersion> {
  const [data, stat] = await Promise.all([fs.readFile(file), fs.stat(file)])
  return { mtimeMs: stat.mtimeMs, size: stat.size, hash: crypto.createHash('sha256').update(data).digest('hex') }
}

function sameVersion(left: WorkbookVersion, right: WorkbookVersion): boolean {
  return left.size === right.size && left.hash === right.hash
}

function textValue(value: ExcelJS.CellValue): string {
  const plain = plainCellValue(value)
  return plain == null ? '' : String(plain).trim()
}

function plainCellValue(value: ExcelJS.CellValue): unknown {
  if (value && typeof value === 'object') {
    if ('result' in value) return value.result
    if ('text' in value) return value.text
    if ('richText' in value) return value.richText.map((part) => part.text).join('')
  }
  return value
}

function scalarValue(value: LubanCellValue | undefined): string {
  return value == null || Array.isArray(value) ? '' : String(value)
}

function listValue(value: LubanCellValue | undefined): string[] {
  if (Array.isArray(value)) return value.map(String)
  return value == null || value === '' ? [] : String(value).split(',').map((item) => item.trim()).filter(Boolean)
}

function matchesImageSignature(data: Buffer, extension: string): boolean {
  if (extension === '.png') return data.subarray(0, 4).equals(Buffer.from([0x89, 0x50, 0x4e, 0x47]))
  if (extension === '.jpg' || extension === '.jpeg') return data[0] === 0xff && data[1] === 0xd8 && data[2] === 0xff
  if (extension === '.webp') return data.subarray(0, 4).toString('ascii') === 'RIFF' && data.subarray(8, 12).toString('ascii') === 'WEBP'
  return false
}

async function exists(file: string): Promise<boolean> {
  try {
    await fs.access(file)
    return true
  } catch {
    return false
  }
}

function timestamp(): string {
  const now = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}-${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}-${now.getMilliseconds()}`
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}

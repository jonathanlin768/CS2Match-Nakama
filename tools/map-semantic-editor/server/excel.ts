import fs from 'node:fs/promises'
import path from 'node:path'
import ExcelJS from 'exceljs'
import type { ExportTable } from '../src/lib/exportTables'
import { dataFilePath, dataRoot, resolveInside } from './paths'

export interface WriteEntry {
  tableName: string
  file: string
  rows: number
  backup?: string
  warnings: string[]
}

export interface WriteResult {
  ok: true
  root: string
  entries: WriteEntry[]
}

export async function writeExportTables(tables: ExportTable[], root = dataRoot): Promise<WriteResult> {
  const stamp = timestamp()
  const entries: WriteEntry[] = []

  await fs.mkdir(root, { recursive: true })

  for (const table of tables) {
    const target = dataFilePath(table.fileName, root)
    const backup = await backupExisting(target, stamp)
    await writeTableWorkbook(target, table)
    entries.push({
      tableName: table.tableName,
      file: path.relative(root, target),
      rows: table.rows.length,
      backup: backup ? path.relative(root, backup) : undefined,
      warnings: table.rows.length === 0 ? ['表为空，仅写入表头'] : [],
    })
  }

  return { ok: true, root, entries }
}

export async function readLubanSummary(): Promise<{ file: string; rows: number; columns: number }[]> {
  const entries = await fs.readdir(dataRoot, { withFileTypes: true })
  const result: { file: string; rows: number; columns: number }[] = []

  for (const entry of entries) {
    if (!entry.isFile() || !entry.name.endsWith('.xlsx')) continue
    const workbook = new ExcelJS.Workbook()
    await workbook.xlsx.readFile(path.join(dataRoot, entry.name))
    const sheet = workbook.worksheets[0]
    result.push({
      file: entry.name,
      rows: Math.max(0, sheet.rowCount - 4),
      columns: sheet.columnCount,
    })
  }

  return result.sort((a, b) => a.file.localeCompare(b.file))
}

async function writeTableWorkbook(target: string, table: ExportTable): Promise<void> {
  const workbook = new ExcelJS.Workbook()
  workbook.creator = 'map-semantic-editor'
  workbook.created = new Date()
  const sheet = workbook.addWorksheet('Sheet1')
  const fields = table.fields

  sheet.addRow(['##var', ...fields.map((field) => field.key)])
  sheet.addRow(['##type', ...fields.map((field) => field.type)])
  sheet.addRow(['##group', ...fields.map((field) => field.group ?? '')])
  sheet.addRow(['##', ...fields.map((field) => field.comment)])

  for (const row of table.rows) {
    sheet.addRow([null, ...fields.map((field) => normalizeCell(row[field.key]))])
  }

  for (let column = 1; column <= fields.length + 1; column += 1) {
    sheet.getColumn(column).width = column === 1 ? 12 : Math.max(14, Math.min(34, String(sheet.getRow(1).getCell(column).value ?? '').length + 8))
  }

  const temp = `${target}.tmp-${process.pid}-${Date.now()}`
  await workbook.xlsx.writeFile(temp)
  await fs.rename(temp, target)
}

function normalizeCell(value: unknown): string | number | boolean | null {
  if (value === undefined || value === null) return null
  if (typeof value === 'number' || typeof value === 'boolean') return value
  return String(value)
}

async function backupExisting(target: string, stamp: string): Promise<string | undefined> {
  try {
    await fs.access(target)
  } catch {
    return undefined
  }

  const backupDir = resolveInside(path.dirname(target), '.bak', 'map-semantic-editor', stamp)
  await fs.mkdir(backupDir, { recursive: true })
  const backup = resolveInside(backupDir, path.basename(target))
  await fs.copyFile(target, backup)
  return backup
}

function timestamp(): string {
  const now = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}-${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`
}

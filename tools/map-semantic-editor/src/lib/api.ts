import { parseProject, type MapProject } from './model'
import type { ValidationIssue } from './validation'
import type { ConfigReference, ConfigValidationIssue, LubanTableDocument, LubanTableSummary } from './luban'

export interface WriteLogEntry {
  tableName: string
  file: string
  rows: number
  backup?: string
  warnings: string[]
}

export interface ImportLogEntry {
  tableName: string
  file: string
  rows: number
  warnings: string[]
}

export interface GenConfigOutput {
  status: 'success' | 'failed' | 'timeout' | 'running'
  exitCode: number | null
  durationMs: number
  stdout: string
  stderr: string
}

export async function fetchProject(): Promise<MapProject> {
  const response = await fetch('/api/project')
  const body = await response.json() as ApiBody
  if (!response.ok || !body.ok) throw new Error(body.error ?? '读取工程失败')
  return parseProject(body.project)
}

export async function fetchPublishedProject(projectName = 'de_dust2.json'): Promise<MapProject> {
  const response = await fetch(`/api/luban/project?name=${encodeURIComponent(projectName)}`)
  const body = await response.json() as ApiBody
  if (!response.ok || !body.ok) throw new Error(body.error ?? '读取发布快照失败')
  return parseProject(body.project)
}

export async function importLubanConfig(project: MapProject): Promise<{ project: MapProject; summary: ImportLogEntry[]; warnings: string[] }> {
  const response = await fetch('/api/luban/import', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ project }),
  })
  const body = await response.json() as ApiBody
  if (!response.ok || !body.ok) throw new Error(body.error ?? '读取 Excel 配置失败')
  return {
    project: parseProject(body.project),
    summary: Array.isArray(body.summary) ? body.summary as ImportLogEntry[] : [],
    warnings: Array.isArray(body.warnings) ? body.warnings as string[] : [],
  }
}

export async function saveProject(project: MapProject): Promise<void> {
  const response = await fetch('/api/project/save', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: `${project.map_id}.json`, project }),
  })
  const body = await response.json() as ApiBody
  if (!response.ok || !body.ok) throw new Error(body.error ?? '保存工程失败')
}

export async function writeLuban(project: MapProject): Promise<{ entries: WriteLogEntry[]; issues: ValidationIssue[] }> {
  const response = await fetch('/api/luban/write', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ project }),
  })
  const body = await response.json() as ApiBody
  if (!response.ok || !body.ok) {
    if (Array.isArray(body.issues)) return Promise.reject(Object.assign(new Error('校验未通过，已阻止写入'), { issues: body.issues }))
    throw new Error(body.error ?? '写入 Luban 失败')
  }
  return { entries: (body.result as { entries: WriteLogEntry[] }).entries, issues: body.issues as ValidationIssue[] }
}

export async function runGenConfig(): Promise<GenConfigOutput> {
  const response = await fetch('/api/gen-config', { method: 'POST' })
  const body = await readApiBody(response)
  if (!response.ok) throw new Error(body.error ? `导表接口失败 (${response.status}): ${body.error}` : `导表接口失败 (${response.status})`)
  if (!isGenConfigOutput(body.result)) throw new Error('导表接口返回格式异常')
  return body.result as GenConfigOutput
}

export async function fetchConfigTables(): Promise<LubanTableSummary[]> {
  const response = await fetch('/api/config/tables')
  const body = await readApiBody(response)
  if (!response.ok || !body.ok || !Array.isArray(body.tables)) throw new Error(body.error ?? '读取配置表列表失败')
  return body.tables as LubanTableSummary[]
}

export async function fetchConfigTable(fileName: string): Promise<LubanTableDocument> {
  const response = await fetch(`/api/config/table?file=${encodeURIComponent(fileName)}`)
  const body = await readApiBody(response)
  if (!response.ok || !body.ok || !body.document) throw new Error(body.error ?? `读取 ${fileName} 失败`)
  return body.document as LubanTableDocument
}

export async function saveConfigTable(document: LubanTableDocument): Promise<{ document: LubanTableDocument; backup: string; issues: ConfigValidationIssue[] }> {
  const response = await fetch('/api/config/table/save', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ document }),
  })
  const body = await readApiBody(response)
  if (!response.ok || !body.ok || !body.result) throw configApiError(body, `保存 ${document.fileName} 失败`)
  const result = body.result as { document: LubanTableDocument; backup: string }
  return { ...result, issues: Array.isArray(body.issues) ? body.issues as ConfigValidationIssue[] : [] }
}

export async function saveConfigTables(documents: LubanTableDocument[]): Promise<{ results: { document: LubanTableDocument; backup: string }[]; issues: ConfigValidationIssue[] }> {
  const response = await fetch('/api/config/tables/save', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ documents }),
  })
  const body = await readApiBody(response)
  if (!response.ok || !body.ok || !Array.isArray(body.results)) throw configApiError(body, '保存全部已修改表失败')
  return { results: body.results as { document: LubanTableDocument; backup: string }[], issues: Array.isArray(body.issues) ? body.issues as ConfigValidationIssue[] : [] }
}

export async function fetchConfigReferences(table: string, id: string): Promise<ConfigReference[]> {
  const response = await fetch(`/api/config/references?table=${encodeURIComponent(table)}&id=${encodeURIComponent(id)}`)
  const body = await readApiBody(response)
  if (!response.ok || !body.ok || !Array.isArray(body.references)) throw new Error(body.error ?? '读取引用位置失败')
  return body.references as ConfigReference[]
}

export async function syncConfigId(fileName: string, oldId: string, newId: string): Promise<{ documents: LubanTableDocument[]; references: ConfigReference[] }> {
  const response = await fetch('/api/config/id-sync', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ fileName, oldId, newId }),
  })
  const body = await readApiBody(response)
  if (!response.ok || !body.ok || !Array.isArray(body.documents) || !Array.isArray(body.references)) throw new Error(body.error ?? '同步 ID 失败')
  return { documents: body.documents as LubanTableDocument[], references: body.references as ConfigReference[] }
}

export async function uploadConfigImage(kind: 'portrait' | 'team' | 'player-card', file: File, overwrite = false): Promise<{ path: string; url: string }> {
  const response = await fetch('/api/config/image', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ kind, fileName: file.name, dataBase64: await fileToBase64(file), overwrite }),
  })
  const body = await readApiBody(response)
  if (!response.ok || !body.ok || !body.result) throw new Error(body.error ?? '复制图片失败')
  return body.result as { path: string; url: string }
}

export async function configAssetExists(assetPath: string): Promise<boolean> {
  const response = await fetch(`/api/config/asset?path=${encodeURIComponent(assetPath)}`)
  const body = await readApiBody(response)
  return response.ok && body.ok === true && body.exists === true
}

export function configAssetUrl(assetPath: string): string {
  return `/api/config/asset-file?path=${encodeURIComponent(assetPath.replace(/^\//, ''))}`
}

async function readApiBody(response: Response): Promise<ApiBody> {
  const text = await response.text()
  if (!text) return {}
  try {
    return JSON.parse(text) as ApiBody
  } catch {
    return { error: text }
  }
}

interface ApiBody {
  ok?: boolean
  error?: string
  project?: unknown
  summary?: unknown
  warnings?: unknown
  result?: unknown
  issues?: unknown
  tables?: unknown
  document?: unknown
  documents?: unknown
  references?: unknown
  results?: unknown
  exists?: unknown
}

function isGenConfigOutput(value: unknown): value is GenConfigOutput {
  if (!value || typeof value !== 'object') return false
  const output = value as Partial<GenConfigOutput>
  return (
    (output.status === 'success' || output.status === 'failed' || output.status === 'timeout' || output.status === 'running') &&
    (typeof output.exitCode === 'number' || output.exitCode === null) &&
    typeof output.durationMs === 'number' &&
    typeof output.stdout === 'string' &&
    typeof output.stderr === 'string'
  )
}

function configApiError(body: ApiBody, fallback: string): Error {
  const error = new Error(body.error ?? fallback) as Error & { issues?: ConfigValidationIssue[] }
  if (Array.isArray(body.issues)) error.issues = body.issues as ConfigValidationIssue[]
  return error
}

async function fileToBase64(file: File): Promise<string> {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  let binary = ''
  const chunkSize = 0x8000
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize))
  }
  return btoa(binary)
}

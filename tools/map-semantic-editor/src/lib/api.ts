import { parseProject, type MapProject } from './model'
import type { ValidationIssue } from './validation'

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

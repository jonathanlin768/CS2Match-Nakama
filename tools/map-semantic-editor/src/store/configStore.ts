import { create } from 'zustand'
import {
  fetchConfigReferences,
  fetchConfigTable,
  fetchConfigTables,
  runGenConfig,
  saveConfigTable,
  saveConfigTables,
  syncConfigId,
  uploadConfigImage,
  type GenConfigOutput,
} from '../lib/api'
import { cloneDocument, rowId, type ConfigReference, type ConfigValidationIssue, type LubanCellValue, type LubanRow, type LubanTableDocument, type LubanTableSummary } from '../lib/luban'
import { validateDocuments } from '../lib/lubanValidation'

export interface ConfigLogEntry {
  time: string
  level: 'info' | 'error'
  message: string
}

interface ConfigState {
  summaries: LubanTableSummary[]
  documents: Record<string, LubanTableDocument>
  activeFile: string | null
  issues: ConfigValidationIssue[]
  logs: ConfigLogEntry[]
  genConfig: GenConfigOutput | null
  serviceMessage: string
  busy: boolean
  initialized: boolean
  load: () => Promise<void>
  reload: () => Promise<void>
  setActiveFile: (fileName: string) => void
  updateCell: (fileName: string, rowIndex: number, field: string, value: LubanCellValue) => void
  updateCells: (fileName: string, rowIndex: number, values: Record<string, LubanCellValue>) => void
  addRow: (fileName: string) => void
  duplicateRow: (fileName: string, rowIndex: number) => void
  deleteRow: (fileName: string, rowIndex: number) => Promise<boolean>
  changeId: (fileName: string, rowIndex: number, newId: string) => Promise<boolean>
  setTutorialEnabled: (rowIndex: number, enabled: boolean) => void
  saveCurrent: (fileName?: string) => Promise<boolean>
  saveAll: () => Promise<boolean>
  runExport: () => Promise<void>
  uploadImage: (kind: 'portrait' | 'team' | 'player-card', file: File) => Promise<string | null>
  referencesFor: (fileName: string, rowIndex: number) => Promise<ConfigReference[]>
  validate: () => ConfigValidationIssue[]
}

type ConfigSet = (updater: Partial<ConfigState> | ((state: ConfigState) => Partial<ConfigState>)) => void

export const useConfigStore = create<ConfigState>((set, get) => ({
  summaries: [],
  documents: {},
  activeFile: null,
  issues: [],
  logs: [],
  genConfig: null,
  serviceMessage: '配置服务: 未连接',
  busy: false,
  initialized: false,

  load: async () => {
    if (get().initialized) return
    await loadAll(set, get)
  },

  reload: async () => {
    const dirty = Object.values(get().documents).some((document) => document.dirty)
    if (dirty && !window.confirm('从项目重新读取会丢失尚未保存的修改，是否继续？')) return
    await loadAll(set, get)
  },

  setActiveFile: (fileName) => set({ activeFile: fileName }),

  updateCell: (fileName, rowIndex, field, value) => {
    mutateDocument(set, get, fileName, (document) => {
      if (!document.rows[rowIndex]) return
      document.rows[rowIndex][field] = value
    })
  },

  updateCells: (fileName, rowIndex, values) => {
    mutateDocument(set, get, fileName, (document) => {
      if (!document.rows[rowIndex]) return
      Object.assign(document.rows[rowIndex], values)
    })
  },

  addRow: (fileName) => {
    mutateDocument(set, get, fileName, (document) => {
      const row = Object.fromEntries(document.fields.map((field) => [field.key, field.list ? [] : field.kind === 'bool' ? false : null])) as LubanRow
      document.rows.push(row)
    })
  },

  duplicateRow: (fileName, rowIndex) => {
    mutateDocument(set, get, fileName, (document) => {
      const source = document.rows[rowIndex]
      if (!source) return
      const copy = structuredClone(source)
      copy.id = uniqueRowId(document, `${rowId(source) || 'record'}_copy`)
      document.rows.splice(rowIndex + 1, 0, copy)
    })
  },

  deleteRow: async (fileName, rowIndex) => {
    const references = await get().referencesFor(fileName, rowIndex)
    if (references.length > 0) {
      set({ serviceMessage: `删除已阻止：${references.map(referenceLabel).join('；')}` })
      appendLog(set, 'error', `删除已阻止，记录仍被 ${references.length} 处引用`)
      return false
    }
    mutateDocument(set, get, fileName, (document) => document.rows.splice(rowIndex, 1))
    return true
  },

  changeId: async (fileName, rowIndex, newId) => {
    const document = get().documents[fileName]
    const row = document?.rows[rowIndex]
    const oldId = row ? rowId(row) : ''
    const trimmed = newId.trim()
    if (!document || !row || !oldId || !trimmed || oldId === trimmed) return false
    if (document.rows.some((candidate, index) => index !== rowIndex && rowId(candidate) === trimmed)) {
      set({ serviceMessage: `ID ${trimmed} 已存在` })
      return false
    }
    const references = await fetchConfigReferences(document.tableName, oldId)
    if (references.length === 0) {
      get().updateCell(fileName, rowIndex, 'id', trimmed)
      return true
    }
    if (Object.values(get().documents).some((candidate) => candidate.dirty)) {
      set({ serviceMessage: '同步 ID 前请先保存或重新读取所有未保存修改' })
      return false
    }
    const detail = references.map(referenceLabel).join('\n')
    if (!window.confirm(`修改 ID 将同步更新以下引用：\n${detail}\n\n是否继续？`)) return false
    set({ busy: true, serviceMessage: '正在同步 ID 与引用' })
    try {
      const result = await syncConfigId(fileName, oldId, trimmed)
      const documents = { ...get().documents }
      for (const changed of result.documents) documents[changed.fileName] = changed
      const issues = validateDocuments(Object.values(documents))
      set({ documents, issues, busy: false, serviceMessage: `已同步修改 ${oldId} -> ${trimmed}` })
      appendLog(set, 'info', `已同步修改 ${oldId} -> ${trimmed}，更新 ${references.length} 处引用`)
      return true
    } catch (error) {
      set({ busy: false, serviceMessage: errorMessage(error) })
      appendLog(set, 'error', errorMessage(error))
      return false
    }
  },

  setTutorialEnabled: (rowIndex, enabled) => {
    const fileName = '#TutorialBattle.xlsx'
    const document = get().documents[fileName]
    if (!document?.rows[rowIndex]) return
    if (enabled && document.rows.some((row, index) => index !== rowIndex && row.enabled === true)) {
      if (!window.confirm('启用该配置会关闭当前已启用的新手战斗配置，是否继续？')) return
    }
    mutateDocument(set, get, fileName, (draft) => {
      draft.rows.forEach((row, index) => { row.enabled = index === rowIndex ? enabled : enabled ? false : row.enabled })
    })
  },

  saveCurrent: async (fileName) => {
    const target = fileName ?? get().activeFile
    const document = target ? get().documents[target] : undefined
    if (!document || !document.dirty) return true
    const issues = validateDocuments(Object.values(get().documents))
    set({ issues })
    if (issues.some((issue) => issue.severity === 'ERROR')) {
      set({ serviceMessage: '校验存在 ERROR，已阻止保存' })
      return false
    }
    set({ busy: true, serviceMessage: `正在保存 ${document.fileName}` })
    try {
      const result = await saveConfigTable(document)
      const documents = { ...get().documents, [document.fileName]: result.document }
      set({ documents, issues: result.issues, busy: false, serviceMessage: `${document.fileName} 已保存` })
      appendLog(set, 'info', `${document.fileName} 已保存，备份: ${result.backup}`)
      return true
    } catch (error) {
      const maybeIssues = (error as Error & { issues?: ConfigValidationIssue[] }).issues
      set({ busy: false, issues: maybeIssues ?? issues, serviceMessage: errorMessage(error) })
      appendLog(set, 'error', errorMessage(error))
      return false
    }
  },

  saveAll: async () => {
    const dirty = Object.values(get().documents).filter((document) => document.dirty)
    if (dirty.length === 0) return true
    const issues = validateDocuments(Object.values(get().documents))
    set({ issues })
    if (issues.some((issue) => issue.severity === 'ERROR')) {
      set({ serviceMessage: '校验存在 ERROR，已阻止批量保存' })
      return false
    }
    set({ busy: true, serviceMessage: `正在保存 ${dirty.length} 张表` })
    try {
      const result = await saveConfigTables(dirty)
      const documents = { ...get().documents }
      for (const saved of result.results) documents[saved.document.fileName] = saved.document
      set({ documents, issues: result.issues, busy: false, serviceMessage: `已保存 ${dirty.length} 张表` })
      result.results.forEach((saved) => appendLog(set, 'info', `${saved.document.fileName} 已保存，备份: ${saved.backup}`))
      return true
    } catch (error) {
      const maybeIssues = (error as Error & { issues?: ConfigValidationIssue[] }).issues
      set({ busy: false, issues: maybeIssues ?? issues, serviceMessage: errorMessage(error) })
      appendLog(set, 'error', errorMessage(error))
      return false
    }
  },

  runExport: async () => {
    const dirty = Object.values(get().documents).filter((document) => document.dirty)
    if (dirty.length > 0 && !window.confirm(`当前有 ${dirty.length} 张表尚未保存。继续导表将使用磁盘上的旧数据，是否继续？`)) return
    set({ busy: true, genConfig: { status: 'running', exitCode: null, durationMs: 0, stdout: '', stderr: '' }, serviceMessage: '导表运行中' })
    try {
      const genConfig = await runGenConfig()
      set({ genConfig, busy: false, serviceMessage: genConfig.status === 'success' ? '导表成功' : '导表失败' })
      appendLog(set, genConfig.status === 'success' ? 'info' : 'error', `导表${genConfig.status === 'success' ? '成功' : '失败'}，退出码 ${genConfig.exitCode ?? 'null'}`)
    } catch (error) {
      set({ busy: false, serviceMessage: errorMessage(error), genConfig: { status: 'failed', exitCode: null, durationMs: 0, stdout: '', stderr: errorMessage(error) } })
      appendLog(set, 'error', errorMessage(error))
    }
  },

  uploadImage: async (kind, file) => {
    set({ busy: true, serviceMessage: `正在复制 ${file.name}` })
    try {
      let result
      try {
        result = await uploadConfigImage(kind, file)
      } catch (error) {
        if (!errorMessage(error).includes('已存在') || !window.confirm(`${file.name} 已存在，是否覆盖？`)) throw error
        result = await uploadConfigImage(kind, file, true)
      }
      set({ busy: false, serviceMessage: `图片已复制到 ${result.path}` })
      appendLog(set, 'info', `图片已复制到 ${result.path}`)
      return result.path
    } catch (error) {
      set({ busy: false, serviceMessage: errorMessage(error) })
      appendLog(set, 'error', errorMessage(error))
      return null
    }
  },

  referencesFor: async (fileName, rowIndex) => {
    const document = get().documents[fileName]
    const id = document?.rows[rowIndex] ? rowId(document.rows[rowIndex]) : ''
    if (!document || !id) return []
    try {
      return await fetchConfigReferences(document.tableName, id)
    } catch (error) {
      set({ serviceMessage: errorMessage(error) })
      return []
    }
  },

  validate: () => {
    const issues = validateDocuments(Object.values(get().documents))
    set({ issues, serviceMessage: issues.some((issue) => issue.severity === 'ERROR') ? `校验发现 ${issues.length} 个问题` : '校验通过' })
    return issues
  },
}))

async function loadAll(set: ConfigSet, get: () => ConfigState): Promise<void> {
  set({ busy: true, serviceMessage: '正在读取 Luban 配置表' })
  try {
    const summaries = await fetchConfigTables()
    const editable = summaries.filter((summary) => summary.editable && summary.status === 'ready')
    const loaded = await Promise.all(editable.map((summary) => fetchConfigTable(summary.fileName)))
    const documents = Object.fromEntries(loaded.map((document) => [document.fileName, document]))
    const issues = validateDocuments(loaded)
    set({
      summaries,
      documents,
      issues,
      activeFile: get().activeFile && documents[get().activeFile!] ? get().activeFile : editable[0]?.fileName ?? null,
      busy: false,
      initialized: true,
      serviceMessage: `已读取 ${loaded.length} 张可编辑配置表`,
    })
  } catch (error) {
    set({ busy: false, initialized: true, serviceMessage: errorMessage(error) })
    appendLog(set, 'error', errorMessage(error))
  }
}

function mutateDocument(set: ConfigSet, get: () => ConfigState, fileName: string, mutate: (document: LubanTableDocument) => void): void {
  const current = get().documents[fileName]
  if (!current) return
  const document = cloneDocument(current)
  mutate(document)
  document.dirty = true
  const documents = { ...get().documents, [fileName]: document }
  set({ documents, issues: validateDocuments(Object.values(documents)), serviceMessage: `${fileName} 有未保存修改` })
}

function uniqueRowId(document: LubanTableDocument, base: string): string {
  const existing = new Set(document.rows.map(rowId))
  if (!existing.has(base)) return base
  let index = 2
  while (existing.has(`${base}_${index}`)) index += 1
  return `${base}_${index}`
}

function referenceLabel(reference: ConfigReference): string {
  return `${reference.sourceFile}.${reference.field} (${reference.sourceRowId})`
}

function appendLog(set: ConfigSet, level: ConfigLogEntry['level'], message: string): void {
  set((state) => ({ logs: [...state.logs, { time: new Date().toLocaleTimeString(), level, message }] }))
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}

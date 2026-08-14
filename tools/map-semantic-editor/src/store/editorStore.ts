import { create } from 'zustand'
import { fetchProject, fetchPublishedProject, importLubanConfig, runGenConfig, saveProject, updateLocalConfig, writeLuban, type GenConfigOutput, type ImportLogEntry, type WriteLogEntry } from '../lib/api'
import type { MapEdge, MapNode, MapProject, Point, RangeDraft, Route, SelectedObject, ToolMode, Visibility } from '../lib/model'
import { cellToPoints } from '../lib/model'
import { sampleNode } from '../lib/sampling'
import { createSampleProject } from '../lib/sampleProject'
import { hasBlockingIssues, validateProject, type ValidationIssue } from '../lib/validation'

interface EditorState {
  project: MapProject
  tool: ToolMode
  selected: SelectedObject
  located: SelectedObject
  pendingNodeId: string | null
  routeDraft: string[]
  rangeDraft: RangeDraft | null
  history: MapProject[]
  future: MapProject[]
  issues: ValidationIssue[]
  samples: Point[]
  writeLog: WriteLogEntry[]
  importLog: ImportLogEntry[]
  genConfig: GenConfigOutput | null
  serviceMessage: string
  busy: boolean
  setTool: (tool: ToolMode) => void
  select: (selected: SelectedObject) => void
  locate: (selected: SelectedObject) => void
  newProject: () => void
  load: () => Promise<void>
  loadPublished: () => Promise<void>
  importFromExcel: () => Promise<void>
  save: () => Promise<void>
  validate: () => ValidationIssue[]
  previewSamples: () => void
  clearPreview: () => void
  write: () => Promise<void>
  runExport: () => Promise<void>
  updateLocal: () => Promise<void>
  addNode: (point: Point, risk?: boolean) => void
  moveNode: (id: string, point: Point) => void
  updateNode: (id: string, patch: Partial<MapNode>) => void
  updateEdge: (id: string, patch: Partial<MapEdge>) => void
  updateVisibility: (id: string, patch: Partial<Visibility>) => void
  updateRoute: (id: string, patch: Partial<Route>) => void
  updateLayer: (layer: string, patch: { visible?: boolean; locked?: boolean; color?: string }) => void
  deleteSelected: () => void
  duplicateSelected: () => void
  handleNodeToolClick: (nodeId: string) => void
  setNodePointsFromText: (id: string, value: string) => void
  startRangeDraw: (nodeId: string, shape: RangeDraft['shape']) => void
  appendRangePoint: (point: Point) => void
  finishRangeDraw: () => void
  cancelRangeDraw: () => void
  finishRouteDraft: () => void
  cancelRouteDraft: () => void
  undo: () => void
  redo: () => void
  centerView: () => void
}

export const useEditorStore = create<EditorState>((set, get) => ({
  project: createSampleProject(),
  tool: 'select',
  selected: null,
  located: null,
  pendingNodeId: null,
  routeDraft: [],
  rangeDraft: null,
  history: [],
  future: [],
  issues: [],
  samples: [],
  writeLog: [],
  importLog: [],
  genConfig: null,
  serviceMessage: '服务: 未连接',
  busy: false,

  setTool: (tool) => set({ tool, pendingNodeId: null, routeDraft: [], rangeDraft: null }),
  select: (selected) => set({ selected }),
  locate: (selected) => set({
    selected,
    located: selected,
    serviceMessage: selected ? `已定位 ${selected.kind} ${selected.id}` : '该校验项没有可定位对象',
  }),
  newProject: () => {
    const project = createSampleProject()
    set({
      project,
      selected: null,
      located: null,
      pendingNodeId: null,
      routeDraft: [],
      rangeDraft: null,
      history: [],
      future: [],
      issues: validateProject(project),
      samples: [],
      writeLog: [],
      importLog: [],
      genConfig: null,
      serviceMessage: `已新建 ${project.map_id} 工程模板`,
    })
  },

  load: async () => {
    set({ busy: true })
    try {
      const project = await fetchProject()
      set({ project, located: null, serviceMessage: '服务: 已连接', issues: validateProject(project), importLog: [], busy: false })
    } catch (error) {
      set({ serviceMessage: errorMessage(error), busy: false })
    }
  },

  loadPublished: async () => {
    const { project: current } = get()
    set({ busy: true })
    try {
      const project = await fetchPublishedProject(`${current.map_id}.json`)
      set({
        project,
        selected: null,
        located: null,
        pendingNodeId: null,
        routeDraft: [],
        rangeDraft: null,
        samples: [],
        history: [],
        future: [],
        importLog: [],
        serviceMessage: `已读回 ${project.map_id}.json`,
        issues: validateProject(project),
        busy: false,
      })
    } catch (error) {
      set({ serviceMessage: errorMessage(error), busy: false })
    }
  },

  importFromExcel: async () => {
    const { project: current } = get()
    set({ busy: true })
    try {
      const { project, summary, warnings } = await importLubanConfig(current)
      const issues = validateProject(project)
      set({
        project,
        selected: null,
        located: null,
        pendingNodeId: null,
        routeDraft: [],
        rangeDraft: null,
        samples: [],
        history: [],
        future: [],
        issues,
        importLog: summary,
        serviceMessage: warnings.length > 0
          ? `已读取 Excel 配置（${summary.length} 张表，${warnings.length} 条警告）`
          : `已读取 Excel 配置（${summary.length} 张表）`,
        busy: false,
      })
    } catch (error) {
      set({ serviceMessage: errorMessage(error), busy: false })
    }
  },

  save: async () => {
    const { project } = get()
    set({ busy: true })
    try {
      await saveProject(project)
      set({ serviceMessage: '工程已保存', busy: false })
    } catch (error) {
      set({ serviceMessage: errorMessage(error), busy: false })
    }
  },

  validate: () => {
    const { project, located } = get()
    const issues = validateProject(project)
    set({ issues, located: retainLocatedIfIssueExists(located, issues) })
    return issues
  },

  previewSamples: () => {
    const { project, selected } = get()
    const node = selected?.kind === 'node' ? project.nodes.find((item) => item.id === selected.id) : project.nodes.find((item) => item.area_usages.includes('KillSample'))
    set({ samples: node ? sampleNode(node, 20) : [] })
  },

  clearPreview: () => {
    set({ samples: [], serviceMessage: '已清除采样预览' })
  },

  write: async () => {
    const { project, located } = get()
    const issues = validateProject(project)
    set({ issues, located: retainLocatedIfIssueExists(located, issues) })
    if (hasBlockingIssues(issues)) {
      set({ serviceMessage: '校验存在 ERROR，已阻止写入 Luban' })
      return
    }
    set({ busy: true })
    try {
      const result = await writeLuban(get().project)
      set({ writeLog: result.entries, issues: result.issues, located: retainLocatedIfIssueExists(get().located, result.issues), serviceMessage: 'Luban Excel 与工程快照写入完成', busy: false })
    } catch (error) {
      const maybeIssues = (error as Error & { issues?: ValidationIssue[] }).issues
      const nextIssues = maybeIssues ?? issues
      set({ serviceMessage: errorMessage(error), issues: nextIssues, located: retainLocatedIfIssueExists(get().located, nextIssues), busy: false })
    }
  },

  runExport: async () => {
    set({
      busy: true,
      genConfig: { operation: 'gen-config', status: 'running', exitCode: null, durationMs: 0, stdout: '', stderr: '' },
      serviceMessage: '导表运行中',
    })
    try {
      const output = await runGenConfig()
      set({ genConfig: output, serviceMessage: output.status === 'success' ? '导表成功' : '导表失败', busy: false })
    } catch (error) {
      set({
        genConfig: { operation: 'gen-config', status: 'failed', exitCode: null, durationMs: 0, stdout: '', stderr: errorMessage(error) },
        serviceMessage: '导表失败',
        busy: false,
      })
    }
  },

  updateLocal: async () => {
    set({
      busy: true,
      genConfig: { operation: 'update-local', status: 'running', exitCode: null, durationMs: 0, stdout: '', stderr: '' },
      serviceMessage: '正在更新本地前后端',
    })
    try {
      const output = await updateLocalConfig()
      set({ genConfig: output, serviceMessage: output.status === 'success' ? '本地前后端更新成功' : '本地前后端更新失败', busy: false })
    } catch (error) {
      set({
        genConfig: { operation: 'update-local', status: 'failed', exitCode: null, durationMs: 0, stdout: '', stderr: errorMessage(error) },
        serviceMessage: '本地前后端更新失败',
        busy: false,
      })
    }
  },

  addNode: (point, risk = false) => {
    commit((project) => {
      const id = uniqueId(project.nodes, risk ? 'RISK_POINT' : 'NODE')
      const node: MapNode = {
        id,
        map_id: project.map_id,
        name: id,
        zone: '未分区',
        site: 'None',
        node_type: risk ? 'Cover' : 'Lane',
        default_side: 'None',
        x: clamp(point.x),
        y: clamp(point.y),
        floor: 'Ground',
        area_usages: risk ? ['Risk'] : [],
        shape: risk ? 'Circle' : 'None',
        radius: risk ? 0.025 : null,
        points: [],
      }
      project.nodes.push(node)
      return { project, selected: { kind: 'node', id } }
    }, set, get)
  },

  moveNode: (id, point) => {
    commit((project) => {
      const node = project.nodes.find((item) => item.id === id)
      if (node) {
        node.x = clamp(point.x)
        node.y = clamp(point.y)
      }
      return { project }
    }, set, get)
  },

  updateNode: (id, patch) => {
    commit((project) => {
      const node = project.nodes.find((item) => item.id === id)
      if (node) Object.assign(node, patch)
      return { project }
    }, set, get)
  },

  updateEdge: (id, patch) => {
    commit((project) => {
      const edge = project.edges.find((item) => item.id === id)
      if (edge) Object.assign(edge, patch)
      return { project }
    }, set, get)
  },

  updateVisibility: (id, patch) => {
    commit((project) => {
      const visibility = project.visibility.find((item) => item.id === id)
      if (visibility) Object.assign(visibility, patch)
      return { project }
    }, set, get)
  },

  updateRoute: (id, patch) => {
    commit((project) => {
      const route = project.routes.find((item) => item.id === id)
      if (route) Object.assign(route, patch)
      return { project }
    }, set, get)
  },

  updateLayer: (layer, patch) => {
    commit((project) => {
      const current = project.layers[layer]
      if (current) project.layers[layer] = { ...current, ...patch }
      return { project }
    }, set, get)
  },

  deleteSelected: () => {
    const { selected } = get()
    if (!selected) return
    commit((project) => {
      if (selected.kind === 'node') {
        project.nodes = project.nodes.filter((item) => item.id !== selected.id)
        project.edges = project.edges.filter((item) => item.from !== selected.id && item.to !== selected.id)
        project.visibility = project.visibility.filter((item) => item.from !== selected.id && item.to !== selected.id)
        project.routes = project.routes.map((route) => ({ ...route, nodes: route.nodes.filter((id) => id !== selected.id) }))
      }
      if (selected.kind === 'edge') project.edges = project.edges.filter((item) => item.id !== selected.id)
      if (selected.kind === 'visibility') project.visibility = project.visibility.filter((item) => item.id !== selected.id)
      if (selected.kind === 'route') project.routes = project.routes.filter((item) => item.id !== selected.id)
      return { project, selected: null, routeDraft: [], rangeDraft: null }
    }, set, get)
  },

  duplicateSelected: () => {
    const { project, selected } = get()
    if (selected?.kind !== 'node') return
    const source = project.nodes.find((item) => item.id === selected.id)
    if (!source) return
    commit((draft) => {
      const id = uniqueId(draft.nodes, `${source.id}_COPY`)
      draft.nodes.push({ ...structuredClone(source), id, name: `${source.name} Copy`, x: clamp(source.x + 0.025), y: clamp(source.y + 0.025) })
      return { project: draft, selected: { kind: 'node', id } }
    }, set, get)
  },

  handleNodeToolClick: (nodeId) => {
    const { tool, pendingNodeId, routeDraft } = get()
    if (tool === 'select' || tool === 'node' || tool === 'risk') {
      set({ selected: { kind: 'node', id: nodeId } })
      return
    }
    if (tool === 'edge' && pendingNodeId && pendingNodeId !== nodeId) {
      commit((project) => {
        const id = uniqueId(project.edges, `EDGE_${pendingNodeId}_${nodeId}`)
        project.edges.push({ id, from: pendingNodeId, to: nodeId, base_time: 8, stamina_cost: 5, risk: 10, noise: 10, risk_points: [], intercept_nodes: [], bidirectional: true })
        return { project, selected: { kind: 'edge', id }, pendingNodeId: null }
      }, set, get)
      return
    }
    if (tool === 'visibility' && pendingNodeId && pendingNodeId !== nodeId) {
      commit((project) => {
        const id = uniqueId(project.visibility, `VIS_${pendingNodeId}_${nodeId}`)
        project.visibility.push({ id, from: pendingNodeId, to: nodeId, visible: true, range: 'Mid', angle_advantage: 'None', elevation: 'SameLevel', cover_modifier: 0, exposure_modifier: 0 })
        return { project, selected: { kind: 'visibility', id }, pendingNodeId: null }
      }, set, get)
      return
    }
    if (tool === 'route') {
      const next = routeDraft.includes(nodeId) ? routeDraft : [...routeDraft, nodeId]
      set({
        routeDraft: next,
        pendingNodeId: nodeId,
        selected: { kind: 'node', id: nodeId },
        serviceMessage: `路线草稿已加入 ${next.length} 个点位，按 Enter 或点击完成路线`,
      })
      return
    }
    set({ pendingNodeId: nodeId, selected: { kind: 'node', id: nodeId } })
  },

  setNodePointsFromText: (id, value) => {
    try {
      get().updateNode(id, { points: cellToPoints(value) })
    } catch {
      set({ serviceMessage: 'points 格式应为 x1,y1;x2,y2;x3,y3' })
    }
  },

  startRangeDraw: (nodeId, shape) => {
    const node = get().project.nodes.find((item) => item.id === nodeId)
    if (!node) return
    set({
      tool: 'select',
      selected: { kind: 'node', id: nodeId },
      pendingNodeId: null,
      routeDraft: [],
      rangeDraft: { nodeId, shape, points: [] },
      serviceMessage: shape === 'Circle' ? '点击圆形范围边界点以设置 radius' : '逐点点击多边形顶点，按 Enter 或点击完成范围',
    })
  },

  appendRangePoint: (point) => {
    const { rangeDraft, project } = get()
    if (!rangeDraft) return
    const normalized = { x: clamp(point.x), y: clamp(point.y) }
    if (rangeDraft.shape === 'Circle') {
      const node = project.nodes.find((item) => item.id === rangeDraft.nodeId)
      if (!node) return
      const radius = clampDistance(Math.hypot(normalized.x - node.x, normalized.y - node.y))
      commit((draft) => {
        const target = draft.nodes.find((item) => item.id === rangeDraft.nodeId)
        if (target) {
          target.shape = 'Circle'
          target.radius = radius
          target.points = []
        }
        return {
          project: draft,
          selected: { kind: 'node', id: rangeDraft.nodeId },
          rangeDraft: null,
          serviceMessage: `圆形范围已设置 radius=${radius}`,
        }
      }, set, get)
      return
    }
    set({
      rangeDraft: { ...rangeDraft, points: [...rangeDraft.points, normalized] },
      serviceMessage: `多边形范围顶点 ${rangeDraft.points.length + 1}，至少需要 3 个顶点`,
    })
  },

  finishRangeDraw: () => {
    const { rangeDraft } = get()
    if (!rangeDraft) return
    if (rangeDraft.shape === 'Circle') {
      set({ rangeDraft: null, serviceMessage: '已取消圆形范围绘制' })
      return
    }
    if (rangeDraft.points.length < 3) {
      set({ serviceMessage: '多边形范围至少需要 3 个顶点' })
      return
    }
    commit((project) => {
      const node = project.nodes.find((item) => item.id === rangeDraft.nodeId)
      if (node) {
        node.shape = 'Polygon'
        node.radius = null
        node.points = rangeDraft.points
      }
      return {
        project,
        selected: { kind: 'node', id: rangeDraft.nodeId },
        rangeDraft: null,
        serviceMessage: `多边形范围已写入 points=${rangeDraft.points.length}`,
      }
    }, set, get)
  },

  cancelRangeDraw: () => {
    set({ rangeDraft: null, serviceMessage: '已取消范围绘制' })
  },

  finishRouteDraft: () => {
    const { routeDraft } = get()
    if (routeDraft.length < 2) {
      set({ serviceMessage: '路线至少需要 2 个点位' })
      return
    }
    commit((project) => {
      const id = uniqueId(project.routes, `ROUTE_${routeDraft[0]}_${routeDraft[routeDraft.length - 1]}`)
      project.routes.push({ id, name: id, side: 'T', target_site: 'None', nodes: routeDraft, min_players: 1, max_players: 5, style_tags: [] })
      return {
        project,
        selected: { kind: 'route', id },
        pendingNodeId: null,
        routeDraft: [],
        serviceMessage: `路线已创建：${id}`,
      }
    }, set, get)
  },

  cancelRouteDraft: () => {
    set({ routeDraft: [], pendingNodeId: null, serviceMessage: '已取消路线草稿' })
  },

  undo: () => {
    const { history, project, future } = get()
    const previous = history.at(-1)
    if (!previous) return
    const issues = validateProject(previous)
    set({ project: previous, history: history.slice(0, -1), future: [structuredClone(project), ...future], issues, located: retainLocatedIfIssueExists(get().located, issues), routeDraft: [], rangeDraft: null })
  },

  redo: () => {
    const { future, project, history } = get()
    const next = future[0]
    if (!next) return
    const issues = validateProject(next)
    set({ project: next, history: [...history, structuredClone(project)], future: future.slice(1), issues, located: retainLocatedIfIssueExists(get().located, issues), routeDraft: [], rangeDraft: null })
  },

  centerView: () => {
    set((state) => ({ project: { ...state.project, viewport: { ...state.project.viewport, zoom: 1, offset_x: 0, offset_y: 0 } } }))
  },
}))

function commit(
  recipe: (project: MapProject) => Partial<EditorState> & { project: MapProject },
  set: (partial: Partial<EditorState>) => void,
  get: () => EditorState,
): void {
  const current = get()
  const draft = structuredClone(current.project)
  const next = recipe(draft)
  const issues = validateProject(next.project)
  const nextLocated = Object.prototype.hasOwnProperty.call(next, 'located')
    ? (next.located ?? null)
    : retainLocatedIfIssueExists(current.located, issues)
  set({
    ...next,
    history: [...current.history, structuredClone(current.project)].slice(-80),
    future: [],
    issues,
    located: nextLocated,
  })
}

function retainLocatedIfIssueExists(located: SelectedObject, issues: ValidationIssue[]): SelectedObject {
  if (!located) return null
  return issues.some((issueItem) => issueItem.object?.kind === located.kind && issueItem.object.id === located.id) ? located : null
}

function uniqueId<T extends { id: string }>(items: T[], prefix: string): string {
  const base = prefix.toUpperCase().replace(/[^A-Z0-9_]/g, '_')
  let index = 1
  let id = base
  const ids = new Set(items.map((item) => item.id))
  while (ids.has(id)) {
    index += 1
    id = `${base}_${index}`
  }
  return id
}

function clamp(value: number): number {
  return Math.min(1, Math.max(0, Math.round(value * 10000) / 10000))
}

function clampDistance(value: number): number {
  return Math.min(1, Math.max(0.001, Math.round(value * 10000) / 10000))
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}

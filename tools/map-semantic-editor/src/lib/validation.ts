import type { MapProject, Point, SelectedObject } from './model'

export type ValidationSeverity = 'ERROR' | 'WARN' | 'INFO'

export interface ValidationIssue {
  id: string
  severity: ValidationSeverity
  object: SelectedObject
  field: string
  message: string
}

export function validateProject(project: MapProject): ValidationIssue[] {
  const issues: ValidationIssue[] = []
  const nodeIds = new Set<string>()
  const edgeIds = new Set<string>()
  const visibilityIds = new Set<string>()
  const routeIds = new Set<string>()

  for (const node of project.nodes) {
    if (nodeIds.has(node.id)) {
      issues.push(issue('ERROR', { kind: 'node', id: node.id }, 'id', `MapNode ${node.id} ID 重复`))
    }
    nodeIds.add(node.id)

    if (!inUnitRange(node.x) || !inUnitRange(node.y)) {
      issues.push(issue('ERROR', { kind: 'node', id: node.id }, 'x/y', `MapNode ${node.id} 坐标必须在 0..1 内`))
    }
    if (node.shape === 'Circle' && (!node.radius || node.radius <= 0)) {
      issues.push(issue('ERROR', { kind: 'node', id: node.id }, 'radius', `MapNode ${node.id} 圆形范围缺少合法半径`))
    }
    if (node.shape === 'Polygon') {
      if (node.points.length < 3) {
        issues.push(issue('ERROR', { kind: 'node', id: node.id }, 'points', `MapNode ${node.id} 多边形顶点少于 3 个`))
      } else if (!isPolygonValid(node.points)) {
        issues.push(issue('ERROR', { kind: 'node', id: node.id }, 'points', `MapNode ${node.id} 多边形存在自交或非法顶点`))
      }
    }
    if (node.area_usages.includes('KillSample') && node.shape === 'None') {
      issues.push(issue('WARN', { kind: 'node', id: node.id }, 'shape', `MapNode ${node.id} 标记 KillSample 但没有几何范围，导出后会回退 x/y`))
    }
  }

  for (const edge of project.edges) {
    if (edgeIds.has(edge.id)) {
      issues.push(issue('ERROR', { kind: 'edge', id: edge.id }, 'id', `MapEdge ${edge.id} ID 重复`))
    }
    edgeIds.add(edge.id)
    for (const [field, value] of [
      ['from', edge.from],
      ['to', edge.to],
      ...edge.risk_points.map((id) => ['risk_points', id] as const),
      ...edge.intercept_nodes.map((id) => ['intercept_nodes', id] as const),
    ] as const) {
      if (!nodeIds.has(value)) {
        issues.push(issue('ERROR', { kind: 'edge', id: edge.id }, field, `MapEdge ${edge.id}.${field} 引用不存在的 MapNode ${value}`))
      }
    }
  }

  for (const visibility of project.visibility) {
    if (visibilityIds.has(visibility.id)) {
      issues.push(issue('ERROR', { kind: 'visibility', id: visibility.id }, 'id', `Visibility ${visibility.id} ID 重复`))
    }
    visibilityIds.add(visibility.id)
    for (const field of ['from', 'to'] as const) {
      if (!nodeIds.has(visibility[field])) {
        issues.push(issue('ERROR', { kind: 'visibility', id: visibility.id }, field, `Visibility ${visibility.id}.${field} 引用不存在的 MapNode ${visibility[field]}`))
      }
    }
    if (visibility.elevation === 'HeightBlocked' && visibility.visible) {
      issues.push(issue('WARN', { kind: 'visibility', id: visibility.id }, 'elevation', `Visibility ${visibility.id} 标记 HeightBlocked 但 visible=true`))
    }
  }

  for (const route of project.routes) {
    if (routeIds.has(route.id)) {
      issues.push(issue('ERROR', { kind: 'route', id: route.id }, 'id', `Route ${route.id} ID 重复`))
    }
    routeIds.add(route.id)
    for (const nodeId of route.nodes) {
      if (!nodeIds.has(nodeId)) {
        issues.push(issue('ERROR', { kind: 'route', id: route.id }, 'nodes', `Route ${route.id}.nodes 引用不存在的 MapNode ${nodeId}`))
      }
    }
    for (let index = 0; index < route.nodes.length - 1; index += 1) {
      const from = route.nodes[index]
      const to = route.nodes[index + 1]
      const connection = routeConnection(project, from, to)
      if (connection === 'reverse_blocked') {
        issues.push(issue('ERROR', { kind: 'route', id: route.id }, 'nodes', `Route ${route.id} 方向冲突：${from} -> ${to} 只存在反向单向路径`))
      } else if (connection === 'missing') {
        issues.push(issue('ERROR', { kind: 'route', id: route.id }, 'nodes', `Route ${route.id} 断裂：${from} -> ${to} 没有可达路径`))
      }
    }
  }

  const connectedNodes = new Set<string>()
  for (const edge of project.edges) {
    if (nodeIds.has(edge.from)) connectedNodes.add(edge.from)
    if (nodeIds.has(edge.to)) connectedNodes.add(edge.to)
  }
  for (const node of project.nodes) {
    if (isCriticalNode(node.node_type) && !connectedNodes.has(node.id)) {
      issues.push(issue('WARN', { kind: 'node', id: node.id }, 'edges', `MapNode ${node.id} 是关键节点但没有连接任何 MapEdge`))
    }
  }

  if (issues.length === 0) {
    issues.push(issue('INFO', null, 'project', '校验通过，可以写入 Luban 配表'))
  }

  return issues
}

export function hasBlockingIssues(issues: ValidationIssue[]): boolean {
  return issues.some((item) => item.severity === 'ERROR')
}

function issue(
  severity: ValidationSeverity,
  object: SelectedObject,
  field: string,
  message: string,
): ValidationIssue {
  return {
    id: `${severity}:${object?.kind ?? 'project'}:${object?.id ?? 'root'}:${field}:${message}`,
    severity,
    object,
    field,
    message,
  }
}

function inUnitRange(value: number): boolean {
  return Number.isFinite(value) && value >= 0 && value <= 1
}

function routeConnection(project: MapProject, from: string, to: string): 'direct' | 'reverse_bidirectional' | 'reverse_blocked' | 'missing' {
  let hasReverseBlocked = false
  for (const edge of project.edges) {
    if (edge.from === from && edge.to === to) return 'direct'
    if (edge.from === to && edge.to === from) {
      if (edge.bidirectional) return 'reverse_bidirectional'
      hasReverseBlocked = true
    }
  }
  return hasReverseBlocked ? 'reverse_blocked' : 'missing'
}

function isCriticalNode(nodeType: string): boolean {
  return nodeType === 'Spawn' || nodeType === 'Site' || nodeType === 'Connector'
}

function isPolygonValid(points: Point[]): boolean {
  if (points.some((point) => !inUnitRange(point.x) || !inUnitRange(point.y))) return false
  for (let i = 0; i < points.length; i += 1) {
    for (let j = i + 1; j < points.length; j += 1) {
      const a1 = points[i]
      const a2 = points[(i + 1) % points.length]
      const b1 = points[j]
      const b2 = points[(j + 1) % points.length]
      const adjacent = Math.abs(i - j) <= 1 || (i === 0 && j === points.length - 1)
      if (!adjacent && segmentsIntersect(a1, a2, b1, b2)) return false
    }
  }
  return true
}

function segmentsIntersect(a1: Point, a2: Point, b1: Point, b2: Point): boolean {
  const d1 = direction(a1, a2, b1)
  const d2 = direction(a1, a2, b2)
  const d3 = direction(b1, b2, a1)
  const d4 = direction(b1, b2, a2)
  return d1 * d2 < 0 && d3 * d4 < 0
}

function direction(a: Point, b: Point, c: Point): number {
  return (c.x - a.x) * (b.y - a.y) - (b.x - a.x) * (c.y - a.y)
}

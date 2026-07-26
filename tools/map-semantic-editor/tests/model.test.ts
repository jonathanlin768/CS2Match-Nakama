import { describe, expect, it } from 'vitest'
import { toExportTables } from '../src/lib/exportTables'
import { parseProject } from '../src/lib/model'
import { pointInPolygon, sampleNode } from '../src/lib/sampling'
import { createSampleProject } from '../src/lib/sampleProject'
import { hasBlockingIssues, validateProject } from '../src/lib/validation'

describe('map semantic editor model', () => {
  it('parses the Dust2 sample project', () => {
    const project = parseProject(createSampleProject())
    expect(project.map_id).toBe('de_dust2')
    expect(project.nodes.some((node) => node.id === 'A_SITE')).toBe(true)
  })

  it('exports all Luban map semantic tables', () => {
    const tables = toExportTables(createSampleProject())
    expect(tables.map((table) => table.tableName)).toEqual([
      'tb_route_template',
      'tb_scenario',
      'tb_map_tag',
      'tb_encounter_modifier',
      'tb_map_node',
      'tb_map_edge',
      'tb_visibility',
      'tb_route',
      'tb_combat_const',
    ])
    const mapNode = tables.find((table) => table.tableName === 'tb_map_node')
    expect(mapNode?.fileName).toBe('#MapNode.xlsx')
    expect(mapNode?.fields.map((field) => field.key)).toContain('area_usages')
    expect(mapNode?.fields.map((field) => field.key)).toContain('points')
    const mapEdge = tables.find((table) => table.tableName === 'tb_map_edge')
    expect(mapEdge?.fields.map((field) => field.key)).toContain('from_node')
    expect(mapEdge?.fields.map((field) => field.key)).toContain('to_node')
    expect(mapEdge?.fields.map((field) => field.key)).not.toContain('from')
    const visibility = tables.find((table) => table.tableName === 'tb_visibility')
    expect(visibility?.fields.map((field) => field.key)).toContain('from_node')
    expect(visibility?.fields.map((field) => field.key)).toContain('to_node')
    expect(visibility?.fields.map((field) => field.key)).not.toContain('from')
  })

  it('validates missing route references', () => {
    const project = createSampleProject()
    project.routes[0].nodes.push('MISSING_NODE')
    const issues = validateProject(project)
    expect(hasBlockingIssues(issues)).toBe(true)
    expect(issues.some((issue) => issue.message.includes('MISSING_NODE'))).toBe(true)
  })

  it('blocks route gaps and one-way edge direction conflicts', () => {
    const missingEdgeProject = createSampleProject()
    missingEdgeProject.edges = missingEdgeProject.edges.filter((edge) => edge.id !== 'EDGE_A_SITE')
    const gapIssues = validateProject(missingEdgeProject)
    expect(gapIssues.some((issue) => issue.severity === 'ERROR' && issue.message.includes('没有可达路径'))).toBe(true)

    const reversedProject = createSampleProject()
    const edge = reversedProject.edges.find((item) => item.id === 'EDGE_A_SITE')
    expect(edge).toBeDefined()
    Object.assign(edge!, { from: 'A_SITE', to: 'A_LONG', bidirectional: false })
    const directionIssues = validateProject(reversedProject)
    expect(directionIssues.some((issue) => issue.severity === 'ERROR' && issue.message.includes('方向冲突'))).toBe(true)
  })

  it('warns when critical map nodes are isolated', () => {
    const project = createSampleProject()
    project.nodes.push({
      id: 'ISOLATED_SITE',
      map_id: project.map_id,
      name: 'Isolated Site',
      zone: 'A Area',
      site: 'A',
      node_type: 'Site',
      default_side: 'CT',
      x: 0.82,
      y: 0.18,
      floor: 'Ground',
      area_usages: [],
      shape: 'None',
      radius: null,
      points: [],
    })
    const issues = validateProject(project)
    expect(issues.some((issue) => issue.severity === 'WARN' && issue.message.includes('ISOLATED_SITE'))).toBe(true)
  })

  it('keeps Risk usage nodes in the map node collection', () => {
    const project = createSampleProject()
    const riskNode = project.nodes.find((item) => item.area_usages.includes('Risk'))
    expect(riskNode).toBeDefined()
    expect(project.nodes.map((node) => node.id)).toContain(riskNode!.id)
    expect(project.layers).not.toHaveProperty('risks')
  })

  it('samples polygon nodes inside their range', () => {
    const project = createSampleProject()
    const node = project.nodes.find((item) => item.id === 'A_SITE')
    expect(node).toBeDefined()
    const samples = sampleNode(node!, 16, 7)
    expect(samples.every((sample) => pointInPolygon(sample, node!.points))).toBe(true)
  })

  it('samples circle nodes inside radius', () => {
    const project = createSampleProject()
    const node = project.nodes.find((item) => item.id === 'B_SITE')
    expect(node?.shape).toBe('Circle')
    const samples = sampleNode(node!, 16, 9)
    expect(samples.every((sample) => Math.hypot(sample.x - node!.x, sample.y - node!.y) <= (node!.radius ?? 0) + 0.00001)).toBe(true)
  })

  it('falls back near x/y for nodes without geometry', () => {
    const project = createSampleProject()
    const node = project.nodes.find((item) => item.id === 'LONG_DOOR')
    expect(node?.shape).toBe('None')
    const samples = sampleNode(node!, 8, 11)
    expect(samples.every((sample) => Math.abs(sample.x - node!.x) < 0.01 && Math.abs(sample.y - node!.y) < 0.01)).toBe(true)
  })
})

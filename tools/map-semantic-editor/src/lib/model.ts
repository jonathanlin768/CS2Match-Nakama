import { z } from 'zod'

export const PROJECT_SCHEMA_VERSION = 1

export const toolModes = ['select', 'node', 'edge', 'visibility', 'route', 'risk'] as const
export const nodeShapes = ['None', 'Circle', 'Polygon'] as const
export const nodeUsages = ['KillSample', 'Plant', 'Control', 'Encounter', 'Sound', 'Risk'] as const
export const sides = ['T', 'CT', 'Both', 'None'] as const
export const sites = ['A', 'B', 'None'] as const
export const nodeTypes = ['Spawn', 'Lane', 'Cover', 'Site', 'Connector'] as const
export const floors = ['Ground', 'Upper', 'Ramp', 'Unknown'] as const

export type ToolMode = (typeof toolModes)[number]
export type NodeShape = (typeof nodeShapes)[number]
export type NodeUsage = (typeof nodeUsages)[number]
export type Side = (typeof sides)[number]
export type Site = (typeof sites)[number]
export type NodeType = (typeof nodeTypes)[number]
export type Floor = (typeof floors)[number]

export type Point = z.infer<typeof pointSchema>
export type MapNode = z.infer<typeof mapNodeSchema>
export type MapEdge = z.infer<typeof mapEdgeSchema>
export type Visibility = z.infer<typeof visibilitySchema>
export type Route = z.infer<typeof routeSchema>
export type MapProject = z.infer<typeof mapProjectSchema>
export type RangeDraft = {
  nodeId: string
  shape: 'Circle' | 'Polygon'
  points: Point[]
}
export type SelectedObject =
  | { kind: 'node'; id: string }
  | { kind: 'edge'; id: string }
  | { kind: 'visibility'; id: string }
  | { kind: 'route'; id: string }
  | null

export const pointSchema = z.object({
  x: z.number().min(0).max(1),
  y: z.number().min(0).max(1),
})

export const mapNodeSchema = z.object({
  id: z.string().min(1),
  map_id: z.string().min(1),
  name: z.string().min(1),
  zone: z.string().default(''),
  site: z.enum(sites).default('None'),
  node_type: z.enum(nodeTypes).default('Lane'),
  default_side: z.enum(sides).default('None'),
  x: z.number().min(0).max(1),
  y: z.number().min(0).max(1),
  floor: z.enum(floors).default('Ground'),
  area_usages: z.array(z.enum(nodeUsages)).default([]),
  shape: z.enum(nodeShapes).default('None'),
  radius: z.number().min(0).max(1).nullable().default(null),
  points: z.array(pointSchema).default([]),
})

export const mapEdgeSchema = z.object({
  id: z.string().min(1),
  from: z.string().min(1),
  to: z.string().min(1),
  base_time: z.number().int().min(0).default(8),
  stamina_cost: z.number().int().min(0).default(5),
  risk: z.number().int().min(0).max(100).default(10),
  noise: z.number().int().min(0).max(100).default(10),
  risk_points: z.array(z.string()).default([]),
  intercept_nodes: z.array(z.string()).default([]),
  bidirectional: z.boolean().default(true),
})

export const visibilitySchema = z.object({
  id: z.string().min(1),
  from: z.string().min(1),
  to: z.string().min(1),
  visible: z.boolean().default(true),
  range: z.enum(['Close', 'Mid', 'Long']).default('Mid'),
  angle_advantage: z.enum(['T', 'CT', 'None']).default('None'),
  elevation: z.enum(['HighToLow', 'LowToHigh', 'SameLevel', 'HeightBlocked']).default('SameLevel'),
  cover_modifier: z.number().int().default(0),
  exposure_modifier: z.number().int().default(0),
})

export const routeSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  side: z.enum(['T', 'CT']).default('T'),
  target_site: z.enum(sites).default('None'),
  nodes: z.array(z.string()).default([]),
  min_players: z.number().int().min(1).max(5).default(1),
  max_players: z.number().int().min(1).max(5).default(5),
  style_tags: z.array(z.string()).default([]),
})

export const routeTemplateSchema = z.object({
  id: z.string().min(1),
  map_id: z.string().min(1),
  target_site: z.enum(sites).default('None'),
  tempo: z.enum(['Fast', 'Default', 'Slow', 'Late']).default('Default'),
  recommended_min: z.number().int().min(1).max(5).default(1),
  recommended_max: z.number().int().min(1).max(5).default(5),
  required_roles: z.array(z.string()).default([]),
  key_attributes: z.string().default(''),
  scenario_ids: z.array(z.string()).default([]),
  map_tag_ids: z.array(z.string()).default([]),
  success_next_phase: z.string().default(''),
  failure_fallbacks: z.array(z.string()).default([]),
})

export const scenarioSchema = z.object({
  id: z.string().min(1),
  route: z.string().default('A_Long'),
  phase: z.string().default('OpeningDuel'),
  range: z.enum(['Close', 'Mid', 'Long']).default('Mid'),
  site: z.enum(sites).default('None'),
  tempo: z.string().default('SlowDefault'),
  posture: z.string().default('Even'),
  utility_context: z.string().default('Even'),
  map_tag_ids: z.array(z.string()).default([]),
  base_time_cost: z.number().int().min(0).default(0),
  base_weight: z.number().int().default(0),
})

export const mapTagSchema = z.object({
  id: z.string().min(1),
  map_id: z.string().min(1),
  category: z.string().default('Range'),
  value: z.string().min(1),
  side: z.enum(sides).default('Both'),
  weight: z.number().int().default(0),
  reason_code: z.string().default(''),
  description: z.string().default(''),
})

export const encounterModifierSchema = z.object({
  id: z.string().min(1),
  scenario_id: z.string().min(1),
  factor: z.string().min(1),
  side: z.enum(sides).default('Both'),
  attribute: z.string().default(''),
  weight: z.number().int().default(0),
  reason_code: z.string().default(''),
})

export const combatConstSchema = z.object({
  key: z.string().min(1),
  category: z.string().default('Decision'),
  value_type: z.enum(['Int', 'Float', 'Bool', 'String']).default('Int'),
  value: z.string().default('0'),
  min_value: z.string().default(''),
  max_value: z.string().default(''),
  unit: z.string().default('none'),
  description: z.string().default(''),
})

export const layerStateSchema = z.object({
  visible: z.boolean(),
  locked: z.boolean(),
  color: z.string(),
})

export const mapProjectSchema = z.object({
  schema_version: z.literal(PROJECT_SCHEMA_VERSION),
  map_id: z.string().min(1),
  name: z.string().min(1),
  version: z.string().min(1),
  radar_image: z.string().min(1),
  coordinate_normalization: z.object({
    min_x: z.number().default(0),
    max_x: z.number().default(1),
    min_y: z.number().default(0),
    max_y: z.number().default(1),
  }),
  layers: z.record(z.string(), layerStateSchema),
  viewport: z.object({
    zoom: z.number().positive(),
    offset_x: z.number(),
    offset_y: z.number(),
    snap: z.boolean(),
  }),
  nodes: z.array(mapNodeSchema),
  edges: z.array(mapEdgeSchema),
  visibility: z.array(visibilitySchema),
  routes: z.array(routeSchema),
  route_templates: z.array(routeTemplateSchema).default([]),
  scenarios: z.array(scenarioSchema).default([]),
  map_tags: z.array(mapTagSchema).default([]),
  encounter_modifiers: z.array(encounterModifierSchema).default([]),
  combat_consts: z.array(combatConstSchema).default([]),
})

export function parseProject(value: unknown): MapProject {
  return mapProjectSchema.parse(value)
}

export function pointsToCell(points: Point[]): string {
  return points.map((point) => `${round(point.x)},${round(point.y)}`).join(';')
}

export function cellToPoints(value: string): Point[] {
  if (!value.trim()) return []
  return value.split(';').map((item) => {
    const [x, y] = item.split(',').map(Number)
    return pointSchema.parse({ x, y })
  })
}

export function listToCell(values: string[]): string {
  return values.join(',')
}

export function round(value: number): number {
  return Math.round(value * 10000) / 10000
}

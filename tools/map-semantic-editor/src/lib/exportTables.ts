import type { MapProject } from './model'
import { allocationsToCell, listToCell, pointsToCell, round } from './model'

export interface ExportField {
  key: string
  type: string
  group?: string
  comment: string
}

export type ExportRow = Record<string, string | number | boolean | null>

export interface ExportTable {
  tableName: string
  fileName: string
  key: ImportTableKey
  fields: ExportField[]
  rows: ExportRow[]
}

export type ImportTableKey =
  | 'route_templates'
  | 'scenarios'
  | 'map_tags'
  | 'encounter_modifiers'
  | 'nodes'
  | 'edges'
  | 'visibility'
  | 'routes'
  | 'combat_consts'

export interface LubanTableSpec {
  tableName: string
  fileName: string
  key: ImportTableKey
}

const group = 'c,s,e'

interface TableSpec extends LubanTableSpec {
  fields: ExportField[]
  toRows: (project: MapProject) => ExportRow[]
}

const routeTemplateFields: ExportField[] = [
  field('id', 'string', '模板ID', ''),
  field('map_id', 'string', '地图ID'),
  field('side', 'string', '阵营'),
  field('target_site', 'string', '目标包点'),
  field('tempo', 'string', '节奏'),
  field('recommended_min', 'int', '推荐最少人数'),
  field('recommended_max', 'int', '推荐最多人数'),
  field('required_roles', '(list#sep=,),string', '关键角色'),
  field('key_attributes', 'string', '属性权重'),
  field('route_ids', '(list#sep=,),string', '路线ID'),
  field('route_allocations', '(map#sep=,),string,int', '路线人数分配'),
  field('scenario_ids', '(list#sep=,),string', '可生成场景'),
  field('map_tag_ids', '(list#sep=,),string', '地图标签'),
  field('common_ct_setup_ids', '(list#sep=,),string', '常见CT配置先验'),
  field('success_next_phase', 'string', '成功后阶段'),
  field('failure_fallbacks', '(list#sep=,),string', '失败候选策略'),
]

const scenarioFields: ExportField[] = [
  field('id', 'string', '场景ID', ''),
  field('route', 'string', '路线类型'),
  field('phase', 'string', '阶段'),
  field('range', 'string', '距离'),
  field('site', 'string', '包点'),
  field('tempo', 'string', '节奏'),
  field('posture', 'string', '姿态'),
  field('utility_context', 'string', '道具上下文'),
  field('map_tag_ids', '(list#sep=,),string', '地图标签'),
  field('base_time_cost', 'int', '基础耗时'),
  field('base_weight', 'int', '基础权重'),
]

const mapTagFields: ExportField[] = [
  field('id', 'string', '标签ID', ''),
  field('map_id', 'string', '地图ID'),
  field('category', 'string', '类别'),
  field('value', 'string', '标签值'),
  field('side', 'string', '阵营'),
  field('weight', 'int', '权重'),
  field('reason_code', 'string', '原因码'),
  field('description', 'string', '说明'),
]

const encounterModifierFields: ExportField[] = [
  field('id', 'string', '修正ID', ''),
  field('scenario_id', 'string', '场景ID'),
  field('factor', 'string', '修正因子'),
  field('side', 'string', '阵营'),
  field('attribute', 'string', '关联属性'),
  field('weight', 'int', '权重'),
  field('reason_code', 'string', '原因码'),
]

const mapNodeFields: ExportField[] = [
  field('id', 'string', '节点ID', ''),
  field('map_id', 'string', '地图ID'),
  field('name', 'string', '显示名称'),
  field('zone', 'string', '宏观区域'),
  field('site', 'string', '包点'),
  field('node_type', 'string', '节点类型'),
  field('default_side', 'string', '默认优势方'),
  field('x', 'float', '归一化X'),
  field('y', 'float', '归一化Y'),
  field('floor', 'string', '楼层'),
  field('area_usages', '(list#sep=,),string', '节点范围用途'),
  field('shape', 'string', '节点范围形状'),
  field('radius', 'float', '圆形半径'),
  field('points', 'string', '多边形顶点 x1,y1;x2,y2'),
]

const mapEdgeFields: ExportField[] = [
  field('id', 'string', '路径ID', ''),
  field('from_node', 'string', '起点'),
  field('to_node', 'string', '终点'),
  field('base_time', 'int', '基础移动时间'),
  field('stamina_cost', 'int', '体能消耗'),
  field('risk', 'int', '转移风险'),
  field('noise', 'int', '暴露概率'),
  field('risk_points', '(list#sep=,),string', '风险热点'),
  field('intercept_nodes', '(list#sep=,),string', '拦截候选点'),
  field('bidirectional', 'bool', '是否双向'),
]

const visibilityFields: ExportField[] = [
  field('id', 'string', '视野ID', ''),
  field('from_node', 'string', '观察点'),
  field('to_node', 'string', '被观察点'),
  field('visible', 'bool', '是否可见'),
  field('range', 'string', '距离'),
  field('angle_advantage', 'string', '角度优势'),
  field('elevation', 'string', '高低关系'),
  field('cover_modifier', 'int', '掩体修正'),
  field('exposure_modifier', 'int', '暴露修正'),
]

const routeFields: ExportField[] = [
  field('id', 'string', '路线ID', ''),
  field('name', 'string', '显示名称'),
  field('side', 'string', '阵营'),
  field('target_site', 'string', '目标包点'),
  field('nodes', '(list#sep=,),string', '节点序列'),
  field('min_players', 'int', '最少人数'),
  field('max_players', 'int', '最多人数'),
  field('style_tags', '(list#sep=,),string', '路线标签'),
]

const combatConstFields: ExportField[] = [
  field('key', 'string', '常量键', ''),
  field('category', 'string', '类别'),
  field('value_type', 'string', '值类型'),
  field('value', 'string', '原始值'),
  field('min_value', 'string', '下限'),
  field('max_value', 'string', '上限'),
  field('unit', 'string', '单位'),
  field('description', 'string', '说明'),
]

const tableSpecs: TableSpec[] = [
  {
    tableName: 'tb_route_template',
    fileName: '#RouteTemplate.xlsx',
    key: 'route_templates',
    fields: routeTemplateFields,
    toRows: (project) => project.route_templates.map((item) => ({
      ...item,
      required_roles: listToCell(item.required_roles),
      route_ids: listToCell(item.route_ids),
      route_allocations: allocationsToCell(item.route_allocations),
      scenario_ids: listToCell(item.scenario_ids),
      map_tag_ids: listToCell(item.map_tag_ids),
      common_ct_setup_ids: listToCell(item.common_ct_setup_ids),
      failure_fallbacks: listToCell(item.failure_fallbacks),
    })),
  },
  {
    tableName: 'tb_scenario',
    fileName: '#Scenario.xlsx',
    key: 'scenarios',
    fields: scenarioFields,
    toRows: (project) => project.scenarios.map((item) => ({
      ...item,
      map_tag_ids: listToCell(item.map_tag_ids),
    })),
  },
  {
    tableName: 'tb_map_tag',
    fileName: '#MapTag.xlsx',
    key: 'map_tags',
    fields: mapTagFields,
    toRows: (project) => project.map_tags,
  },
  {
    tableName: 'tb_encounter_modifier',
    fileName: '#EncounterModifier.xlsx',
    key: 'encounter_modifiers',
    fields: encounterModifierFields,
    toRows: (project) => project.encounter_modifiers,
  },
  {
    tableName: 'tb_map_node',
    fileName: '#MapNode.xlsx',
    key: 'nodes',
    fields: mapNodeFields,
    toRows: (project) => project.nodes.map((node) => ({
      ...node,
      x: round(node.x),
      y: round(node.y),
      area_usages: listToCell(node.area_usages),
      radius: node.radius ?? '',
      points: pointsToCell(node.points),
    })),
  },
  {
    tableName: 'tb_map_edge',
    fileName: '#MapEdge.xlsx',
    key: 'edges',
    fields: mapEdgeFields,
    toRows: (project) => project.edges.map((edge) => ({
      id: edge.id,
      from_node: edge.from,
      to_node: edge.to,
      base_time: edge.base_time,
      stamina_cost: edge.stamina_cost,
      risk: edge.risk,
      noise: edge.noise,
      risk_points: listToCell(edge.risk_points),
      intercept_nodes: listToCell(edge.intercept_nodes),
      bidirectional: edge.bidirectional,
    })),
  },
  {
    tableName: 'tb_visibility',
    fileName: '#Visibility.xlsx',
    key: 'visibility',
    fields: visibilityFields,
    toRows: (project) => project.visibility.map((visibility) => ({
      id: visibility.id,
      from_node: visibility.from,
      to_node: visibility.to,
      visible: visibility.visible,
      range: visibility.range,
      angle_advantage: visibility.angle_advantage,
      elevation: visibility.elevation,
      cover_modifier: visibility.cover_modifier,
      exposure_modifier: visibility.exposure_modifier,
    })),
  },
  {
    tableName: 'tb_route',
    fileName: '#Route.xlsx',
    key: 'routes',
    fields: routeFields,
    toRows: (project) => project.routes.map((route) => ({
      ...route,
      nodes: listToCell(route.nodes),
      style_tags: listToCell(route.style_tags),
    })),
  },
  {
    tableName: 'tb_combat_const',
    fileName: '#CombatConst.xlsx',
    key: 'combat_consts',
    fields: combatConstFields,
    toRows: (project) => project.combat_consts,
  },
]

export const LUBAN_TABLE_SPECS: LubanTableSpec[] = tableSpecs.map(({ tableName, fileName, key }) => ({ tableName, fileName, key }))

export function toExportTables(project: MapProject): ExportTable[] {
  return tableSpecs.map((spec) => ({
    tableName: spec.tableName,
    fileName: spec.fileName,
    key: spec.key,
    fields: spec.fields,
    rows: spec.toRows(project),
  }))
}

function field(key: string, type: string, comment: string, exportGroup = group): ExportField {
  return { key, type, group: exportGroup, comment }
}

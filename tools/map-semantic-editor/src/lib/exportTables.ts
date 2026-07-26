import type { MapProject } from './model'
import { listToCell, pointsToCell, round } from './model'

export interface ExportField {
  key: string
  type: string
  group?: string
  comment: string
}

export interface ExportTable {
  tableName: string
  fileName: string
  fields: ExportField[]
  rows: Record<string, string | number | boolean | null>[]
}

const group = 'c,s,e'

export function toExportTables(project: MapProject): ExportTable[] {
  return [
    table('tb_route_template', '#RouteTemplate.xlsx', routeTemplateFields, project.route_templates.map((item) => ({
      ...item,
      required_roles: listToCell(item.required_roles),
      scenario_ids: listToCell(item.scenario_ids),
      map_tag_ids: listToCell(item.map_tag_ids),
      failure_fallbacks: listToCell(item.failure_fallbacks),
    }))),
    table('tb_scenario', '#Scenario.xlsx', scenarioFields, project.scenarios.map((item) => ({
      ...item,
      map_tag_ids: listToCell(item.map_tag_ids),
    }))),
    table('tb_map_tag', '#MapTag.xlsx', mapTagFields, project.map_tags),
    table('tb_encounter_modifier', '#EncounterModifier.xlsx', encounterModifierFields, project.encounter_modifiers),
    table('tb_map_node', '#MapNode.xlsx', mapNodeFields, project.nodes.map((node) => ({
      ...node,
      x: round(node.x),
      y: round(node.y),
      area_usages: listToCell(node.area_usages),
      radius: node.radius ?? '',
      points: pointsToCell(node.points),
    }))),
    table('tb_map_edge', '#MapEdge.xlsx', mapEdgeFields, project.edges.map((edge) => ({
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
    }))),
    table('tb_visibility', '#Visibility.xlsx', visibilityFields, project.visibility.map((visibility) => ({
      id: visibility.id,
      from_node: visibility.from,
      to_node: visibility.to,
      visible: visibility.visible,
      range: visibility.range,
      angle_advantage: visibility.angle_advantage,
      elevation: visibility.elevation,
      cover_modifier: visibility.cover_modifier,
      exposure_modifier: visibility.exposure_modifier,
    }))),
    table('tb_route', '#Route.xlsx', routeFields, project.routes.map((route) => ({
      ...route,
      nodes: listToCell(route.nodes),
      style_tags: listToCell(route.style_tags),
    }))),
    table('tb_combat_const', '#CombatConst.xlsx', combatConstFields, project.combat_consts),
  ]
}

function table(
  tableName: string,
  fileName: string,
  fields: ExportField[],
  rows: Record<string, string | number | boolean | null>[],
): ExportTable {
  return { tableName, fileName, fields, rows }
}

const routeTemplateFields: ExportField[] = [
  field('id', 'string', '模板ID', ''),
  field('map_id', 'string', '地图ID'),
  field('target_site', 'string', '目标包点'),
  field('tempo', 'string', '节奏'),
  field('recommended_min', 'int', '推荐最少人数'),
  field('recommended_max', 'int', '推荐最多人数'),
  field('required_roles', '(list#sep=,),string', '关键角色'),
  field('key_attributes', 'string', '属性权重'),
  field('scenario_ids', '(list#sep=,),string', '可生成场景'),
  field('map_tag_ids', '(list#sep=,),string', '地图标签'),
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

function field(key: string, type: string, comment: string, exportGroup = group): ExportField {
  return { key, type, group: exportGroup, comment }
}

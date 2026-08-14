import type { LubanTableDocument } from './luban'

const fieldLabels: Record<string, string> = {
  id: '唯一标识',
  version: '版本',
  enabled: '是否启用',
  budget: '阵容预算',
  rosterSize: '阵容人数',
  mapId: '地图 ID',
  tier5PlayerIds: '5 费候选选手',
  tier4PlayerIds: '4 费候选选手',
  tier3PlayerIds: '3 费候选选手',
  tier2PlayerIds: '2 费候选选手',
  tier1PlayerIds: '1 费候选选手',
  opponentTeamId: '对手战队',
  opponentPlayerIds: '对手阵容',
  name: '显示名称',
  teamId: '所属战队',
  nationality: '国籍',
  positions: '位置标签',
  rarity: '稀有度',
  entry: '突破',
  aim: '精准',
  trade: '补枪',
  clutch: '残局',
  firepower: '火力',
  gamesense: '意识',
  reaction: '反应',
  positioning: '站位',
  awareness: '感知',
  teamplay: '团队配合',
  utility: '道具',
  composure: '沉着',
  mobility: '机动',
  endurance: '耐力',
  discipline: '纪律',
  portrait: '头像路径',
  cardImage: '完整卡面',
  avatarCropX: '头像裁切 X',
  avatarCropY: '头像裁切 Y',
  avatarCropWidth: '头像裁切宽度',
  avatarCropHeight: '头像裁切高度',
  shortName: '战队简称',
  nickname: '中文昵称',
  logo: 'Logo 路径',
  zone: '宏观区域',
  site: '包点',
  node_type: '节点类型',
  default_side: '默认优势方',
  floor: '楼层',
  x: '归一化 X 坐标',
  y: '归一化 Y 坐标',
  shape: '范围形状',
  radius: '圆形半径',
  area_usages: '节点范围用途',
  points: '多边形顶点',
  from: '起点',
  to: '终点',
  from_node: '起点',
  to_node: '终点',
  base_time: '基础移动时间',
  stamina_cost: '体能消耗',
  risk: '转移风险',
  noise: '暴露概率',
  risk_points: '风险热点',
  intercept_nodes: '拦截候选点',
  bidirectional: '是否双向',
  visible: '是否可见',
  range: '距离',
  angle_advantage: '角度优势',
  elevation: '高低关系',
  cover_modifier: '掩体修正',
  exposure_modifier: '暴露修正',
  side: '阵营',
  target_site: '目标包点',
  nodes: '节点序列',
  min_players: '最少人数',
  max_players: '最多人数',
  style_tags: '路线标签',
}

const tableLabels: Record<string, string> = {
  TbPlayer: '选手配置',
  TbTeam: '战队配置',
  TbTutorialBattle: '新手战斗配置',
}

export function configFieldLabel(key: string, comment?: string): string {
  const chineseComment = comment?.trim() && /[\u3400-\u9fff]/u.test(comment) ? comment.trim() : ''
  return `${key}(${chineseComment || fieldLabels[key] || '待补充中文说明'})`
}

export function documentFieldLabel(document: LubanTableDocument | undefined, key: string): string {
  return configFieldLabel(key, document?.fields.find((field) => field.key === key)?.comment)
}

export function configTableLabel(tableName: string, fallback?: string): string {
  return `${tableName}(${tableLabels[tableName] || fallback || '配置表'})`
}

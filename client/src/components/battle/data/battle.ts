// ---------------------------------------------------------------------------
// 比赛进行中（Battle）页面的数据模型与一份 mock 数据。
// 当前为静态 mock，后续可由 Nakama 的实时对局状态（match state）填充。
// ---------------------------------------------------------------------------

/** 阵营：ct = 反恐精英（蓝），t = 恐怖分子（金/黄） */
export type Side = "ct" | "t"

/** 手雷 / 投掷物类型 */
export type GrenadeType = "flash" | "smoke" | "he" | "molotov" | "decoy"

export interface BattlePlayer {
  /** 整场比赛内唯一且不透明的实例 ID。旧 mock 可省略。 */
  instanceId?: string
  /** 对应 TbPlayer.id，仅用于读取策划配置和视觉资料。 */
  configPlayerId?: string
  /** 选手 ID / 昵称，如 "s1mple" */
  id: string
  /** 头像 URL */
  avatar: string
  /** 旧头像回退路径 */
  portrait?: string
  /** 完整选手卡面路径 */
  cardImage?: string
  /** 归一化 5:7 头像裁切 */
  avatarCrop?: { x: number; y: number; width: number; height: number }
  /** 是否存活 */
  alive: boolean
  /** 血量 0-100 */
  health: number
  /** 护甲 0-100 */
  armor: number
  /** 是否有头甲（头盔） */
  helmet: boolean
  /** 当前回合金钱 */
  money: number
  /** 击杀 */
  kills: number
  /** 死亡 */
  deaths: number
  /** 助攻 */
  assists: number
  /** 主武器名称，如 "AK-47" / "AWP" */
  weapon: string
  /** 是否携带拆弹器（仅 CT 有意义） */
  defuseKit?: boolean
  /** 携带的手雷 */
  grenades: GrenadeType[]
}

export interface BattleTeam {
  /** 战队名称，如 "Natus Vincere" */
  name: string
  /** 队标缩写，如 "NAVI" / "G2"（暂无图片资源，用缩写徽标占位） */
  tag: string
  /** 当前阵营 */
  side: Side
  /** 当前比分（已赢回合数） */
  score: number
  /** 5 名选手 */
  players: BattlePlayer[]
}

export interface KillEvent {
  /** 唯一 id（用于 React key） */
  id: number
  /** 击杀者昵称 */
  killer: string
  /** 击杀者阵营 */
  killerSide: Side
  /** 被击杀者昵称 */
  victim: string
  /** 被击杀者阵营 */
  victimSide: Side
  /** 使用武器 */
  weapon: string
  /** 是否爆头 */
  headshot: boolean
}

export interface BattleState {
  /** SubPage 标题 */
  title: string
  /** 当前回合数 */
  round: number
  /** 总回合数（MR12 → 24 回合制） */
  maxRounds: number
  /** 左侧战队 */
  teamA: BattleTeam
  /** 右侧战队 */
  teamB: BattleTeam
  /** 实时击杀播报（最新在前） */
  killFeed: KillEvent[]
}

const AVATAR = "/images/star-player.png"

export const battleState: BattleState = {
  title: "比赛进行中",
  round: 4,
  maxRounds: 24,
  teamA: {
    name: "Natus Vincere",
    tag: "NAVI",
    side: "ct",
    score: 1,
    players: [
      { id: "s1mple",  avatar: AVATAR, alive: true,  health: 100, armor: 100, helmet: true,  money: 2300, kills: 20, deaths: 12, assists: 4, weapon: "AWP",    defuseKit: true,  grenades: ["flash", "smoke"] },
      { id: "b1t",     avatar: AVATAR, alive: true,  health: 87,  armor: 91,  helmet: true,  money: 1500, kills: 14, deaths: 15, assists: 6, weapon: "M4A4",   defuseKit: false, grenades: ["he", "flash"] },
      { id: "Aleksib", avatar: AVATAR, alive: true,  health: 100, armor: 100, helmet: true,  money: 800,  kills: 9,  deaths: 16, assists: 9, weapon: "M4A1-S", defuseKit: true,  grenades: ["smoke", "flash", "he"] },
      { id: "jL",      avatar: AVATAR, alive: true,  health: 45,  armor: 28,  helmet: false, money: 350,  kills: 12, deaths: 14, assists: 3, weapon: "M4A4",   defuseKit: false, grenades: ["molotov"] },
      { id: "iM",      avatar: AVATAR, alive: false, health: 0,   armor: 0,   helmet: false, money: 1200, kills: 7,  deaths: 17, assists: 5, weapon: "USP-S",  defuseKit: false, grenades: [] },
    ],
  },
  teamB: {
    name: "G2 Esports",
    tag: "G2",
    side: "t",
    score: 2,
    players: [
      { id: "m0NESY",  avatar: AVATAR, alive: true,  health: 100, armor: 100, helmet: true,  money: 2700, kills: 22, deaths: 11, assists: 3, weapon: "AWP",    grenades: ["smoke", "flash"] },
      { id: "NiKo",    avatar: AVATAR, alive: true,  health: 76,  armor: 84,  helmet: true,  money: 1900, kills: 19, deaths: 13, assists: 7, weapon: "AK-47",  grenades: ["he", "flash"] },
      { id: "huNter-", avatar: AVATAR, alive: true,  health: 100, armor: 100, helmet: true,  money: 1200, kills: 16, deaths: 14, assists: 8, weapon: "AK-47",  grenades: ["molotov", "smoke"] },
      { id: "HooXi",   avatar: AVATAR, alive: false, health: 0,   armor: 0,   helmet: false, money: 600,  kills: 6,  deaths: 18, assists: 11, weapon: "Galil",  grenades: [] },
      { id: "nexa",    avatar: AVATAR, alive: true,  health: 55,  armor: 47,  helmet: true,  money: 950,  kills: 11, deaths: 15, assists: 6, weapon: "AK-47",  grenades: ["flash"] },
    ],
  },
  killFeed: [
    { id: 6, killer: "m0NESY",  killerSide: "t",  victim: "b1t",   victimSide: "ct", weapon: "AWP",    headshot: false },
    { id: 5, killer: "s1mple",  killerSide: "ct", victim: "nexa",  victimSide: "t",  weapon: "AWP",    headshot: true },
    { id: 4, killer: "NiKo",    killerSide: "t",  victim: "jL",    victimSide: "ct", weapon: "AK-47",  headshot: true },
    { id: 3, killer: "huNter-", killerSide: "t",  victim: "iM",    victimSide: "ct", weapon: "AK-47",  headshot: false },
    { id: 2, killer: "b1t",     killerSide: "ct", victim: "HooXi", victimSide: "t",  weapon: "M4A4",   headshot: false },
    { id: 1, killer: "Aleksib", killerSide: "ct", victim: "NiKo",  victimSide: "t",  weapon: "M4A1-S", headshot: true },
  ],
}

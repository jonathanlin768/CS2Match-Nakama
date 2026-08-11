import { rowId, type ConfigValidationIssue, type LubanField, type LubanRow, type LubanTableDocument } from './luban'

const playerAbilityFields = [
  'entry', 'aim', 'trade', 'clutch', 'firepower', 'gamesense', 'reaction', 'positioning',
  'awareness', 'teamplay', 'utility', 'composure', 'mobility', 'endurance', 'discipline',
]

const tutorialTierFields = ['tier5PlayerIds', 'tier4PlayerIds', 'tier3PlayerIds', 'tier2PlayerIds', 'tier1PlayerIds'] as const

export function validateDocuments(documents: Iterable<LubanTableDocument>): ConfigValidationIssue[] {
  const docs = [...documents]
  const indexes = new Map<string, Set<string>>()
  for (const document of docs) indexes.set(document.tableName, new Set(document.rows.map(rowId).filter(Boolean)))

  const issues = docs.flatMap((document) => validateGeneric(document, indexes))
  const player = docs.find((document) => document.fileName === '#Player.xlsx')
  const team = docs.find((document) => document.fileName === '#Team.xlsx')
  const tutorial = docs.find((document) => document.fileName === '#TutorialBattle.xlsx')

  if (player) issues.push(...validatePlayer(player))
  if (tutorial) issues.push(...validateTutorial(tutorial, player, team))
  return dedupeIssues(issues)
}

export function issuesForFile(issues: ConfigValidationIssue[], fileName: string): ConfigValidationIssue[] {
  return issues.filter((issue) => issue.fileName === fileName)
}

function validateGeneric(document: LubanTableDocument, indexes: Map<string, Set<string>>): ConfigValidationIssue[] {
  const issues: ConfigValidationIssue[] = []
  const ids = new Set<string>()
  document.rows.forEach((row, rowIndex) => {
    const id = rowId(row)
    if (!id) issues.push(issue(document, rowIndex, row, 'id', 'ERROR', 'ID 不能为空'))
    else if (ids.has(id)) issues.push(issue(document, rowIndex, row, 'id', 'ERROR', `ID ${id} 重复`))
    else ids.add(id)

    for (const field of document.fields) {
      const value = row[field.key]
      validateType(document, rowIndex, row, field, value, issues)
      if (!field.refTable) continue
      const target = indexes.get(field.refTable)
      const values = field.list ? listValue(value) : scalarValue(value) ? [scalarValue(value)] : []
      for (const referencedId of values) {
        if (!target?.has(referencedId)) {
          issues.push(issue(document, rowIndex, row, field.key, 'ERROR', `${field.key} 引用不存在的 ${field.refTable}.${referencedId}`))
        }
      }
    }
  })
  return issues
}

function validateType(document: LubanTableDocument, rowIndex: number, row: LubanRow, field: LubanField, value: unknown, issues: ConfigValidationIssue[]): void {
  if (value === null || value === undefined || value === '') return
  if (field.list && !Array.isArray(value)) {
    issues.push(issue(document, rowIndex, row, field.key, 'ERROR', `${field.key} 必须是可序列化列表`))
    return
  }
  if (field.kind === 'int' && (!Number.isInteger(value) || typeof value !== 'number')) {
    issues.push(issue(document, rowIndex, row, field.key, 'ERROR', `${field.key} 必须是整数`))
  }
  if (field.kind === 'float' && (typeof value !== 'number' || !Number.isFinite(value))) {
    issues.push(issue(document, rowIndex, row, field.key, 'ERROR', `${field.key} 必须是数字`))
  }
  if (field.kind === 'bool' && typeof value !== 'boolean') {
    issues.push(issue(document, rowIndex, row, field.key, 'ERROR', `${field.key} 必须是布尔值`))
  }
}

function validatePlayer(document: LubanTableDocument): ConfigValidationIssue[] {
  const issues: ConfigValidationIssue[] = []
  document.rows.forEach((row, rowIndex) => {
    for (const field of playerAbilityFields) {
      const value = row[field]
      if (typeof value !== 'number' || !Number.isInteger(value) || value < 0 || value > 100) {
        issues.push(issue(document, rowIndex, row, field, 'ERROR', `${field} 必须是 0..100 的整数`))
      }
    }
    const cardImage = scalarValue(row.cardImage)
    const crop = ['avatarCropX', 'avatarCropY', 'avatarCropWidth', 'avatarCropHeight'].map((field) => Number(row[field]))
    if (!cardImage) {
      if (crop.some((value) => Number.isFinite(value) && value !== 0)) issues.push(issue(document, rowIndex, row, 'cardImage', 'WARNING', '未配置完整卡面，裁切参数不会生效'))
      return
    }
    const [x, y, width, height] = crop
    if (crop.some((value) => !Number.isFinite(value))) {
      issues.push(issue(document, rowIndex, row, 'avatarCropX', 'ERROR', '头像裁切参数必须是有限数字'))
      return
    }
    if (x < 0 || y < 0 || width <= 0 || height <= 0 || x + width > 1.000001 || y + height > 1.000001) {
      issues.push(issue(document, rowIndex, row, 'avatarCropX', 'ERROR', '头像裁切矩形必须位于 0..1 原图范围内且宽高为正数'))
      return
    }
    const pixelAspect = (width * 2) / (height * 3)
    if (Math.abs(pixelAspect - 5 / 7) > 0.005) {
      issues.push(issue(document, rowIndex, row, 'avatarCropWidth', 'ERROR', '完整卡面按 2:3 计算后，头像裁切必须保持 5:7 比例'))
    }
  })
  return issues
}

function validateTutorial(document: LubanTableDocument, player?: LubanTableDocument, team?: LubanTableDocument): ConfigValidationIssue[] {
  const issues: ConfigValidationIssue[] = []
  const players = new Map((player?.rows ?? []).map((row) => [rowId(row), row]))
  const teams = new Set((team?.rows ?? []).map(rowId))
  const enabledRows = document.rows.filter((row) => row.enabled === true)
  if (enabledRows.length > 1) {
    for (const row of enabledRows) {
      issues.push(issue(document, document.rows.indexOf(row), row, 'enabled', 'ERROR', '最多只能启用一条新手战斗配置'))
    }
  }

  document.rows.forEach((row, rowIndex) => {
    for (const field of tutorialTierFields) {
      if (listValue(row[field]).length === 0) issues.push(issue(document, rowIndex, row, field, 'ERROR', `${field} 至少需要一个候选选手`))
    }
    if (row.enabled !== true) return

    const budget = numberValue(row.budget)
    const rosterSize = numberValue(row.rosterSize)
    if (budget <= 0) issues.push(issue(document, rowIndex, row, 'budget', 'ERROR', '启用配置的 budget 必须大于 0'))
    if (rosterSize !== 5) issues.push(issue(document, rowIndex, row, 'rosterSize', 'ERROR', '启用配置的 rosterSize 必须等于 5'))

    const seen = new Set<string>()
    const prices: number[] = []
    tutorialTierFields.forEach((field, index) => {
      const price = 5 - index
      for (const playerId of listValue(row[field])) {
        if (seen.has(playerId)) issues.push(issue(document, rowIndex, row, field, 'ERROR', `选手 ${playerId} 在多个费用档重复`))
        seen.add(playerId)
        prices.push(price)
      }
    })
    prices.sort((a, b) => a - b)
    if (prices.length < rosterSize) issues.push(issue(document, rowIndex, row, 'rosterSize', 'ERROR', '候选池无法组成完整阵容'))
    else if (prices.slice(0, rosterSize).reduce((sum, price) => sum + price, 0) > budget) {
      issues.push(issue(document, rowIndex, row, 'budget', 'ERROR', '候选池无法在预算内组成完整阵容'))
    }

    const opponentTeamId = scalarValue(row.opponentTeamId)
    if (!opponentTeamId || !teams.has(opponentTeamId)) issues.push(issue(document, rowIndex, row, 'opponentTeamId', 'ERROR', '对手战队不存在'))
    const opponents = listValue(row.opponentPlayerIds)
    if (opponents.length !== rosterSize) issues.push(issue(document, rowIndex, row, 'opponentPlayerIds', 'ERROR', `对手阵容必须恰好包含 ${rosterSize} 名选手`))
    if (new Set(opponents).size !== opponents.length) issues.push(issue(document, rowIndex, row, 'opponentPlayerIds', 'ERROR', '对手阵容包含重复选手'))
    for (const opponentId of opponents) {
      const opponent = players.get(opponentId)
      if (!opponent || scalarValue(opponent.teamId) !== opponentTeamId) {
        issues.push(issue(document, rowIndex, row, 'opponentPlayerIds', 'ERROR', `对手选手 ${opponentId} 不属于 ${opponentTeamId || '所选战队'}`))
      }
    }
  })
  return issues
}

function issue(document: LubanTableDocument, rowIndex: number, row: LubanRow, field: string, severity: 'ERROR' | 'WARNING', message: string): ConfigValidationIssue {
  return { severity, fileName: document.fileName, rowIndex, rowId: rowId(row), field, message }
}

function scalarValue(value: unknown): string {
  return value === null || value === undefined || Array.isArray(value) ? '' : String(value)
}

function listValue(value: unknown): string[] {
  if (Array.isArray(value)) return value.map(String).filter(Boolean)
  if (value === null || value === undefined || value === '') return []
  return String(value).split(',').map((item) => item.trim()).filter(Boolean)
}

function numberValue(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function dedupeIssues(issues: ConfigValidationIssue[]): ConfigValidationIssue[] {
  const seen = new Set<string>()
  return issues.filter((item) => {
    const key = `${item.fileName}:${item.rowIndex}:${item.field}:${item.message}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

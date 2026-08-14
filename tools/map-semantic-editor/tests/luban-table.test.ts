import { describe, expect, it } from 'vitest'
import { parseCellValue, parseLubanFieldType, serializeCellValue, type LubanField, type LubanTableDocument } from '../src/lib/luban'
import { validateDocuments } from '../src/lib/lubanValidation'

describe('Luban 通用表模型', () => {
  it('解析基础、列表和引用类型', () => {
    expect(parseLubanFieldType('string')).toEqual({ kind: 'string', list: false })
    expect(parseLubanFieldType('int')).toEqual({ kind: 'int', list: false })
    expect(parseLubanFieldType('(list#sep=,),string')).toEqual({ kind: 'list', list: true })
    expect(parseLubanFieldType('string#ref=TbTeam')).toEqual({ kind: 'ref', list: false, refTable: 'TbTeam' })
    expect(parseLubanFieldType('(list#sep=,),string#ref=TbPlayer')).toEqual({ kind: 'ref', list: true, refTable: 'TbPlayer' })
    expect(parseLubanFieldType('vector3')).toEqual({ kind: 'unknown', list: false })
  })

  it('在列表单元格和内存数组之间往返', () => {
    const field: LubanField = { key: 'positions', type: '(list#sep=,),string', group: 'c,s', comment: '', kind: 'list', list: true, column: 2 }
    expect(parseCellValue('IGL,AWPer', field)).toEqual(['IGL', 'AWPer'])
    expect(serializeCellValue(['IGL', 'AWPer'], field)).toBe('IGL,AWPer')
  })

  it('严格校验 TutorialBattle 的唯一启用、预算和对手阵容', () => {
    const team = document('#Team.xlsx', 'TbTeam', [{ id: 'team_a', name: 'A' }])
    const player = document('#Player.xlsx', 'TbPlayer', [
      playerRow('p1', 'team_a'), playerRow('p2', 'team_a'), playerRow('p3', 'team_a'), playerRow('p4', 'team_a'), playerRow('p5', 'team_a'),
    ])
    const tutorial = document('#TutorialBattle.xlsx', 'TbTutorialBattle', [{
      id: 'tutorial', enabled: true, budget: 3, rosterSize: 5,
      tier5PlayerIds: ['p1'], tier4PlayerIds: ['p2'], tier3PlayerIds: ['p3'], tier2PlayerIds: ['p4'], tier1PlayerIds: ['p5'],
      opponentTeamId: 'team_a', opponentPlayerIds: ['p1'],
    }])
    const messages = validateDocuments([team, player, tutorial]).map((issue) => issue.message)
    expect(messages).toContain('候选池无法在预算内组成完整阵容')
    expect(messages).toContain('对手阵容必须恰好包含 5 名选手')
    expect(messages).not.toContain('对手选手 p1 同时存在于费用档候选池')
	})

  it('允许 TutorialBattle 候选池与对手阵容共享选手', () => {
    const team = document('#Team.xlsx', 'TbTeam', [{ id: 'team_a', name: 'A' }])
    const player = document('#Player.xlsx', 'TbPlayer', [
      playerRow('p1', 'team_a'), playerRow('p2', 'team_a'), playerRow('p3', 'team_a'), playerRow('p4', 'team_a'), playerRow('p5', 'team_a'),
    ])
    const tutorial = document('#TutorialBattle.xlsx', 'TbTutorialBattle', [{
      id: 'tutorial', enabled: true, budget: 15, rosterSize: 5,
      tier5PlayerIds: ['p1'], tier4PlayerIds: ['p2'], tier3PlayerIds: ['p3'], tier2PlayerIds: ['p4'], tier1PlayerIds: ['p5'],
      opponentTeamId: 'team_a', opponentPlayerIds: ['p1', 'p2', 'p3', 'p4', 'p5'],
    }])
    const messages = validateDocuments([team, player, tutorial]).map((issue) => issue.message)
    expect(messages.some((message) => message.includes('同时存在于费用档候选池'))).toBe(false)
  })

  it('校验 Player 的归一化 5:7 头像裁切', () => {
    const valid = playerRow('valid', 'team_a')
    Object.assign(valid, { cardImage: 'player-cards/valid.png', avatarCropX: 0.2, avatarCropY: 0.08, avatarCropWidth: 0.6, avatarCropHeight: 0.56 })
    const invalid = playerRow('invalid', 'team_a')
    Object.assign(invalid, { cardImage: 'player-cards/invalid.png', avatarCropX: 0.8, avatarCropY: 0, avatarCropWidth: 0.6, avatarCropHeight: 0.2 })
    const messages = validateDocuments([document('#Player.xlsx', 'TbPlayer', [valid, invalid])])
    expect(messages.some((issue) => issue.rowId === 'valid' && issue.field.startsWith('avatarCrop'))).toBe(false)
    expect(messages).toEqual(expect.arrayContaining([expect.objectContaining({ rowId: 'invalid', field: 'avatarCropX', severity: 'ERROR' })]))
  })
})

function document(fileName: string, tableName: string, rows: Record<string, import('../src/lib/luban').LubanCellValue>[]): LubanTableDocument {
  const keys = [...new Set(rows.flatMap((row) => Object.keys(row)))]
  return {
    fileName,
    tableName,
    owner: fileName.includes('Player') ? 'player' : fileName.includes('Team') ? 'team' : 'tutorial',
    editable: true,
    fields: keys.map((key, index) => {
      const value = rows[0]?.[key]
      const kind = key.startsWith('avatarCrop') ? 'float' : Array.isArray(value) ? 'list' : typeof value === 'boolean' ? 'bool' : typeof value === 'number' ? 'int' : 'string'
      const type = kind === 'list' ? '(list#sep=,),string' : kind
      return { key, type, group: 'c,s', comment: '', kind, list: Array.isArray(value), column: index + 2 }
    }),
    rows,
    originalMeta: { sheetName: 'Sheet1', version: { mtimeMs: 0, size: 0, hash: '' } },
    warnings: [],
    dirty: false,
  }
}

function playerRow(id: string, teamId: string): Record<string, import('../src/lib/luban').LubanCellValue> {
  return { id, teamId, entry: 50, aim: 50, trade: 50, clutch: 50, firepower: 50, gamesense: 50, reaction: 50, positioning: 50, awareness: 50, teamplay: 50, utility: 50, composure: 50, mobility: 50, endurance: 50, discipline: 50 }
}

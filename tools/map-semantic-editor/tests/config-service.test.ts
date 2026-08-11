import fs from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import ExcelJS from 'exceljs'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { readLubanWorkbook } from '../server/importTables'
import { findReferences, readLubanTable, saveConfigImage, saveLubanTable, saveLubanTables, syncRecordId } from '../server/configTables'

let root = ''

beforeEach(async () => {
  root = await fs.mkdtemp(path.join(os.tmpdir(), 'cs2-config-editor-'))
})

afterEach(async () => {
  await fs.rm(root, { recursive: true, force: true })
})

describe('通用 Luban workbook 服务', () => {
  it('读取四行表头、列表和引用元数据', async () => {
    await writeWorkbook('#Player.xlsx', [
      ['id', 'string'], ['teamId', 'string#ref=TbTeam'], ['positions', '(list#sep=,),string'],
    ], [['p1', 'team_a', 'IGL,AWPer']])
    const document = await readLubanTable('#Player.xlsx', root)
    expect(document.fields.find((field) => field.key === 'teamId')).toMatchObject({ kind: 'ref', refTable: 'TbTeam', list: false })
    expect(document.rows[0].positions).toEqual(['IGL', 'AWPer'])
  })

  it('只重写数据区并保留表头、列宽和备份', async () => {
    const target = await writeWorkbook('#item.xlsx', [['id', 'int'], ['name', 'string']], [[1, '旧名称']])
    const before = await readLubanTable('#item.xlsx', root)
    before.rows[0].name = '新名称'
    const result = await saveLubanTable(before, root)
    const workbook = await readLubanWorkbook(target)
    const sheet = workbook.worksheets[0]
    expect(sheet.getRow(1).getCell(1).value).toBe('##var')
    expect(sheet.getRow(5).getCell(3).value).toBe('新名称')
    expect(sheet.getColumn(3).width).toBe(31)
    await expect(fs.access(path.join(root, result.backup))).resolves.toBeUndefined()
  })

  it('阻止覆盖外部修改且不触碰原内容', async () => {
    const target = await writeWorkbook('#item.xlsx', [['id', 'int'], ['name', 'string']], [[1, '原值']])
    const document = await readLubanTable('#item.xlsx', root)
    await fs.appendFile(target, Buffer.from([0]))
    const changed = await fs.readFile(target)
    await expect(saveLubanTable(document, root)).rejects.toThrow('已被外部修改')
    expect(await fs.readFile(target)).toEqual(changed)
  })

  it('损坏 workbook 直接报错且不自动重建', async () => {
    const target = path.join(root, '#item.xlsx')
    await fs.writeFile(target, 'broken workbook')
    await expect(readLubanTable('#item.xlsx', root)).rejects.toThrow('无法读取')
    expect(await fs.readFile(target, 'utf8')).toBe('broken workbook')
  })

  it('批量保存准备失败时不写入其他表', async () => {
    const firstTarget = await writeWorkbook('#Team.xlsx', [['id', 'string'], ['name', 'string']], [['team_a', 'A']])
    const secondTarget = await writeWorkbook('#item.xlsx', [['id', 'int'], ['name', 'string']], [[1, 'Item']])
    const first = await readLubanTable('#Team.xlsx', root)
    const second = await readLubanTable('#item.xlsx', root)
    first.rows[0].name = 'Changed'
    await fs.appendFile(secondTarget, Buffer.from([0]))
    await expect(saveLubanTables([first, second], root)).rejects.toThrow('已被外部修改')
    const reloaded = await readLubanTable('#Team.xlsx', root)
    expect(reloaded.rows[0].name).toBe('A')
    await expect(fs.access(firstTarget)).resolves.toBeUndefined()
  })

  it('显式同步 Team ID 及单选引用', async () => {
    await writeWorkbook('#Team.xlsx', [['id', 'string'], ['name', 'string']], [['team_a', 'A']])
    await writeWorkbook('#Player.xlsx', playerFields(), [playerValues('p1', 'team_a')])
    const references = await findReferences('TbTeam', 'team_a', root)
    expect(references).toHaveLength(1)
    await syncRecordId('#Team.xlsx', 'team_a', 'team_new', root)
    expect((await readLubanTable('#Team.xlsx', root)).rows[0].id).toBe('team_new')
    expect((await readLubanTable('#Player.xlsx', root)).rows[0].teamId).toBe('team_new')
  })

  it('图片只能写入受控目录并校验格式与同名冲突', async () => {
    const portraits = path.join(root, 'portraits')
    const teams = path.join(root, 'teams')
    const png = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB', 'base64').toString('base64')
    const result = await saveConfigImage('team', 'logo.png', png, false, { portrait: portraits, team: teams })
    expect(result.path).toBe('teams/logo.png')
    await expect(fs.access(path.join(teams, 'logo.png'))).resolves.toBeUndefined()
    await expect(saveConfigImage('team', 'logo.png', png, false, { portrait: portraits, team: teams })).rejects.toThrow('已存在')
    await expect(saveConfigImage('team', '../logo.png', png, false, { portrait: portraits, team: teams })).rejects.toThrow('不合法')
  })

  it('完整卡面只能写入 player-cards 目录并支持显式覆盖', async () => {
    const portraits = path.join(root, 'portraits')
    const teams = path.join(root, 'teams')
    const playerCard = path.join(root, 'player-cards')
    const png = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB', 'base64').toString('base64')
    const roots = { portrait: portraits, team: teams, playerCard }
    const result = await saveConfigImage('player-card', 'player-one.png', png, false, roots)
    expect(result.path).toBe('player-cards/player-one.png')
    await expect(fs.access(path.join(playerCard, 'player-one.png'))).resolves.toBeUndefined()
    await expect(saveConfigImage('player-card', 'player-one.png', png, false, roots)).rejects.toThrow('已存在')
    await expect(saveConfigImage('player-card', 'player-one.png', png, true, roots)).resolves.toMatchObject({ path: 'player-cards/player-one.png' })
    await expect(saveConfigImage('player-card', 'player-one.gif', png, false, roots)).rejects.toThrow('只支持')
  })
})

async function writeWorkbook(fileName: string, fields: [string, string][], rows: unknown[][]): Promise<string> {
  const workbook = new ExcelJS.Workbook()
  const sheet = workbook.addWorksheet('Sheet1')
  sheet.addRow(['##var', ...fields.map(([key]) => key)])
  sheet.addRow(['##type', ...fields.map(([, type]) => type)])
  sheet.addRow(['##group', ...fields.map(() => 'c,s,e')])
  sheet.addRow(['##', ...fields.map(([key]) => key)])
  rows.forEach((row) => sheet.addRow([null, ...row]))
  sheet.getColumn(3).width = 31
  const target = path.join(root, fileName)
  await workbook.xlsx.writeFile(target)
  return target
}

function playerFields(): [string, string][] {
  return [['id', 'string'], ['teamId', 'string#ref=TbTeam'], ...['entry', 'aim', 'trade', 'clutch', 'firepower', 'gamesense', 'reaction', 'positioning', 'awareness', 'teamplay', 'utility', 'composure', 'mobility', 'endurance', 'discipline'].map((key) => [key, 'int'] as [string, string])]
}

function playerValues(id: string, teamId: string): unknown[] {
  return [id, teamId, ...Array.from({ length: 15 }, () => 50)]
}

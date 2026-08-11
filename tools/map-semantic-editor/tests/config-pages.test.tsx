import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AllTablesPage } from '../src/pages/AllTablesPage'
import { ConfigEditorShell } from '../src/components/ConfigEditorShell'
import { PlayerConfigPage } from '../src/pages/PlayerConfigPage'
import { TeamConfigPage } from '../src/pages/TeamConfigPage'
import { TutorialConfigPage } from '../src/pages/TutorialConfigPage'
import type { LubanField, LubanRow, LubanTableDocument, LubanTableOwner, LubanTableSummary } from '../src/lib/luban'
import { useConfigStore } from '../src/store/configStore'

const version = { mtimeMs: 1, size: 1, hash: 'fixture' }

function field(key: string, type = 'string', column = 1): LubanField {
  const list = type.startsWith('list')
  const refTable = type.match(/#ref=([A-Za-z0-9_]+)/)?.[1]
  return {
    key,
    type,
    group: '*',
    comment: key,
    kind: refTable ? 'ref' : list ? 'list' : type === 'int' ? 'int' : type === 'bool' ? 'bool' : 'string',
    refTable,
    list,
    column,
  }
}

function document(fileName: string, tableName: string, owner: LubanTableOwner, fields: LubanField[], rows: LubanRow[]): LubanTableDocument {
  return {
    fileName,
    tableName,
    owner,
    editable: true,
    fields,
    rows,
    originalMeta: { sheetName: 'Sheet1', version },
    warnings: [],
    dirty: false,
  }
}

const teamFields = [field('id'), field('name', 'string', 2), field('shortName', 'string', 3), field('nickname', 'string', 4), field('logo', 'string', 5)]
const playerFields = [
  field('id'), field('name', 'string', 2), field('teamId', 'string#ref=TbTeam', 3), field('nationality', 'string', 4),
  field('rarity', 'string', 5), field('positions', 'list,string', 6), field('entry', 'int', 7), field('portrait', 'string', 8),
  field('cardImage', 'string', 9), field('avatarCropX', 'float', 10), field('avatarCropY', 'float', 11), field('avatarCropWidth', 'float', 12), field('avatarCropHeight', 'float', 13),
]
const tutorialFields = [
  field('id'), field('enabled', 'bool', 2), field('version', 'int', 3), field('mapId', 'string', 4), field('budget', 'int', 5),
  field('rosterSize', 'int', 6), field('tier5PlayerIds', 'list,string#ref=TbPlayer', 7), field('tier4PlayerIds', 'list,string#ref=TbPlayer', 8),
  field('tier3PlayerIds', 'list,string#ref=TbPlayer', 9), field('tier2PlayerIds', 'list,string#ref=TbPlayer', 10), field('tier1PlayerIds', 'list,string#ref=TbPlayer', 11),
  field('opponentTeamId', 'string#ref=TbTeam', 12), field('opponentPlayerIds', 'list,string#ref=TbPlayer', 13),
]

function seed(extra: LubanTableDocument[] = []) {
  const teams = document('#Team.xlsx', 'TbTeam', 'team', teamFields, [
    { id: 'TEAM_A', name: 'Alpha', shortName: 'A', nickname: 'Alpha', logo: 'teams/a.png' },
    { id: 'TEAM_B', name: 'Bravo', shortName: 'B', nickname: 'Bravo', logo: 'teams/b.png' },
  ])
  const players = document('#Player.xlsx', 'TbPlayer', 'player', playerFields, [
    { id: 'P1', name: 'Player One', teamId: 'TEAM_A', nationality: 'CN', rarity: 'A', positions: ['rifler'], entry: 70, portrait: 'portraits/p1.png', cardImage: 'player-cards/p1.png', avatarCropX: 0.2, avatarCropY: 0.08, avatarCropWidth: 0.6, avatarCropHeight: 0.56 },
  ])
  const tutorials = document('#TutorialBattle.xlsx', 'TbTutorialBattle', 'tutorial', tutorialFields, [
    { id: 'T1', enabled: true, version: 1, mapId: 'de_dust2', budget: 15, rosterSize: 5, tier5PlayerIds: ['P1'], tier4PlayerIds: ['P1'], tier3PlayerIds: ['P1'], tier2PlayerIds: ['P1'], tier1PlayerIds: ['P1'], opponentTeamId: 'TEAM_A', opponentPlayerIds: ['P1'] },
    { id: 'T2', enabled: false, version: 2, mapId: 'de_dust2', budget: 15, rosterSize: 5, tier5PlayerIds: ['P1'], tier4PlayerIds: ['P1'], tier3PlayerIds: ['P1'], tier2PlayerIds: ['P1'], tier1PlayerIds: ['P1'], opponentTeamId: 'TEAM_A', opponentPlayerIds: ['P1'] },
  ])
  const docs = [teams, players, tutorials, ...extra]
  const summaries: LubanTableSummary[] = docs.map((item) => ({
    fileName: item.fileName,
    tableName: item.tableName,
    owner: item.owner,
    editable: item.editable,
    rows: item.rows.length,
    columns: item.fields.length,
    status: 'ready',
    warnings: [],
  }))
  useConfigStore.setState({
    summaries,
    documents: Object.fromEntries(docs.map((item) => [item.fileName, item])),
    activeFile: docs[0].fileName,
    issues: [],
    logs: [],
    genConfig: null,
    serviceMessage: 'ready',
    busy: false,
    initialized: true,
    referencesFor: vi.fn(async () => []),
  })
}

beforeEach(() => {
  seed()
  vi.spyOn(window, 'confirm').mockReturnValue(true)
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  window.location.hash = ''
  useConfigStore.setState(useConfigStore.getInitialState(), true)
})

describe('config editor pages', () => {
  it('enables current and batch save commands only after a table becomes dirty', () => {
    const item = document('#item.xlsx', 'Tbitem', 'generic', [field('id'), field('name', 'string', 2)], [{ id: 'AK47', name: 'AK-47' }])
    seed([item])
    useConfigStore.setState({ activeFile: '#item.xlsx' })
    window.location.hash = '#tables'
    render(<ConfigEditorShell />)

    expect(screen.getByRole('button', { name: '保存当前表' })).toBeDisabled()
    expect(screen.getByRole('button', { name: '保存全部' })).toBeDisabled()
    const table = screen.getByRole('table')
    fireEvent.change(within(table).getAllByRole('textbox')[1], { target: { value: 'Changed' } })
    expect(screen.getByRole('button', { name: '保存当前表' })).toBeEnabled()
    expect(screen.getByRole('button', { name: '保存全部 (1)' })).toBeEnabled()
  })

  it('edits, adds and deletes rows in a generic table', async () => {
    const item = document('#item.xlsx', 'Tbitem', 'generic', [field('id'), field('name', 'string', 2)], [{ id: 'AK47', name: 'AK-47' }])
    seed([item])
    useConfigStore.setState({ activeFile: '#item.xlsx' })
    render(<AllTablesPage onOpenMap={vi.fn()} />)

    const table = screen.getByRole('table')
    const nameInput = within(table).getAllByRole('textbox')[1]
    fireEvent.change(nameInput, { target: { value: 'AK-47 Updated' } })
    expect(useConfigStore.getState().documents['#item.xlsx'].rows[0].name).toBe('AK-47 Updated')

    fireEvent.click(screen.getByRole('button', { name: /新增行/ }))
    expect(useConfigStore.getState().documents['#item.xlsx'].rows).toHaveLength(2)
    fireEvent.click(screen.getByRole('button', { name: /删除行/ }))
    await waitFor(() => expect(useConfigStore.getState().documents['#item.xlsx'].rows).toHaveLength(1))
  })

  it('routes map-owned workbooks back to the map editor', () => {
    const map = document('#MapNode.xlsx', 'TbMapNode', 'map', [field('id')], [{ id: 'A_SITE' }])
    seed([map])
    useConfigStore.setState({ activeFile: '#MapNode.xlsx' })
    const onOpenMap = vi.fn()
    render(<AllTablesPage onOpenMap={onOpenMap} />)

    fireEvent.click(screen.getByRole('button', { name: /前往地图配置/ }))
    expect(onOpenMap).toHaveBeenCalledOnce()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('edits player references, abilities and position tags', () => {
    render(<PlayerConfigPage />)
    const teamSelect = screen.getByLabelText('teamId')
    fireEvent.change(teamSelect, { target: { value: 'TEAM_B' } })
    const entry = screen.getByLabelText('entry')
    fireEvent.change(entry, { target: { value: '101' } })
    const positionInput = screen.getByPlaceholderText(/输入标签后按 Enter/)
    fireEvent.change(positionInput, { target: { value: 'awper' } })
    fireEvent.keyDown(positionInput, { key: 'Enter' })

    const row = useConfigStore.getState().documents['#Player.xlsx'].rows[0]
    expect(row.teamId).toBe('TEAM_B')
    expect(row.entry).toBe(101)
    expect(row.positions).toEqual(['rifler', 'awper'])
    expect(useConfigStore.getState().issues.some((issue) => issue.field === 'entry' && issue.severity === 'ERROR')).toBe(true)
  })

  it('copies a selected team image into the configured project path', async () => {
    const uploadImage = vi.fn(async () => 'teams/new-logo.png')
    useConfigStore.setState({ uploadImage })
    const { container } = render(<TeamConfigPage />)
    const input = container.querySelector('input[type="file"]') as HTMLInputElement
    const image = new File([new Uint8Array([137, 80, 78, 71])], 'new-logo.png', { type: 'image/png' })
    fireEvent.change(input, { target: { files: [image] } })

    await waitFor(() => expect(uploadImage).toHaveBeenCalledWith('team', image))
    expect(useConfigStore.getState().documents['#Team.xlsx'].rows[0].logo).toBe('teams/new-logo.png')
  })

  it('uploads a full player card and resets its crop for the new image', async () => {
    const uploadImage = vi.fn(async () => 'player-cards/new-card.png')
    useConfigStore.setState({ uploadImage })
    const { container } = render(<PlayerConfigPage />)
    const input = container.querySelector('.playerVisualEditor input[type="file"]') as HTMLInputElement
    const image = new File([new Uint8Array([137, 80, 78, 71])], 'new-card.png', { type: 'image/png' })
    fireEvent.change(input, { target: { files: [image] } })

    await waitFor(() => expect(uploadImage).toHaveBeenCalledWith('player-card', image))
    const row = useConfigStore.getState().documents['#Player.xlsx'].rows[0]
    expect(row).toMatchObject({ cardImage: 'player-cards/new-card.png', avatarCropX: 0, avatarCropY: 0, avatarCropWidth: 0, avatarCropHeight: 0 })
  })

  it('resets a loaded 2:3 card to a centered maximum 5:7 crop', () => {
    render(<PlayerConfigPage />)
    const cropImage = screen.getByAltText('头像裁切源图') as HTMLImageElement
    Object.defineProperty(cropImage, 'naturalWidth', { configurable: true, value: 1024 })
    Object.defineProperty(cropImage, 'naturalHeight', { configurable: true, value: 1536 })
    fireEvent.click(screen.getByRole('button', { name: /重置裁切/ }))

    const row = useConfigStore.getState().documents['#Player.xlsx'].rows[0]
    expect(row.avatarCropX).toBeCloseTo(0, 5)
    expect(row.avatarCropY).toBeCloseTo(1 / 30, 5)
    expect(row.avatarCropWidth).toBeCloseTo(1, 5)
    expect(row.avatarCropHeight).toBeCloseTo(14 / 15, 5)
  })

  it('keeps only one TutorialBattle plan enabled', () => {
    render(<TutorialConfigPage />)
    fireEvent.click(screen.getByRole('button', { name: /T2/ }))
    const enabled = screen.getByRole('checkbox', { name: /启用/ })
    fireEvent.click(enabled)

    const rows = useConfigStore.getState().documents['#TutorialBattle.xlsx'].rows
    expect(rows.map((row) => row.enabled)).toEqual([false, true])
    expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining('关闭当前已启用'))
  })

  it('reports an empty TutorialBattle tier and blocks saving', async () => {
    render(<TutorialConfigPage />)
    fireEvent.click(screen.getAllByRole('button', { name: /Player One ×/ })[0])
    const result = await useConfigStore.getState().saveCurrent('#TutorialBattle.xlsx')

    expect(result).toBe(false)
    expect(useConfigStore.getState().issues).toEqual(expect.arrayContaining([
      expect.objectContaining({ field: 'tier5PlayerIds', severity: 'ERROR' }),
    ]))
    expect(useConfigStore.getState().serviceMessage).toContain('阻止保存')
  })
})

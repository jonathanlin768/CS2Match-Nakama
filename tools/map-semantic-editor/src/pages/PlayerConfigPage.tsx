import { Copy, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { ConfigBottomPanel } from '../components/ConfigBottomPanel'
import { FormField, ImageField, RecordList, ReferenceSelect, TagEditor } from '../components/ConfigControls'
import { PlayerVisualEditor } from '../components/PlayerVisualEditor'
import { toListValue, toNumberValue, toStringValue } from '../lib/configValues'
import { rowId, type LubanRow } from '../lib/luban'
import { configTableLabel } from '../lib/configLabels'
import { useConfigStore } from '../store/configStore'

const fileName = '#Player.xlsx'
const abilities = ['entry', 'aim', 'trade', 'clutch', 'firepower', 'gamesense', 'reaction', 'positioning', 'awareness', 'teamplay', 'utility', 'composure', 'mobility', 'endurance', 'discipline']
const emptyRows: LubanRow[] = []

export function PlayerConfigPage() {
  const document = useConfigStore((state) => state.documents[fileName])
  const teamDocument = useConfigStore((state) => state.documents['#Team.xlsx'])
  const teams = teamDocument?.rows ?? emptyRows
  const busy = useConfigStore((state) => state.busy)
  const updateCell = useConfigStore((state) => state.updateCell)
  const updateCells = useConfigStore((state) => state.updateCells)
  const addRow = useConfigStore((state) => state.addRow)
  const duplicateRow = useConfigStore((state) => state.duplicateRow)
  const deleteRow = useConfigStore((state) => state.deleteRow)
  const changeId = useConfigStore((state) => state.changeId)
  const referencesFor = useConfigStore((state) => state.referencesFor)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [search, setSearch] = useState('')
  const [teamFilter, setTeamFilter] = useState('')
  const [rarityFilter, setRarityFilter] = useState('')
  const [positionFilter, setPositionFilter] = useState('')
  const [idEdit, setIdEdit] = useState({ key: '', value: '' })
  const [referenceState, setReferenceState] = useState({ key: '', count: 0 })
  const row = document?.rows[selectedIndex]
  const recordKey = row ? `${selectedIndex}:${rowId(row)}` : ''
  const idDraft = idEdit.key === recordKey ? idEdit.value : row ? rowId(row) : ''
  const referenceCount = referenceState.key === recordKey ? referenceState.count : 0
  const rarities = useMemo(() => [...new Set((document?.rows ?? []).map((candidate) => toStringValue(candidate.rarity)).filter(Boolean))].sort(), [document?.rows])
  const positions = useMemo(() => [...new Set((document?.rows ?? []).flatMap((candidate) => toListValue(candidate.positions)))].sort(), [document?.rows])
  const filteredRows = useMemo(() => (document?.rows ?? []).filter((candidate) => {
    const queryMatch = !search || `${toStringValue(candidate.name)} ${rowId(candidate)}`.toLowerCase().includes(search.toLowerCase())
    return queryMatch && (!teamFilter || candidate.teamId === teamFilter) && (!rarityFilter || candidate.rarity === rarityFilter) && (!positionFilter || toListValue(candidate.positions).includes(positionFilter))
  }), [document?.rows, positionFilter, rarityFilter, search, teamFilter])

  useEffect(() => {
    let active = true
    if (!row) return
    void referencesFor(fileName, selectedIndex).then((references) => { if (active) setReferenceState({ key: recordKey, count: references.length }) })
    return () => { active = false }
  }, [recordKey, referencesFor, row, selectedIndex])

  if (!document) return <div className="configWorkspacePage"><div className="emptyState large">正在读取 #Player.xlsx…</div><ConfigBottomPanel fileName={fileName} /></div>

  function selectFiltered(index: number) {
    const original = document!.rows.indexOf(filteredRows[index])
    if (original >= 0) setSelectedIndex(original)
  }

  function createRow() {
    addRow(fileName)
    setSelectedIndex(document!.rows.length)
  }

  async function removeRow() {
    if (!row || !window.confirm(`确定删除选手 ${rowId(row) || '未命名记录'}？`)) return
    if (await deleteRow(fileName, selectedIndex)) setSelectedIndex(Math.max(0, selectedIndex - 1))
  }

  return (
    <div className="configWorkspacePage">
      <section className="configWorkspaceMain recordEditorLayout playerLayout">
        <aside className="configSidebar">
          <div className="panelHeading"><div><h1>{configTableLabel(document.tableName, '选手配置')}</h1><p>{document.rows.length} 名选手 {document.dirty ? '· 未保存' : ''}</p></div></div>
          <input className="searchInput" value={search} placeholder="搜索选手或 ID" onChange={(event) => setSearch(event.target.value)} />
          <div className="filterGrid"><select value={teamFilter} onChange={(event) => setTeamFilter(event.target.value)}><option value="">全部战队</option>{teams.map((team) => <option key={rowId(team)} value={rowId(team)}>{toStringValue(team.shortName) || rowId(team)}</option>)}</select><select value={rarityFilter} onChange={(event) => setRarityFilter(event.target.value)}><option value="">全部稀有度</option>{rarities.map((rarity) => <option key={rarity}>{rarity}</option>)}</select><select value={positionFilter} onChange={(event) => setPositionFilter(event.target.value)}><option value="">全部位置</option>{positions.map((position) => <option key={position}>{position}</option>)}</select></div>
          <div className="rowCommands compact"><button type="button" className="iconButton" title="新增选手" onClick={createRow}><Plus size={16} /></button><button type="button" className="iconButton" title="复制选手" disabled={!row} onClick={() => duplicateRow(fileName, selectedIndex)}><Copy size={16} /></button><button type="button" className="iconButton danger" title="删除选手" disabled={!row || busy} onClick={() => void removeRow()}><Trash2 size={16} /></button></div>
          <RecordList rows={filteredRows} selectedIndex={filteredRows.indexOf(row!)} onSelect={selectFiltered} secondary={(candidate) => `${toStringValue(candidate.rarity)} · ${toListValue(candidate.positions).join('/')}`} />
        </aside>
        <section className="recordDetailPane">
          {!row ? <div className="emptyState large">新增或选择一名选手。</div> : <div className="detailForm">
            <header className="detailHeader"><div><h2>{toStringValue(row.name) || '未命名选手'}</h2><p>{rowId(row)} · {referenceCount} 处 TutorialBattle 引用</p></div></header>
            <section className="formSection"><h3>基础资料</h3><div className="formGrid threeColumns"><FormField label="id" hint={referenceCount > 0 ? '应用后同步费用档与对手阵容引用' : '当前没有外部引用'}><div className="inlineField"><input value={idDraft} onChange={(event) => setIdEdit({ key: recordKey, value: event.target.value })} /><button type="button" className="commandButton" disabled={!idDraft || idDraft === rowId(row)} onClick={() => void changeId(fileName, selectedIndex, idDraft)}><RefreshCw size={14} />应用 ID</button></div></FormField><FormField label="name"><input value={toStringValue(row.name)} onChange={(event) => updateCell(fileName, selectedIndex, 'name', event.target.value)} /></FormField><FormField label="teamId"><ReferenceSelect value={toStringValue(row.teamId)} rows={teams} onChange={(value) => updateCell(fileName, selectedIndex, 'teamId', value)} /></FormField><FormField label="nationality"><input value={toStringValue(row.nationality)} onChange={(event) => updateCell(fileName, selectedIndex, 'nationality', event.target.value)} /></FormField><FormField label="rarity"><select value={toStringValue(row.rarity)} onChange={(event) => updateCell(fileName, selectedIndex, 'rarity', event.target.value)}><option value="">请选择</option>{['C', 'B', 'A', 'S', 'SS'].map((rarity) => <option key={rarity}>{rarity}</option>)}</select></FormField><FormField label="positions"><TagEditor value={toListValue(row.positions)} onChange={(value) => updateCell(fileName, selectedIndex, 'positions', value)} /></FormField></div></section>
            <section className="formSection"><h3>能力值</h3><div className="abilityGrid">{abilities.map((ability) => <FormField key={ability} label={ability}><input type="number" min={0} max={100} step={1} value={toNumberValue(row[ability])} onChange={(event) => updateCell(fileName, selectedIndex, ability, Number(event.target.value))} /></FormField>)}</div></section>
            <section className="formSection" id="player-visual-editor"><h3>选手视觉资源</h3><PlayerVisualEditor key={recordKey} row={row} onChange={(values) => updateCells(fileName, selectedIndex, values)} /><details className="legacyPortrait"><summary>旧头像回退</summary><ImageField label="portrait" kind="portrait" value={toStringValue(row.portrait)} onChange={(value) => updateCell(fileName, selectedIndex, 'portrait', value)} /></details></section>
          </div>}
        </section>
      </section>
      <ConfigBottomPanel fileName={fileName} onIssueClick={(issue) => { setSelectedIndex(issue.rowIndex); if (issue.field.startsWith('avatarCrop') || issue.field === 'cardImage') requestAnimationFrame(() => window.document.getElementById('player-visual-editor')?.scrollIntoView({ block: 'start' })) }} />
    </div>
  )
}

import { Copy, Plus, Search, Trash2 } from 'lucide-react'
import { useMemo, useState } from 'react'
import { ConfigBottomPanel } from '../components/ConfigBottomPanel'
import { FormField, RecordList, ReferenceSelect } from '../components/ConfigControls'
import { toListValue, toNumberValue, toStringValue } from '../lib/configValues'
import { displayLabel, rowId, type LubanRow } from '../lib/luban'
import { configTableLabel, documentFieldLabel } from '../lib/configLabels'
import { useConfigStore } from '../store/configStore'

const fileName = '#TutorialBattle.xlsx'
const tiers = [
  ['tier5PlayerIds', 5],
  ['tier4PlayerIds', 4],
  ['tier3PlayerIds', 3],
  ['tier2PlayerIds', 2],
  ['tier1PlayerIds', 1],
] as const
const emptyRows: LubanRow[] = []

export function TutorialConfigPage() {
  const document = useConfigStore((state) => state.documents[fileName])
  const players = useConfigStore((state) => state.documents['#Player.xlsx']?.rows ?? emptyRows)
  const teams = useConfigStore((state) => state.documents['#Team.xlsx']?.rows ?? emptyRows)
  const updateCell = useConfigStore((state) => state.updateCell)
  const addRow = useConfigStore((state) => state.addRow)
  const duplicateRow = useConfigStore((state) => state.duplicateRow)
  const deleteRow = useConfigStore((state) => state.deleteRow)
  const setTutorialEnabled = useConfigStore((state) => state.setTutorialEnabled)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const row = document?.rows[selectedIndex]

  if (!document) return <div className="configWorkspacePage"><div className="emptyState large">正在读取 #TutorialBattle.xlsx…</div><ConfigBottomPanel fileName={fileName} /></div>

  function createRow() {
    addRow(fileName)
    setSelectedIndex(document!.rows.length)
  }

  async function removeRow() {
    if (!row || !window.confirm(`确定删除新手战斗配置 ${rowId(row) || '未命名记录'}？`)) return
    if (await deleteRow(fileName, selectedIndex)) setSelectedIndex(Math.max(0, selectedIndex - 1))
  }

  return (
    <div className="configWorkspacePage">
      <section className="configWorkspaceMain recordEditorLayout tutorialLayout">
        <aside className="configSidebar">
          <div className="panelHeading"><div><h1>{configTableLabel(document.tableName, '新手战斗配置')}</h1><p>{document.rows.length} 套方案 {document.dirty ? '· 未保存' : ''}</p></div></div>
          <div className="rowCommands compact"><button type="button" className="iconButton" title="新增方案" onClick={createRow}><Plus size={16} /></button><button type="button" className="iconButton" title="复制方案" disabled={!row} onClick={() => duplicateRow(fileName, selectedIndex)}><Copy size={16} /></button><button type="button" className="iconButton danger" title="删除方案" disabled={!row} onClick={() => void removeRow()}><Trash2 size={16} /></button></div>
          <RecordList rows={document.rows} selectedIndex={selectedIndex} onSelect={setSelectedIndex} secondary={(candidate) => candidate.enabled === true ? '已启用' : `版本 ${toNumberValue(candidate.version)}`} />
        </aside>
        <section className="recordDetailPane">
          {!row ? <div className="emptyState large">新增或选择一套配置。</div> : <div className="detailForm">
            <header className="detailHeader"><div><h2>{rowId(row) || '未命名方案'}</h2><p>{row.enabled === true ? '当前启用' : '未启用'} · 预算 {toNumberValue(row.budget)} · {toNumberValue(row.rosterSize)} 人阵容</p></div><label className="switchField"><span>{documentFieldLabel(document, 'enabled')}</span><input type="checkbox" checked={row.enabled === true} onChange={(event) => setTutorialEnabled(selectedIndex, event.target.checked)} /></label></header>
            <section className="formSection"><h3>基础规则</h3><div className="formGrid threeColumns"><FormField label="id"><input value={rowId(row)} onChange={(event) => updateCell(fileName, selectedIndex, 'id', event.target.value)} /></FormField><FormField label="version"><input type="number" min={1} step={1} value={toNumberValue(row.version)} onChange={(event) => updateCell(fileName, selectedIndex, 'version', Number(event.target.value))} /></FormField><FormField label="mapId"><input value={toStringValue(row.mapId)} onChange={(event) => updateCell(fileName, selectedIndex, 'mapId', event.target.value)} /></FormField><FormField label="budget"><input type="number" min={1} step={1} value={toNumberValue(row.budget)} onChange={(event) => updateCell(fileName, selectedIndex, 'budget', Number(event.target.value))} /></FormField><FormField label="rosterSize" hint="当前运行时要求固定为 5"><input type="number" min={1} step={1} value={toNumberValue(row.rosterSize)} onChange={(event) => updateCell(fileName, selectedIndex, 'rosterSize', Number(event.target.value))} /></FormField></div></section>
            <section className="formSection"><h3>费用档候选池</h3><div className="tierEditorGrid">{tiers.map(([field]) => <PlayerMultiPicker key={field} title={documentFieldLabel(document, field)} players={players} selected={toListValue(row[field])} disabledIds={new Set(tiers.filter(([otherField]) => otherField !== field).flatMap(([otherField]) => toListValue(row[otherField])))} onChange={(value) => updateCell(fileName, selectedIndex, field, value)} />)}</div></section>
            <section className="formSection"><h3>对手阵容</h3><div className="formGrid twoColumns"><FormField label="opponentTeamId"><ReferenceSelect value={toStringValue(row.opponentTeamId)} rows={teams} onChange={(value) => updateCell(fileName, selectedIndex, 'opponentTeamId', value)} /></FormField><div /></div><PlayerMultiPicker title={documentFieldLabel(document, 'opponentPlayerIds')} players={players} selected={toListValue(row.opponentPlayerIds)} preferredTeam={toStringValue(row.opponentTeamId)} onChange={(value) => updateCell(fileName, selectedIndex, 'opponentPlayerIds', value)} /></section>
          </div>}
        </section>
      </section>
      <ConfigBottomPanel fileName={fileName} />
    </div>
  )
}

function PlayerMultiPicker({ title, players, selected, disabledIds, preferredTeam, onChange }: { title: string; players: LubanRow[]; selected: string[]; disabledIds?: ReadonlySet<string>; preferredTeam?: string; onChange: (value: string[]) => void }) {
  const [search, setSearch] = useState('')
  const candidates = useMemo(() => [...players].sort((left, right) => {
    if (preferredTeam) {
      const leftPreferred = left.teamId === preferredTeam ? 0 : 1
      const rightPreferred = right.teamId === preferredTeam ? 0 : 1
      if (leftPreferred !== rightPreferred) return leftPreferred - rightPreferred
    }
    return displayLabel(left).localeCompare(displayLabel(right))
  }).filter((player) => `${displayLabel(player)} ${rowId(player)} ${toStringValue(player.teamId)} ${toStringValue(player.rarity)} ${toListValue(player.positions).join(' ')}`.toLowerCase().includes(search.toLowerCase())), [players, preferredTeam, search])

  function toggle(id: string) {
    if (disabledIds?.has(id) && !selected.includes(id)) return
    onChange(selected.includes(id) ? selected.filter((value) => value !== id) : [...selected, id])
  }

  return (
    <div className="playerPicker">
      <header><div><strong>{title}</strong><span>{selected.length} 人</span></div><label><Search size={14} /><input value={search} placeholder="搜索选手、战队、位置" onChange={(event) => setSearch(event.target.value)} /></label></header>
      <div className="selectedChips">{selected.map((id) => { const player = players.find((candidate) => rowId(candidate) === id); return <button type="button" key={id} onClick={() => toggle(id)}>{player ? displayLabel(player) : id} ×</button> })}</div>
      <div className="playerPickerOptions">{candidates.map((player) => { const id = rowId(player); const active = selected.includes(id); const disabled = !active && disabledIds?.has(id) === true; return <label key={id} className={active ? 'active' : disabled ? 'disabled' : ''} title={disabled ? '已在其他阵容或费用档选择' : undefined}><input type="checkbox" aria-label={`${title} ${displayLabel(player)}`} checked={active} disabled={disabled} onChange={() => toggle(id)} /><span><strong>{displayLabel(player)}</strong><small>{disabled ? '已在其他阵容或费用档选择' : `${toStringValue(player.teamId)} · ${toStringValue(player.rarity)} · ${toListValue(player.positions).join('/')}`}</small></span></label> })}</div>
    </div>
  )
}

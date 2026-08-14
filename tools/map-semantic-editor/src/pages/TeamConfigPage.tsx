import { Copy, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { ConfigBottomPanel } from '../components/ConfigBottomPanel'
import { FormField, ImageField, RecordList } from '../components/ConfigControls'
import { toStringValue } from '../lib/configValues'
import { rowId } from '../lib/luban'
import { configTableLabel } from '../lib/configLabels'
import { useConfigStore } from '../store/configStore'

const fileName = '#Team.xlsx'

export function TeamConfigPage() {
  const document = useConfigStore((state) => state.documents[fileName])
  const busy = useConfigStore((state) => state.busy)
  const updateCell = useConfigStore((state) => state.updateCell)
  const addRow = useConfigStore((state) => state.addRow)
  const duplicateRow = useConfigStore((state) => state.duplicateRow)
  const deleteRow = useConfigStore((state) => state.deleteRow)
  const changeId = useConfigStore((state) => state.changeId)
  const referencesFor = useConfigStore((state) => state.referencesFor)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [search, setSearch] = useState('')
  const [idEdit, setIdEdit] = useState({ key: '', value: '' })
  const [referenceState, setReferenceState] = useState<{ key: string; labels: string[] }>({ key: '', labels: [] })
  const row = document?.rows[selectedIndex]
  const recordKey = row ? `${selectedIndex}:${rowId(row)}` : ''
  const idDraft = idEdit.key === recordKey ? idEdit.value : row ? rowId(row) : ''
  const referenceLabels = referenceState.key === recordKey ? referenceState.labels : []
  const filteredRows = useMemo(() => (document?.rows ?? []).filter((candidate) => Object.values(candidate).some((value) => String(value ?? '').toLowerCase().includes(search.toLowerCase()))), [document?.rows, search])

  useEffect(() => {
    let active = true
    if (!row) return
    void referencesFor(fileName, selectedIndex).then((references) => {
      if (active) setReferenceState({ key: recordKey, labels: references.map((reference) => `${reference.sourceFile}.${reference.field} (${reference.sourceRowId})`) })
    })
    return () => { active = false }
  }, [recordKey, referencesFor, row, selectedIndex])

  if (!document) return <div className="configWorkspacePage"><div className="emptyState large">正在读取 #Team.xlsx…</div><ConfigBottomPanel fileName={fileName} /></div>

  function selectFiltered(index: number) {
    const selected = filteredRows[index]
    const original = document!.rows.indexOf(selected)
    if (original >= 0) setSelectedIndex(original)
  }

  function createRow() {
    addRow(fileName)
    setSelectedIndex(document!.rows.length)
  }

  async function removeRow() {
    if (!row || !window.confirm(`确定删除战队 ${rowId(row) || '未命名记录'}？`)) return
    if (await deleteRow(fileName, selectedIndex)) setSelectedIndex(Math.max(0, selectedIndex - 1))
  }

  return (
    <div className="configWorkspacePage">
      <section className="configWorkspaceMain recordEditorLayout">
        <aside className="configSidebar">
          <div className="panelHeading"><div><h1>{configTableLabel(document.tableName, '战队配置')}</h1><p>{document.rows.length} 支战队 {document.dirty ? '· 未保存' : ''}</p></div></div>
          <input className="searchInput" value={search} placeholder="搜索战队" onChange={(event) => setSearch(event.target.value)} />
          <div className="rowCommands compact"><button type="button" className="iconButton" title="新增战队" onClick={createRow}><Plus size={16} /></button><button type="button" className="iconButton" title="复制战队" disabled={!row} onClick={() => duplicateRow(fileName, selectedIndex)}><Copy size={16} /></button><button type="button" className="iconButton danger" title="删除战队" disabled={!row || busy} onClick={() => void removeRow()}><Trash2 size={16} /></button></div>
          <RecordList rows={filteredRows} selectedIndex={filteredRows.indexOf(row!)} onSelect={selectFiltered} secondary={(candidate) => `${toStringValue(candidate.shortName)} · ${rowId(candidate)}`} />
        </aside>
        <section className="recordDetailPane">
          {!row ? <div className="emptyState large">新增或选择一支战队。</div> : (
            <div className="detailForm">
              <header className="detailHeader"><div><h2>{toStringValue(row.name) || '未命名战队'}</h2><p>{rowId(row)} · 被 {referenceLabels.length} 处引用</p></div></header>
              <section className="formSection"><h3>标识与名称</h3><div className="formGrid twoColumns"><FormField label="id" hint={referenceLabels.length > 0 ? '修改时会明确列出并同步引用' : '当前没有外部引用'}><div className="inlineField"><input value={idDraft} onChange={(event) => setIdEdit({ key: recordKey, value: event.target.value })} /><button type="button" className="commandButton" disabled={!idDraft || idDraft === rowId(row)} onClick={() => void changeId(fileName, selectedIndex, idDraft)}><RefreshCw size={14} />应用 ID</button></div></FormField><FormField label="name"><input value={toStringValue(row.name)} onChange={(event) => updateCell(fileName, selectedIndex, 'name', event.target.value)} /></FormField><FormField label="shortName"><input value={toStringValue(row.shortName)} onChange={(event) => updateCell(fileName, selectedIndex, 'shortName', event.target.value)} /></FormField><FormField label="nickname"><input value={toStringValue(row.nickname)} onChange={(event) => updateCell(fileName, selectedIndex, 'nickname', event.target.value)} /></FormField></div></section>
              <section className="formSection"><h3>战队 Logo</h3><ImageField label="logo" kind="team" value={toStringValue(row.logo)} onChange={(value) => updateCell(fileName, selectedIndex, 'logo', value)} /></section>
              <section className="formSection"><h3>引用位置</h3>{referenceLabels.length === 0 ? <p className="emptyState">当前没有 Player 或 TutorialBattle 引用该战队。</p> : <div className="referenceList">{referenceLabels.map((label) => <p key={label}>{label}</p>)}</div>}</section>
            </div>
          )}
        </section>
      </section>
      <ConfigBottomPanel fileName={fileName} />
    </div>
  )
}

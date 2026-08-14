import { Copy, ExternalLink, Plus, Trash2 } from 'lucide-react'
import { useMemo, useState } from 'react'
import { ConfigBottomPanel } from '../components/ConfigBottomPanel'
import { displayLabel, rowId, type LubanCellValue, type LubanField, type LubanRow } from '../lib/luban'
import { configFieldLabel, configTableLabel } from '../lib/configLabels'
import { useConfigStore } from '../store/configStore'

export function AllTablesPage({ onOpenMap }: { onOpenMap: () => void }) {
  const summaries = useConfigStore((state) => state.summaries)
  const documents = useConfigStore((state) => state.documents)
  const activeFile = useConfigStore((state) => state.activeFile)
  const setActiveFile = useConfigStore((state) => state.setActiveFile)
  const updateCell = useConfigStore((state) => state.updateCell)
  const addRow = useConfigStore((state) => state.addRow)
  const duplicateRow = useConfigStore((state) => state.duplicateRow)
  const deleteRow = useConfigStore((state) => state.deleteRow)
  const [selectedRow, setSelectedRow] = useState(0)
  const [search, setSearch] = useState('')
  const summary = summaries.find((item) => item.fileName === activeFile)
  const document = activeFile ? documents[activeFile] : undefined
  const visible = useMemo(() => summaries.filter((item) => item.fileName.toLowerCase().includes(search.toLowerCase()) || item.tableName.toLowerCase().includes(search.toLowerCase())), [search, summaries])

  return (
    <div className="configWorkspacePage">
      <section className="configWorkspaceMain allTablesLayout">
        <aside className="configSidebar">
          <div className="panelHeading"><div><h1>全部表格</h1><p>{summaries.length} 张 Luban auto-import 表</p></div></div>
          <input className="searchInput" value={search} placeholder="搜索文件或表名" onChange={(event) => setSearch(event.target.value)} />
          <div className="tableFileList">{visible.map((item) => <button type="button" key={item.fileName} className={activeFile === item.fileName ? 'tableFileButton active' : 'tableFileButton'} onClick={() => { setSelectedRow(0); setActiveFile(item.fileName) }}><span><strong>{item.fileName}</strong><small>{item.tableName} · {item.rows} 行</small></span><em className={`ownerBadge ${item.owner}`}>{ownerLabel(item.owner)}</em></button>)}</div>
        </aside>
        <section className="tableEditorArea">
          {!summary ? <div className="emptyState large">选择一张表开始查看。</div> : null}
          {summary?.owner === 'map' ? <div className="mapOwnedNotice"><ExternalLink size={28} /><h2>{summary.fileName} 由地图配置维护</h2><p>该表依赖地图工程状态和强校验，不能从通用表格直接修改。</p><button type="button" className="primaryButton" onClick={onOpenMap}>前往地图配置</button></div> : null}
          {summary && summary.owner !== 'map' && !document ? <div className="emptyState large">{summary.warnings.join('；') || '表格读取失败'}</div> : null}
          {document && summary?.owner !== 'map' ? (
            <>
              <header className="tableEditorHeader"><div><h2>{document.fileName}</h2><p>{configTableLabel(document.tableName)} · {document.rows.length} 行 {document.dirty ? '· 有未保存修改' : ''}</p></div><div className="rowCommands"><button type="button" className="commandButton" onClick={() => addRow(document.fileName)}><Plus size={15} />新增行</button><button type="button" className="commandButton" disabled={!document.rows[selectedRow]} onClick={() => duplicateRow(document.fileName, selectedRow)}><Copy size={15} />复制行</button><button type="button" className="dangerButton" disabled={!document.rows[selectedRow]} onClick={() => void deleteRow(document.fileName, selectedRow)}><Trash2 size={15} />删除行</button></div></header>
              <div className="genericTableScroll"><table className="genericConfigTable"><thead><tr><th>#</th>{document.fields.map((field) => <th key={field.key} title={`${field.type}\n${field.comment}`}>{configFieldLabel(field.key, field.comment)}<small>{field.type}</small></th>)}</tr></thead><tbody>{document.rows.map((row, rowIndex) => <tr key={`${rowId(row)}-${rowIndex}`} className={selectedRow === rowIndex ? 'selected' : ''} onClick={() => setSelectedRow(rowIndex)}><td>{rowIndex + 1}</td>{document.fields.map((field) => <td key={field.key}><CellEditor field={field} row={row} documents={documents} onChange={(value) => updateCell(document.fileName, rowIndex, field.key, value)} /></td>)}</tr>)}</tbody></table></div>
            </>
          ) : null}
        </section>
      </section>
      <ConfigBottomPanel fileName={document?.fileName} />
    </div>
  )
}

function CellEditor({ field, row, documents, onChange }: { field: LubanField; row: LubanRow; documents: Record<string, import('../lib/luban').LubanTableDocument>; onChange: (value: LubanCellValue) => void }) {
  const value = row[field.key]
  if (field.kind === 'bool') return <input type="checkbox" checked={value === true} onChange={(event) => onChange(event.target.checked)} />
  if (field.kind === 'ref') {
    const target = Object.values(documents).find((document) => document.tableName === field.refTable)
    if (field.list) {
      const selected = Array.isArray(value) ? value.map(String) : []
      return <select multiple value={selected} onChange={(event) => onChange([...event.target.selectedOptions].map((option) => option.value))}>{target?.rows.map((candidate) => <option key={rowId(candidate)} value={rowId(candidate)}>{displayLabel(candidate)}</option>)}</select>
    }
    return <select value={value == null || Array.isArray(value) ? '' : String(value)} onChange={(event) => onChange(event.target.value)}><option value="">请选择</option>{target?.rows.map((candidate) => <option key={rowId(candidate)} value={rowId(candidate)}>{displayLabel(candidate)}</option>)}</select>
  }
  if (field.list) return <input value={Array.isArray(value) ? value.join(',') : ''} onChange={(event) => onChange(event.target.value.split(',').map((item) => item.trim()).filter(Boolean))} />
  if (field.kind === 'int' || field.kind === 'float') return <input type="number" step={field.kind === 'int' ? 1 : 'any'} value={value == null || Array.isArray(value) ? '' : String(value)} onChange={(event) => onChange(event.target.value === '' ? null : Number(event.target.value))} />
  return <input title={`${field.comment} (${field.type})`} value={value == null || Array.isArray(value) ? '' : String(value)} onChange={(event) => onChange(event.target.value)} />
}

function ownerLabel(owner: string): string {
  if (owner === 'map') return '地图配置'
  if (owner === 'player') return '选手专属'
  if (owner === 'team') return '战队专属'
  if (owner === 'tutorial') return '新手战斗'
  return '通用编辑'
}

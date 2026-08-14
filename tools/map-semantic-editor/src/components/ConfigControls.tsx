import { ImagePlus, X } from 'lucide-react'
import { useRef, useState } from 'react'
import { displayLabel, rowId, type LubanRow } from '../lib/luban'
import { configFieldLabel } from '../lib/configLabels'
import { configAssetUrl } from '../lib/api'
import { useConfigStore } from '../store/configStore'

export function FormField({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return <label className="configField"><span>{configFieldLabel(label)}</span>{children}{hint ? <small>{hint}</small> : null}</label>
}

export function ImageField({ label, kind, value, onChange }: { label: string; kind: 'portrait' | 'team' | 'player-card'; value: string; onChange: (value: string) => void }) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [missing, setMissing] = useState(false)
  const uploadImage = useConfigStore((state) => state.uploadImage)

  async function selectFile(file?: File) {
    if (!file) return
    const assetPath = await uploadImage(kind, file)
    if (assetPath) {
      setMissing(false)
      onChange(assetPath)
    }
    if (inputRef.current) inputRef.current.value = ''
  }

  return (
    <div className="imageField">
      <div className="imagePreview">
        {value && !missing ? <img src={configAssetUrl(value)} alt={`${label}预览`} onLoad={() => setMissing(false)} onError={() => setMissing(true)} /> : <span>{missing ? '图片缺失' : '未选择图片'}</span>}
      </div>
      <div className="imageFieldControls">
        <FormField label={label}><input value={value} onChange={(event) => { setMissing(false); onChange(event.target.value) }} /></FormField>
        <input ref={inputRef} className="visuallyHidden" type="file" accept="image/png,image/jpeg,image/webp" onChange={(event) => void selectFile(event.target.files?.[0])} />
        <button type="button" className="commandButton" onClick={() => inputRef.current?.click()}><ImagePlus size={16} />选择并复制图片</button>
        {missing ? <p className="fieldWarning">文件不存在或无法预览，仍可保存。</p> : null}
      </div>
    </div>
  )
}

export function TagEditor({ value, onChange }: { value: string[]; onChange: (value: string[]) => void }) {
  const [draft, setDraft] = useState('')

  function addTag() {
    const additions = draft.split(',').map((item) => item.trim()).filter(Boolean)
    if (additions.length > 0) onChange([...new Set([...value, ...additions])])
    setDraft('')
  }

  return (
    <div className="tagEditor">
      <div className="tagList">{value.map((tag) => <span className="tag" key={tag}>{tag}<button type="button" title={`移除 ${tag}`} onClick={() => onChange(value.filter((item) => item !== tag))}><X size={12} /></button></span>)}</div>
      <input value={draft} placeholder="输入标签后按 Enter" onChange={(event) => setDraft(event.target.value)} onKeyDown={(event) => { if (event.key === 'Enter') { event.preventDefault(); addTag() } }} onBlur={addTag} />
    </div>
  )
}

export function ReferenceSelect({ value, rows, onChange, placeholder = '请选择' }: { value: string; rows: LubanRow[]; onChange: (value: string) => void; placeholder?: string }) {
  return <select value={value} onChange={(event) => onChange(event.target.value)}><option value="">{placeholder}</option>{rows.map((row) => <option key={rowId(row)} value={rowId(row)}>{displayLabel(row)} ({rowId(row)})</option>)}</select>
}

export function RecordList({ rows, selectedIndex, onSelect, secondary }: { rows: LubanRow[]; selectedIndex: number; onSelect: (index: number) => void; secondary?: (row: LubanRow) => string }) {
  return <div className="recordList">{rows.map((row, index) => <button type="button" key={`${rowId(row)}-${index}`} className={selectedIndex === index ? 'recordButton active' : 'recordButton'} onClick={() => onSelect(index)}><strong>{displayLabel(row)}</strong><span>{secondary?.(row) ?? rowId(row)}</span></button>)}</div>
}

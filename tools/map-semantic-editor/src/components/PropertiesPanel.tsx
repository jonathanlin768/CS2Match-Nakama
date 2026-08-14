import { Check, Copy, PenLine, Trash2, X } from 'lucide-react'
import { nodeShapes, nodeTypes, nodeUsages, floors, sides, sites, pointsToCell } from '../lib/model'
import { useEditorStore } from '../store/editorStore'
import { configFieldLabel } from '../lib/configLabels'

export function PropertiesPanel() {
  const project = useEditorStore((state) => state.project)
  const selected = useEditorStore((state) => state.selected)
  const updateNode = useEditorStore((state) => state.updateNode)
  const updateEdge = useEditorStore((state) => state.updateEdge)
  const updateVisibility = useEditorStore((state) => state.updateVisibility)
  const updateRoute = useEditorStore((state) => state.updateRoute)
  const setNodePointsFromText = useEditorStore((state) => state.setNodePointsFromText)
  const duplicateSelected = useEditorStore((state) => state.duplicateSelected)
  const deleteSelected = useEditorStore((state) => state.deleteSelected)
  const previewSamples = useEditorStore((state) => state.previewSamples)
  const clearPreview = useEditorStore((state) => state.clearPreview)
  const samples = useEditorStore((state) => state.samples)
  const rangeDraft = useEditorStore((state) => state.rangeDraft)
  const startRangeDraw = useEditorStore((state) => state.startRangeDraw)
  const finishRangeDraw = useEditorStore((state) => state.finishRangeDraw)
  const cancelRangeDraw = useEditorStore((state) => state.cancelRangeDraw)

  const node = selected?.kind === 'node' ? project.nodes.find((item) => item.id === selected.id) : undefined
  const edge = selected?.kind === 'edge' ? project.edges.find((item) => item.id === selected.id) : undefined
  const visibility = selected?.kind === 'visibility' ? project.visibility.find((item) => item.id === selected.id) : undefined
  const route = selected?.kind === 'route' ? project.routes.find((item) => item.id === selected.id) : undefined

  return (
    <aside className="sidePanel propertyPanel">
      <section className="panelSection">
        <h2>右侧属性</h2>
        {!selected ? <p className="emptyHint">选择对象后编辑字段。</p> : null}

        {node ? (
          <div className="formGrid">
            <h3>当前对象: MapNode</h3>
            <TextInput label="id" value={node.id} onChange={(value) => updateNode(node.id, { id: value })} />
            <TextInput label="name" value={node.name} onChange={(value) => updateNode(node.id, { name: value })} />
            <TextInput label="zone" value={node.zone} onChange={(value) => updateNode(node.id, { zone: value })} />
            <SelectInput label="site" value={node.site} options={sites} onChange={(value) => updateNode(node.id, { site: value })} />
            <SelectInput label="node_type" value={node.node_type} options={nodeTypes} onChange={(value) => updateNode(node.id, { node_type: value })} />
            <SelectInput label="default_side" value={node.default_side} options={sides} onChange={(value) => updateNode(node.id, { default_side: value })} />
            <SelectInput label="floor" value={node.floor} options={floors} onChange={(value) => updateNode(node.id, { floor: value })} />
            <NumberInput label="x" value={node.x} step={0.001} onChange={(value) => updateNode(node.id, { x: value })} />
            <NumberInput label="y" value={node.y} step={0.001} onChange={(value) => updateNode(node.id, { y: value })} />
            <SelectInput label="shape" value={node.shape} options={nodeShapes} onChange={(value) => updateNode(node.id, { shape: value })} />
            <NumberInput label="radius" value={node.radius ?? 0} step={0.001} onChange={(value) => updateNode(node.id, { radius: value })} />
            <label className="field full">
              <span>{configFieldLabel('area_usages')}</span>
              <div className="checkboxGroup">
                {nodeUsages.map((usage) => (
                  <label key={usage}>
                    <input
                      type="checkbox"
                      checked={node.area_usages.includes(usage)}
                      onChange={(event) => {
                        const next = event.target.checked ? [...node.area_usages, usage] : node.area_usages.filter((item) => item !== usage)
                        updateNode(node.id, { area_usages: next })
                      }}
                    />
                    {usage}
                  </label>
                ))}
              </div>
            </label>
            <label className="field full">
              <span>{configFieldLabel('points')}</span>
              <textarea value={pointsToCell(node.points)} onChange={(event) => setNodePointsFromText(node.id, event.target.value)} rows={3} />
            </label>
            <div className="rangeActions">
              <button type="button" className="iconTextButton" onClick={() => startRangeDraw(node.id, 'Circle')}><PenLine size={15} />划定圆形范围</button>
              <button type="button" className="iconTextButton" onClick={() => startRangeDraw(node.id, 'Polygon')}><PenLine size={15} />划定多边形范围</button>
              {rangeDraft?.nodeId === node.id ? (
                <>
                  <button type="button" className="primaryButton" onClick={finishRangeDraw}><Check size={15} />完成范围</button>
                  <button type="button" className="commandButton" onClick={cancelRangeDraw}><X size={15} />取消范围</button>
                </>
              ) : null}
            </div>
            <div className="propertyActions">
              <button
                type="button"
                className={samples.length > 0 ? 'commandButton active' : 'commandButton'}
                onClick={samples.length > 0 ? clearPreview : previewSamples}
              >
                {samples.length > 0 ? '清除预览' : '预览采样点'}
              </button>
              <button type="button" className="iconTextButton" onClick={duplicateSelected}><Copy size={15} />复制</button>
              <button type="button" className="dangerButton" onClick={deleteSelected}><Trash2 size={15} />删除</button>
            </div>
          </div>
        ) : null}

        {edge ? (
          <div className="formGrid">
            <h3>当前对象: MapEdge</h3>
            <TextInput label="id" value={edge.id} onChange={(value) => updateEdge(edge.id, { id: value })} />
            <NodeSelect label="from" value={edge.from} onChange={(value) => updateEdge(edge.id, { from: value })} />
            <NodeSelect label="to" value={edge.to} onChange={(value) => updateEdge(edge.id, { to: value })} />
            <NumberInput label="base_time" value={edge.base_time} step={1} onChange={(value) => updateEdge(edge.id, { base_time: Math.round(value) })} />
            <NumberInput label="stamina_cost" value={edge.stamina_cost} step={1} onChange={(value) => updateEdge(edge.id, { stamina_cost: Math.round(value) })} />
            <NumberInput label="risk" value={edge.risk} step={1} onChange={(value) => updateEdge(edge.id, { risk: Math.round(value) })} />
            <NumberInput label="noise" value={edge.noise} step={1} onChange={(value) => updateEdge(edge.id, { noise: Math.round(value) })} />
            <TextInput label="risk_points" value={edge.risk_points.join(',')} onChange={(value) => updateEdge(edge.id, { risk_points: splitList(value) })} />
            <TextInput label="intercept_nodes" value={edge.intercept_nodes.join(',')} onChange={(value) => updateEdge(edge.id, { intercept_nodes: splitList(value) })} />
            <label className="field toggle"><span>{configFieldLabel('bidirectional')}</span><input type="checkbox" checked={edge.bidirectional} onChange={(event) => updateEdge(edge.id, { bidirectional: event.target.checked })} /></label>
            <DeleteAction onDelete={deleteSelected} />
          </div>
        ) : null}

        {visibility ? (
          <div className="formGrid">
            <h3>当前对象: Visibility</h3>
            <TextInput label="id" value={visibility.id} onChange={(value) => updateVisibility(visibility.id, { id: value })} />
            <NodeSelect label="from" value={visibility.from} onChange={(value) => updateVisibility(visibility.id, { from: value })} />
            <NodeSelect label="to" value={visibility.to} onChange={(value) => updateVisibility(visibility.id, { to: value })} />
            <SelectInput label="range" value={visibility.range} options={['Close', 'Mid', 'Long']} onChange={(value) => updateVisibility(visibility.id, { range: value })} />
            <SelectInput label="angle_advantage" value={visibility.angle_advantage} options={['T', 'CT', 'None']} onChange={(value) => updateVisibility(visibility.id, { angle_advantage: value })} />
 <SelectInput label="elevation" value={visibility.elevation} options={['HighToLow', 'LowToHigh', 'SameLevel', 'Same', 'HeightBlocked']} onChange={(value) => updateVisibility(visibility.id, { elevation: value })} />
            <NumberInput label="cover_modifier" value={visibility.cover_modifier} step={1} onChange={(value) => updateVisibility(visibility.id, { cover_modifier: Math.round(value) })} />
            <NumberInput label="exposure_modifier" value={visibility.exposure_modifier} step={1} onChange={(value) => updateVisibility(visibility.id, { exposure_modifier: Math.round(value) })} />
            <label className="field toggle"><span>{configFieldLabel('visible')}</span><input type="checkbox" checked={visibility.visible} onChange={(event) => updateVisibility(visibility.id, { visible: event.target.checked })} /></label>
            <DeleteAction onDelete={deleteSelected} />
          </div>
        ) : null}

        {route ? (
          <div className="formGrid">
            <h3>当前对象: Route</h3>
            <TextInput label="id" value={route.id} onChange={(value) => updateRoute(route.id, { id: value })} />
            <TextInput label="name" value={route.name} onChange={(value) => updateRoute(route.id, { name: value })} />
            <SelectInput label="side" value={route.side} options={['T', 'CT']} onChange={(value) => updateRoute(route.id, { side: value })} />
            <SelectInput label="target_site" value={route.target_site} options={sites} onChange={(value) => updateRoute(route.id, { target_site: value })} />
            <TextInput label="nodes" value={route.nodes.join(',')} onChange={(value) => updateRoute(route.id, { nodes: splitList(value) })} />
            <NumberInput label="min_players" value={route.min_players} step={1} onChange={(value) => updateRoute(route.id, { min_players: Math.round(value) })} />
            <NumberInput label="max_players" value={route.max_players} step={1} onChange={(value) => updateRoute(route.id, { max_players: Math.round(value) })} />
            <TextInput label="style_tags" value={route.style_tags.join(',')} onChange={(value) => updateRoute(route.id, { style_tags: splitList(value) })} />
            <DeleteAction onDelete={deleteSelected} />
          </div>
        ) : null}
      </section>
    </aside>
  )
}

function TextInput({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="field">
      <span>{configFieldLabel(label)}</span>
      <input value={value} onChange={(event) => onChange(event.target.value)} />
    </label>
  )
}

function NumberInput({ label, value, step, onChange }: { label: string; value: number; step: number; onChange: (value: number) => void }) {
  return (
    <label className="field">
      <span>{configFieldLabel(label)}</span>
      <input type="number" value={value} step={step} onChange={(event) => onChange(Number(event.target.value))} />
    </label>
  )
}

function SelectInput<T extends string>({ label, value, options, onChange }: { label: string; value: T; options: readonly T[]; onChange: (value: T) => void }) {
  return (
    <label className="field">
      <span>{configFieldLabel(label)}</span>
      <select value={value} onChange={(event) => onChange(event.target.value as T)}>
        {options.map((option) => <option key={option} value={option}>{option}</option>)}
      </select>
    </label>
  )
}

function NodeSelect({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  const nodes = useEditorStore((state) => state.project.nodes)
  return (
    <label className="field">
      <span>{configFieldLabel(label)}</span>
      <select value={value} onChange={(event) => onChange(event.target.value)}>
        {nodes.map((node) => <option key={node.id} value={node.id}>{node.id}</option>)}
      </select>
    </label>
  )
}

function DeleteAction({ onDelete }: { onDelete: () => void }) {
  return (
    <div className="propertyActions">
      <button type="button" className="dangerButton" onClick={onDelete}><Trash2 size={15} />删除</button>
    </div>
  )
}

function splitList(value: string): string[] {
  return value.split(',').map((item) => item.trim()).filter(Boolean)
}

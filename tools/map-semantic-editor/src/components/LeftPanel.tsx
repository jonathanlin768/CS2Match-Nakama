import { Lock, Unlock } from 'lucide-react'
import { useEditorStore } from '../store/editorStore'

const layerLabels: Record<string, string> = {
  nodes: '地图节点',
  ranges: '节点范围',
  edges: '路径',
  visibility: '视野',
  routes: '路线',
}

const visibleLayerKeys = ['nodes', 'ranges', 'edges', 'visibility', 'routes']

export function LeftPanel() {
  const project = useEditorStore((state) => state.project)
  const selected = useEditorStore((state) => state.selected)
  const updateLayer = useEditorStore((state) => state.updateLayer)
  const select = useEditorStore((state) => state.select)

  return (
    <aside className="sidePanel leftPanel">
      <section className="panelSection">
        <h2>项目</h2>
        <dl className="metaList">
          <div><dt>map_id</dt><dd>{project.map_id}</dd></div>
          <div><dt>radar</dt><dd>{project.radar_image}</dd></div>
          <div><dt>坐标</dt><dd>normalized 0..1</dd></div>
        </dl>
      </section>

      <section className="panelSection">
        <h2>图层</h2>
        <div className="layerList">
          {visibleLayerKeys.flatMap((key) => {
            const layer = project.layers[key]
            if (!layer) return []
            return (
            <div className="layerRow" key={key}>
              <input type="checkbox" checked={layer.visible} onChange={(event) => updateLayer(key, { visible: event.target.checked })} />
              <span className="swatch" style={{ background: layer.color }} />
              <span>{layerLabels[key] ?? key}</span>
              <button type="button" className="iconButton small" onClick={() => updateLayer(key, { locked: !layer.locked })} title={layer.locked ? '解锁图层' : '锁定图层'}>
                {layer.locked ? <Lock size={14} /> : <Unlock size={14} />}
              </button>
            </div>
            )
          })}
        </div>
      </section>

      <section className="panelSection objectList">
        <h2>对象列表</h2>
        <h3>MapNode</h3>
        {project.nodes.map((node) => (
          <button key={node.id} type="button" className={selected?.kind === 'node' && selected.id === node.id ? 'objectButton active' : 'objectButton'} onClick={() => select({ kind: 'node', id: node.id })}>
            <span>{node.id}</span>
            <small>{node.area_usages.join(',') || node.node_type}</small>
          </button>
        ))}
        <h3>MapEdge</h3>
        {project.edges.map((edge) => (
          <button key={edge.id} type="button" className={selected?.kind === 'edge' && selected.id === edge.id ? 'objectButton active' : 'objectButton'} onClick={() => select({ kind: 'edge', id: edge.id })}>
            <span>{edge.id}</span>
            <small>{edge.from} {'->'} {edge.to}</small>
          </button>
        ))}
        <h3>Route</h3>
        {project.routes.map((route) => (
          <button key={route.id} type="button" className={selected?.kind === 'route' && selected.id === route.id ? 'objectButton active' : 'objectButton'} onClick={() => select({ kind: 'route', id: route.id })}>
            <span>{route.id}</span>
            <small>{route.nodes.length} nodes</small>
          </button>
        ))}
      </section>
    </aside>
  )
}

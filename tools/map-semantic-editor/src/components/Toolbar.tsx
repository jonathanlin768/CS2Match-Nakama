import { useState } from 'react'
import { Check, CheckCircle2, Download, FilePlus2, FolderOpen, GitBranch, MousePointer2, Play, Radar, Redo2, Route, Save, Search, ShieldAlert, Undo2, Waypoints, X } from 'lucide-react'
import { useEditorStore } from '../store/editorStore'
import type { ToolMode } from '../lib/model'

const toolButtons: { mode: ToolMode; label: string; title: string }[] = [
  { mode: 'select', label: '选择', title: 'S: 选择、移动、框选和删除对象' },
  { mode: 'node', label: '点位 (tb_map_node)', title: 'N: 创建或编辑 tb_map_node' },
  { mode: 'edge', label: '路径 (tb_map_edge)', title: 'E: 连接两个 tb_map_node 生成 tb_map_edge' },
  { mode: 'visibility', label: '视野 (tb_visibility)', title: 'V: 连接观察点和被观察点生成 tb_visibility' },
  { mode: 'route', label: '路线 (tb_route)', title: 'L: 按节点序列生成 tb_route' },
  { mode: 'risk', label: '风险热点 (risk_points)', title: 'R: 创建 Risk 用途 tb_map_node 作为 tb_map_edge.risk_points 的风险热点' },
]

export function Toolbar() {
  const [showNewProjectDialog, setShowNewProjectDialog] = useState(false)
  const project = useEditorStore((state) => state.project)
  const tool = useEditorStore((state) => state.tool)
  const busy = useEditorStore((state) => state.busy)
  const genConfig = useEditorStore((state) => state.genConfig)
  const serviceMessage = useEditorStore((state) => state.serviceMessage)
  const newProject = useEditorStore((state) => state.newProject)
  const setTool = useEditorStore((state) => state.setTool)
  const load = useEditorStore((state) => state.load)
  const loadPublished = useEditorStore((state) => state.loadPublished)
  const save = useEditorStore((state) => state.save)
  const undo = useEditorStore((state) => state.undo)
  const redo = useEditorStore((state) => state.redo)
  const validate = useEditorStore((state) => state.validate)
  const previewSamples = useEditorStore((state) => state.previewSamples)
  const clearPreview = useEditorStore((state) => state.clearPreview)
  const samples = useEditorStore((state) => state.samples)
  const routeDraft = useEditorStore((state) => state.routeDraft)
  const finishRouteDraft = useEditorStore((state) => state.finishRouteDraft)
  const cancelRouteDraft = useEditorStore((state) => state.cancelRouteDraft)
  const write = useEditorStore((state) => state.write)
  const runExport = useEditorStore((state) => state.runExport)
  const centerView = useEditorStore((state) => state.centerView)

  function confirmNewProject() {
    newProject()
    setShowNewProjectDialog(false)
  }

  return (
    <>
      <header className="toolbar">
        <div className="titleStrip">
          <div className="brand">
            <Radar size={18} />
            <span>CS2 Map Semantic Editor</span>
          </div>
          <button type="button" className="commandButton" onClick={() => setShowNewProjectDialog(true)} title="新建 de_dust2 工程模板">
            <FilePlus2 size={16} />
            新建
          </button>
          <button type="button" className="commandButton" onClick={() => void load()} title="从本地工程文件读取">
            <FolderOpen size={16} />
            打开
          </button>
          <button type="button" className="commandButton" disabled={busy} onClick={() => void loadPublished()} title={`从 configs/Datas/${project.map_id}.json 读取工程快照`}>
            <FolderOpen size={16} />
            从项目中读取
          </button>
          <button type="button" className="commandButton" onClick={() => void save()} title="保存工程文件">
            <Save size={16} />
            保存
          </button>
          <span className="mapBadge">{project.map_id} {project.version}</span>
          <span className={serviceMessage.includes('失败') || serviceMessage.includes('ERROR') ? 'serviceBadge error' : 'serviceBadge'}>{serviceMessage}</span>
          <button type="button" className="primaryButton" disabled={busy} onClick={() => void write()} title="写入 configs/Datas/#*.xlsx">
            <Download size={16} />
            写入 Luban
          </button>
          <button type="button" className="primaryButton secondary" disabled={busy} onClick={() => void runExport()} title="执行 scripts/gen-config.ps1">
            <Play size={16} />
            {genConfig?.status === 'running' ? '导表中' : '运行导表'}
          </button>
        </div>

        <div className="toolStrip">
          {toolButtons.map((item) => (
            <button key={item.mode} type="button" className={tool === item.mode ? 'toolButton active' : 'toolButton'} onClick={() => setTool(item.mode)} title={item.title}>
              {iconFor(item.mode)}
              {item.label}
            </button>
          ))}
          <span className="toolSpacer" />
          <button type="button" className="iconButton" onClick={undo} title="撤销">
            <Undo2 size={16} />
          </button>
          <button type="button" className="iconButton" onClick={redo} title="重做">
            <Redo2 size={16} />
          </button>
          <button type="button" className="toolButton compact" onClick={() => validate()} title="运行强校验">
            <CheckCircle2 size={16} />
            校验
          </button>
          {routeDraft.length > 0 ? (
            <>
              <button type="button" className="primaryButton" onClick={finishRouteDraft} title="将当前点位序列写入 tb_route">
                <Check size={16} />
                完成路线
              </button>
              <button type="button" className="commandButton" onClick={cancelRouteDraft} title="放弃当前路线草稿">
                <X size={16} />
                取消路线
              </button>
            </>
          ) : null}
          <button
            type="button"
            className={samples.length > 0 ? 'toolButton compact active' : 'toolButton compact'}
            onClick={samples.length > 0 ? clearPreview : previewSamples}
            title={samples.length > 0 ? '清除画布上的采样预览点' : '预览选中节点或第一个 KillSample 节点'}
          >
            <Search size={16} />
            {samples.length > 0 ? '清除预览' : '预览'}
          </button>
          <button type="button" className="toolButton compact" onClick={centerView} title="重置画布视图">
            <GitBranch size={16} />
            居中
          </button>
        </div>
      </header>

      {showNewProjectDialog ? (
        <div className="modalBackdrop" role="presentation">
          <section className="confirmDialog" role="dialog" aria-modal="true" aria-labelledby="new-project-title">
            <h2 id="new-project-title">新建项目</h2>
            <p>即将新建项目，当前的修改将全部丢失，是否继续？</p>
            <div className="dialogActions">
              <button type="button" className="commandButton" onClick={() => setShowNewProjectDialog(false)}>
                取消
              </button>
              <button type="button" className="dangerButton" onClick={confirmNewProject}>
                继续新建
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </>
  )
}

function iconFor(mode: ToolMode) {
  if (mode === 'select') return <MousePointer2 size={16} />
  if (mode === 'node') return <Waypoints size={16} />
  if (mode === 'edge') return <Route size={16} />
  if (mode === 'visibility') return <Search size={16} />
  if (mode === 'route') return <GitBranch size={16} />
  return <ShieldAlert size={16} />
}

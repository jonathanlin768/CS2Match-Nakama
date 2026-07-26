import { useEffect, useMemo, useState } from 'react'
import { useEditorStore } from '../store/editorStore'

type Tab = 'issues' | 'preview' | 'summary' | 'write' | 'gen'

export function BottomPanel() {
  const [tab, setTab] = useState<Tab>('issues')
  const issues = useEditorStore((state) => state.issues)
  const samples = useEditorStore((state) => state.samples)
  const writeLog = useEditorStore((state) => state.writeLog)
  const genConfig = useEditorStore((state) => state.genConfig)
  const project = useEditorStore((state) => state.project)
  const locate = useEditorStore((state) => state.locate)

  useEffect(() => {
    if (genConfig) setTab('gen')
  }, [genConfig])

  const summary = useMemo(() => ({
    nodes: project.nodes.length,
    ranges: project.nodes.filter((node) => node.shape !== 'None').length,
    edges: project.edges.length,
    visibility: project.visibility.length,
    routes: project.routes.length,
  }), [project])

  return (
    <footer className="bottomPanel">
      <div className="tabStrip">
        <TabButton active={tab === 'issues'} onClick={() => setTab('issues')}>校验结果</TabButton>
        <TabButton active={tab === 'preview'} onClick={() => setTab('preview')}>采样预览</TabButton>
        <TabButton active={tab === 'summary'} onClick={() => setTab('summary')}>导出摘要</TabButton>
        <TabButton active={tab === 'write'} onClick={() => setTab('write')}>写入日志</TabButton>
        <TabButton active={tab === 'gen'} onClick={() => setTab('gen')}>导表输出</TabButton>
      </div>

      <div className="tabContent">
        {tab === 'issues' ? (
          issues.map((issue) => (
            <button key={issue.id} type="button" className={`logLine ${issue.severity.toLowerCase()}`} onClick={() => locate(issue.object)}>
              <strong>{issue.severity}</strong>
              <span>{issue.message}</span>
              <em>{issue.object ? '定位' : 'project'}</em>
            </button>
          ))
        ) : null}

        {tab === 'preview' ? (
          <div className="textOutput">
            <strong>采样点</strong>
            {samples.length === 0 ? <p>点击预览后显示随机采样点。</p> : samples.map((point, index) => <p key={`${point.x}-${point.y}-${index}`}>{index + 1}. x={point.x.toFixed(4)} y={point.y.toFixed(4)}</p>)}
          </div>
        ) : null}

        {tab === 'summary' ? (
          <div className="summaryGrid">
            <Metric label="MapNode" value={summary.nodes} />
            <Metric label="节点范围" value={summary.ranges} />
            <Metric label="MapEdge" value={summary.edges} />
            <Metric label="Visibility" value={summary.visibility} />
            <Metric label="Route" value={summary.routes} />
          </div>
        ) : null}

        {tab === 'write' ? (
          <div className="textOutput">
            {writeLog.length === 0 ? <p>写入 Luban 后显示每张表的路径、行数和备份。</p> : writeLog.map((entry) => (
              <p key={entry.file}>
                <strong>{entry.file}</strong> {entry.rows} rows {entry.backup ? `backup: ${entry.backup}` : 'new file'} {entry.warnings.join(' ')}
              </p>
            ))}
          </div>
        ) : null}

        {tab === 'gen' ? (
          <div className="terminalOutput">
            {!genConfig ? <p>点击运行导表后显示 scripts/gen-config.ps1 输出。</p> : (
              <>
                <p>状态: {genConfig.status} exit={genConfig.exitCode ?? 'null'} duration={(genConfig.durationMs / 1000).toFixed(1)}s</p>
                <h4>stdout</h4>
                <pre>{genConfig.stdout || '<empty>'}</pre>
                <h4>stderr</h4>
                <pre>{genConfig.stderr || '<empty>'}</pre>
              </>
            )}
          </div>
        ) : null}
      </div>
    </footer>
  )
}

function TabButton({ active, children, onClick }: { active: boolean; children: string; onClick: () => void }) {
  return <button type="button" className={active ? 'tabButton active' : 'tabButton'} onClick={onClick}>{children}</button>
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

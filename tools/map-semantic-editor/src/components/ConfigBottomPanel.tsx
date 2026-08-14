import { useState } from 'react'
import type { ConfigValidationIssue } from '../lib/luban'
import { useConfigStore } from '../store/configStore'

type BottomTab = 'issues' | 'logs' | 'gen'

export function ConfigBottomPanel({ fileName, onIssueClick }: { fileName?: string; onIssueClick?: (issue: ConfigValidationIssue) => void }) {
  const [tab, setTab] = useState<BottomTab>('issues')
  const issues = useConfigStore((state) => state.issues)
  const logs = useConfigStore((state) => state.logs)
  const genConfig = useConfigStore((state) => state.genConfig)
  const visibleIssues = fileName ? issues.filter((issue) => issue.fileName === fileName) : issues
  const [previousGenConfig, setPreviousGenConfig] = useState(genConfig)

  if (previousGenConfig !== genConfig) {
    setPreviousGenConfig(genConfig)
    if (genConfig) setTab('gen')
  }

  return (
    <footer className="configBottomPanel">
      <div className="configBottomTabs">
        <button type="button" className={tab === 'issues' ? 'active' : ''} onClick={() => setTab('issues')}>校验结果 ({visibleIssues.length})</button>
        <button type="button" className={tab === 'logs' ? 'active' : ''} onClick={() => setTab('logs')}>写入日志 ({logs.length})</button>
        <button type="button" className={tab === 'gen' ? 'active' : ''} onClick={() => setTab('gen')}>任务输出</button>
      </div>
      <div className="configBottomContent">
        {tab === 'issues' ? (
          visibleIssues.length === 0 ? <p className="emptyState">当前没有校验问题。</p> : visibleIssues.map((issue, index) => (
            <button type="button" key={`${issue.fileName}-${issue.rowIndex}-${issue.field}-${index}`} className={issue.severity === 'ERROR' ? 'issueLine error issueButton' : 'issueLine warning issueButton'} onClick={() => onIssueClick?.(issue)} disabled={!onIssueClick}>
              <strong>{issue.severity}</strong> {issue.fileName} {issue.rowId || `第 ${issue.rowIndex + 1} 行`}.{issue.field}: {issue.message}
            </button>
          ))
        ) : null}
        {tab === 'logs' ? (
          logs.length === 0 ? <p className="emptyState">保存、图片复制和 ID 同步结果会显示在这里。</p> : logs.map((entry, index) => <p key={`${entry.time}-${index}`} className={entry.level === 'error' ? 'issueLine error' : 'logLine'}>[{entry.time}] {entry.message}</p>)
        ) : null}
        {tab === 'gen' ? (
          !genConfig ? <p className="emptyState">运行导表或更新本地前后端后显示脚本输出。</p> : <div className="terminalOutput"><p>任务: {genConfig.operation === 'update-local' ? '更新本地前后端' : '运行导表'} | 状态: {genConfig.status} exit={genConfig.exitCode ?? 'null'} duration={(genConfig.durationMs / 1000).toFixed(1)}s</p><h4>stdout</h4><pre>{genConfig.stdout || '<empty>'}</pre><h4>stderr</h4><pre>{genConfig.stderr || '<empty>'}</pre></div>
        ) : null}
      </div>
    </footer>
  )
}

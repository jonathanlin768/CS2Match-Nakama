import { Database, Map, Play, RefreshCw, Save, SaveAll, ServerCog, Shield, Table2, Users } from 'lucide-react'
import { useEffect, useState } from 'react'
import { AllTablesPage } from '../pages/AllTablesPage'
import { MapConfigPage } from '../pages/MapConfigPage'
import { PlayerConfigPage } from '../pages/PlayerConfigPage'
import { TeamConfigPage } from '../pages/TeamConfigPage'
import { TutorialConfigPage } from '../pages/TutorialConfigPage'
import { useConfigStore } from '../store/configStore'

export type ConfigPage = 'map' | 'player' | 'team' | 'tutorial' | 'tables'

const pageFiles: Partial<Record<ConfigPage, string>> = {
  player: '#Player.xlsx',
  team: '#Team.xlsx',
  tutorial: '#TutorialBattle.xlsx',
}

export function ConfigEditorShell() {
  const [page, setPageState] = useState<ConfigPage>(pageFromHash())
  const activeFile = useConfigStore((state) => state.activeFile)
  const documents = useConfigStore((state) => state.documents)
  const busy = useConfigStore((state) => state.busy)
  const serviceMessage = useConfigStore((state) => state.serviceMessage)
  const reload = useConfigStore((state) => state.reload)
  const saveCurrent = useConfigStore((state) => state.saveCurrent)
  const saveAll = useConfigStore((state) => state.saveAll)
  const runExport = useConfigStore((state) => state.runExport)
  const updateLocal = useConfigStore((state) => state.updateLocal)
  const currentFile = pageFiles[page] ?? (page === 'tables' ? activeFile ?? undefined : undefined)
  const dirtyCount = Object.values(documents).filter((document) => document.dirty).length

  useEffect(() => {
    function syncHash() { setPageState(pageFromHash()) }
    window.addEventListener('hashchange', syncHash)
    return () => window.removeEventListener('hashchange', syncHash)
  }, [])

  function setPage(next: ConfigPage) {
    window.location.hash = next
    setPageState(next)
  }

  return (
    <main className="configShell">
      <header className="configHeader">
        <div className="configBrand"><Database size={19} /><strong>CS2 Config Editor</strong></div>
        <nav className="configNav" aria-label="配置页面">
          <NavButton active={page === 'map'} icon={<Map size={16} />} label="地图配置" onClick={() => setPage('map')} />
          <NavButton active={page === 'player'} icon={<Users size={16} />} label="选手配置" onClick={() => setPage('player')} />
          <NavButton active={page === 'team'} icon={<Shield size={16} />} label="战队配置" onClick={() => setPage('team')} />
          <NavButton active={page === 'tutorial'} icon={<Play size={16} />} label="新手战斗" onClick={() => setPage('tutorial')} />
          <NavButton active={page === 'tables'} icon={<Table2 size={16} />} label="全部表格" onClick={() => setPage('tables')} />
        </nav>
        <div className="configCommands">
          <span className={serviceMessage.includes('失败') || serviceMessage.includes('ERROR') || serviceMessage.includes('阻止') ? 'configStatus error' : 'configStatus'}>{serviceMessage}</span>
          {page !== 'map' ? (
            <>
              <button type="button" className="iconButton" title="从项目重新读取" disabled={busy} onClick={() => void reload()}><RefreshCw size={16} /></button>
              <button type="button" className="commandButton" disabled={busy || !currentFile || !documents[currentFile]?.dirty} onClick={() => void saveCurrent(currentFile)}><Save size={16} />保存当前表</button>
              <button type="button" className="commandButton" disabled={busy || dirtyCount === 0} onClick={() => void saveAll()}><SaveAll size={16} />保存全部{dirtyCount > 0 ? ` (${dirtyCount})` : ''}</button>
              <button type="button" className="primaryButton secondary" disabled={busy} onClick={() => void runExport()}><Play size={16} />运行导表</button>
              <button type="button" className="primaryButton secondary" disabled={busy} onClick={() => void updateLocal()} title="导表、编译后端插件、重建前端并重启本地 Docker 服务"><ServerCog size={16} />更新本地前后端</button>
            </>
          ) : null}
        </div>
      </header>
      <section className="configPageContent">
        {page === 'map' ? <MapConfigPage /> : null}
        {page === 'player' ? <PlayerConfigPage /> : null}
        {page === 'team' ? <TeamConfigPage /> : null}
        {page === 'tutorial' ? <TutorialConfigPage /> : null}
        {page === 'tables' ? <AllTablesPage onOpenMap={() => setPage('map')} /> : null}
      </section>
    </main>
  )
}

function NavButton({ active, icon, label, onClick }: { active: boolean; icon: React.ReactNode; label: string; onClick: () => void }) {
  return <button type="button" className={active ? 'configNavButton active' : 'configNavButton'} onClick={onClick}>{icon}<span>{label}</span></button>
}

function pageFromHash(): ConfigPage {
  const value = window.location.hash.replace(/^#/, '')
  return value === 'player' || value === 'team' || value === 'tutorial' || value === 'tables' ? value : 'map'
}

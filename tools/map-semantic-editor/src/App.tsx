import { useEffect } from 'react'
import { ConfigEditorShell } from './components/ConfigEditorShell'
import { useConfigStore } from './store/configStore'
import { useEditorStore } from './store/editorStore'

export default function App() {
  const loadMap = useEditorStore((state) => state.load)
  const loadConfig = useConfigStore((state) => state.load)
  const hasDirtyTables = useConfigStore((state) => Object.values(state.documents).some((document) => document.dirty))

  useEffect(() => {
    void loadMap()
    void loadConfig()
  }, [loadConfig, loadMap])

  useEffect(() => {
    function warnBeforeUnload(event: BeforeUnloadEvent) {
      if (!hasDirtyTables) return
      event.preventDefault()
    }
    window.addEventListener('beforeunload', warnBeforeUnload)
    return () => window.removeEventListener('beforeunload', warnBeforeUnload)
  }, [hasDirtyTables])

  return <ConfigEditorShell />
}

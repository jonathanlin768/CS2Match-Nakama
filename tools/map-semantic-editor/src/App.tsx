import { type PointerEvent, useEffect, useState } from 'react'
import { BottomPanel } from './components/BottomPanel'
import { LeftPanel } from './components/LeftPanel'
import { PropertiesPanel } from './components/PropertiesPanel'
import { RadarCanvas } from './components/RadarCanvas'
import { Toolbar } from './components/Toolbar'
import { toolForShortcutKey } from './lib/shortcuts'
import { useEditorStore } from './store/editorStore'

export default function App() {
  const [leftPanelWidth, setLeftPanelWidth] = useState(280)
  const [propertyPanelWidth, setPropertyPanelWidth] = useState(340)
  const [bottomPanelHeight, setBottomPanelHeight] = useState(188)
  const load = useEditorStore((state) => state.load)
  const deleteSelected = useEditorStore((state) => state.deleteSelected)
  const undo = useEditorStore((state) => state.undo)
  const redo = useEditorStore((state) => state.redo)
  const setTool = useEditorStore((state) => state.setTool)
  const routeDraft = useEditorStore((state) => state.routeDraft)
  const finishRouteDraft = useEditorStore((state) => state.finishRouteDraft)
  const cancelRouteDraft = useEditorStore((state) => state.cancelRouteDraft)
  const rangeDraft = useEditorStore((state) => state.rangeDraft)
  const finishRangeDraw = useEditorStore((state) => state.finishRangeDraw)
  const cancelRangeDraw = useEditorStore((state) => state.cancelRangeDraw)

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (isEditingText(event.target)) return

      if ((event.ctrlKey || event.metaKey) && !event.altKey) {
        const key = event.key.toLowerCase()
        if (key === 'z' && !event.shiftKey) {
          event.preventDefault()
          undo()
          return
        }
        if (key === 'y' || (key === 'z' && event.shiftKey)) {
          event.preventDefault()
          redo()
          return
        }
      }

      if (!event.ctrlKey && !event.metaKey && !event.altKey) {
        const shortcutTool = toolForShortcutKey(event.key)
        if (shortcutTool) {
          event.preventDefault()
          setTool(shortcutTool)
          return
        }
      }

      if (rangeDraft && event.key === 'Enter') {
        event.preventDefault()
        finishRangeDraw()
        return
      }

      if (rangeDraft && event.key === 'Escape') {
        event.preventDefault()
        cancelRangeDraw()
        return
      }

      if (routeDraft.length > 0 && event.key === 'Enter') {
        event.preventDefault()
        finishRouteDraft()
        return
      }

      if (routeDraft.length > 0 && event.key === 'Escape') {
        event.preventDefault()
        cancelRouteDraft()
        return
      }

      if (event.key !== 'Delete' && event.key !== 'Backspace') return
      event.preventDefault()
      deleteSelected()
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [cancelRangeDraw, cancelRouteDraft, deleteSelected, finishRangeDraw, finishRouteDraft, rangeDraft, redo, routeDraft.length, setTool, undo])

  function startPanelResize(kind: 'left' | 'property' | 'bottom', event: PointerEvent<HTMLDivElement>) {
    event.preventDefault()
    const startX = event.clientX
    const startY = event.clientY
    const startLeft = leftPanelWidth
    const startProperty = propertyPanelWidth
    const startBottom = bottomPanelHeight

    function handlePointerMove(moveEvent: globalThis.PointerEvent) {
      if (kind === 'left') {
        const maxLeft = Math.max(220, window.innerWidth - propertyPanelWidth - 552)
        setLeftPanelWidth(clamp(startLeft + moveEvent.clientX - startX, 220, maxLeft))
      }
      if (kind === 'property') {
        const maxProperty = Math.max(280, window.innerWidth - leftPanelWidth - 552)
        setPropertyPanelWidth(clamp(startProperty - (moveEvent.clientX - startX), 280, maxProperty))
      }
      if (kind === 'bottom') {
        setBottomPanelHeight(clamp(startBottom - (moveEvent.clientY - startY), 120, 420))
      }
    }

    function handlePointerUp() {
      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('pointerup', handlePointerUp)
      document.body.classList.remove('isResizingPanel')
    }

    document.body.classList.add('isResizingPanel')
    window.addEventListener('pointermove', handlePointerMove)
    window.addEventListener('pointerup', handlePointerUp)
  }

  return (
    <main className="appShell" style={{ gridTemplateRows: `auto minmax(0, 1fr) ${bottomPanelHeight}px` }}>
      <Toolbar />
      <section className="workspace" style={{ gridTemplateColumns: `${leftPanelWidth}px 6px minmax(540px, 1fr) 6px ${propertyPanelWidth}px` }}>
        <LeftPanel />
        <div
          className="resizeHandle verticalResizeHandle"
          role="separator"
          aria-orientation="vertical"
          aria-label="调整左侧面板宽度"
          title="调整左侧面板宽度"
          onPointerDown={(event) => startPanelResize('left', event)}
        />
        <RadarCanvas />
        <div
          className="resizeHandle verticalResizeHandle"
          role="separator"
          aria-orientation="vertical"
          aria-label="调整属性面板宽度"
          title="调整属性面板宽度"
          onPointerDown={(event) => startPanelResize('property', event)}
        />
        <PropertiesPanel />
      </section>
      <section className="bottomPanelFrame">
        <div
          className="resizeHandle horizontalResizeHandle"
          role="separator"
          aria-orientation="horizontal"
          aria-label="调整底部面板高度"
          title="调整底部面板高度"
          onPointerDown={(event) => startPanelResize('bottom', event)}
        />
        <BottomPanel />
      </section>
    </main>
  )
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, Math.round(value)))
}

function isEditingText(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  const tagName = target.tagName.toLowerCase()
  return tagName === 'input' || tagName === 'textarea' || tagName === 'select' || target.isContentEditable
}

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useEditorStore } from '../src/store/editorStore'

describe('editor workflow', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    useEditorStore.getState().newProject()
  })

  it('covers new project, node editing, issue location clearing, and Luban write', async () => {
    expect(useEditorStore.getState().project.map_id).toBe('de_dust2')
    expect(useEditorStore.getState().selected).toBeNull()

    useEditorStore.getState().updateNode('A_SITE', { shape: 'Polygon', points: [] })
    const issue = useEditorStore.getState().issues.find((item) => item.object?.kind === 'node' && item.object.id === 'A_SITE' && item.field === 'points')
    expect(issue?.message).toContain('多边形顶点少于 3 个')

    useEditorStore.getState().locate(issue!.object)
    expect(useEditorStore.getState().located).toEqual({ kind: 'node', id: 'A_SITE' })

    useEditorStore.getState().updateNode('A_SITE', { name: 'A Site Edited', shape: 'Circle', radius: 0.05, points: [] })
    expect(useEditorStore.getState().located).toBeNull()
    expect(useEditorStore.getState().project.nodes.find((node) => node.id === 'A_SITE')?.name).toBe('A Site Edited')

    const fetchMock = vi.fn(async (_input: unknown, init?: RequestInit) => {
      const body = JSON.parse(String(init?.body))
      expect(body.project.nodes.find((node: { id: string; name: string }) => node.id === 'A_SITE')?.name).toBe('A Site Edited')
      return {
        ok: true,
        json: async () => ({
          ok: true,
          result: { entries: [{ file: '#MapNode.xlsx', table: 'tb_map_node', rows: 8, backup: null, warnings: [] }] },
          issues: [],
        }),
      }
    })
    vi.stubGlobal('fetch', fetchMock)

    await useEditorStore.getState().write()

    expect(fetchMock).toHaveBeenCalledWith('/api/luban/write', expect.objectContaining({ method: 'POST' }))
    expect(useEditorStore.getState().writeLog).toHaveLength(1)
    expect(useEditorStore.getState().serviceMessage).toBe('Luban Excel 与工程快照写入完成')
  })
})

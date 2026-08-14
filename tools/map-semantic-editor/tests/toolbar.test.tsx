import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import { Toolbar } from '../src/components/Toolbar'
import { useEditorStore } from '../src/store/editorStore'

describe('toolbar dangerous actions', () => {
  beforeEach(() => {
    useEditorStore.getState().newProject()
  })

  it('requires confirmation before creating a new project', () => {
    useEditorStore.getState().updateNode('A_SITE', { name: 'Dirty A Site' })

    render(<Toolbar />)

    fireEvent.click(screen.getByRole('button', { name: /^新建$/ }))

    expect(screen.getByRole('dialog', { name: '新建项目' })).toBeInTheDocument()
    expect(screen.getByText('即将新建项目，当前的修改将全部丢失，是否继续？')).toBeInTheDocument()
    expect(useEditorStore.getState().project.nodes.find((node) => node.id === 'A_SITE')?.name).toBe('Dirty A Site')

    fireEvent.click(screen.getByRole('button', { name: '取消' }))

    expect(screen.queryByRole('dialog', { name: '新建项目' })).not.toBeInTheDocument()
    expect(useEditorStore.getState().project.nodes.find((node) => node.id === 'A_SITE')?.name).toBe('Dirty A Site')

    fireEvent.click(screen.getByRole('button', { name: /^新建$/ }))
    fireEvent.click(screen.getByRole('button', { name: '继续新建' }))

    expect(screen.queryByRole('dialog', { name: '新建项目' })).not.toBeInTheDocument()
    expect(useEditorStore.getState().project.nodes.find((node) => node.id === 'A_SITE')?.name).toBe('A Site')
  })

  it('offers one-click local frontend and backend refresh', () => {
    render(<Toolbar />)

    expect(screen.getByRole('button', { name: '更新本地前后端' })).toBeEnabled()
  })
})

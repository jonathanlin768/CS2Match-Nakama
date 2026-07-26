import path from 'node:path'
import fs from 'node:fs/promises'
import { afterEach, describe, expect, it } from 'vitest'
import { dataFilePath, dataProjectFilePath, dataRoot, projectFilePath, projectFileRoot, resolveInside } from '../server/paths'
import { saveProject } from '../server/projectFiles'
import { createSampleProject } from '../src/lib/sampleProject'

const testProjectName = 'service_test_project.json'

afterEach(async () => {
  await fs.rm(projectFilePath(testProjectName), { force: true })
  const backupFiles = await findFiles(path.join(projectFileRoot, '.bak'), testProjectName)
  await Promise.all(backupFiles.map((file) => fs.rm(file, { force: true })))
})

describe('local service path guard', () => {
  it('allows project files under the editor data directory', () => {
    const file = projectFilePath('de_dust2.json')
    expect(file.startsWith(projectFileRoot)).toBe(true)
  })

  it('rejects project file traversal', () => {
    expect(() => projectFilePath('../outside.json')).toThrow()
    expect(() => projectFilePath('bad/name.json')).toThrow()
  })

  it('allows Luban auto-import xlsx files under configs/Datas', () => {
    const file = dataFilePath('#MapNode.xlsx')
    expect(file).toBe(path.join(dataRoot, '#MapNode.xlsx'))
  })

  it('allows published project snapshots under configs/Datas', () => {
    const file = dataProjectFilePath('de_dust2.json')
    expect(file).toBe(path.join(dataRoot, 'de_dust2.json'))
  })

  it('rejects paths outside the allowed root', () => {
    expect(() => resolveInside(dataRoot, '..', 'outside.xlsx')).toThrow()
    expect(() => dataFilePath('..\\outside.xlsx')).toThrow()
    expect(() => dataProjectFilePath('..\\outside.json')).toThrow()
  })

  it('backs up an existing local project before saving over it', async () => {
    const first = createSampleProject()
    const second = createSampleProject()
    second.nodes.push({
      id: 'BACKUP_TEST_NODE',
      map_id: second.map_id,
      name: 'Backup Test Node',
      zone: 'Test',
      site: 'None',
      node_type: 'Lane',
      default_side: 'None',
      x: 0.5,
      y: 0.5,
      floor: 'Ground',
      area_usages: [],
      shape: 'None',
      radius: null,
      points: [],
    })

    await saveProject(testProjectName, first)
    await saveProject(testProjectName, second)

    const backupRoot = path.join(projectFileRoot, '.bak')
    const backupFiles = await findFiles(backupRoot, testProjectName)
    expect(backupFiles.length).toBeGreaterThan(0)
  })
})

async function findFiles(root: string, fileName: string): Promise<string[]> {
  try {
    const entries = await fs.readdir(root, { withFileTypes: true })
    const nested = await Promise.all(entries.map(async (entry) => {
      const fullPath = path.join(root, entry.name)
      if (entry.isDirectory()) return findFiles(fullPath, fileName)
      return entry.name === fileName ? [fullPath] : []
    }))
    return nested.flat()
  } catch {
    return []
  }
}

import fs from 'node:fs/promises'
import path from 'node:path'
import { parseProject, type MapProject } from '../src/lib/model'
import { createSampleProject } from '../src/lib/sampleProject'
import { dataProjectFilePath, dataRoot, projectFilePath, projectFileRoot, resolveInside } from './paths'

export async function ensureProjectRoot(): Promise<void> {
  await fs.mkdir(projectFileRoot, { recursive: true })
}

export async function readProject(name = 'de_dust2.json'): Promise<MapProject> {
  await ensureProjectRoot()
  const file = projectFilePath(name)
  try {
    const raw = await fs.readFile(file, 'utf8')
    return parseProject(JSON.parse(raw))
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
      const project = createSampleProject()
      await saveProject(name, project)
      return project
    }
    throw error
  }
}

export async function saveProject(name: string, project: MapProject): Promise<{ file: string; project: MapProject }> {
  await ensureProjectRoot()
  const parsed = parseProject(project)
  const file = projectFilePath(name)
  await backupExistingProject(file)
  await fs.writeFile(file, `${JSON.stringify(parsed, null, 2)}\n`, 'utf8')
  return { file, project: parsed }
}

export async function readPublishedProject(name = 'de_dust2.json'): Promise<{ file: string; project: MapProject }> {
  const file = dataProjectFilePath(name)
  const raw = await fs.readFile(file, 'utf8')
  return { file, project: parseProject(JSON.parse(raw)) }
}

export async function savePublishedProject(project: MapProject): Promise<{ tableName: string; file: string; rows: number; backup?: string; warnings: string[] }> {
  await fs.mkdir(dataRoot, { recursive: true })
  const parsed = parseProject(project)
  const target = dataProjectFilePath(`${parsed.map_id}.json`)
  const backup = await backupExistingPublishedProject(target)
  const temp = `${target}.tmp-${process.pid}-${Date.now()}`
  await fs.writeFile(temp, `${JSON.stringify(parsed, null, 2)}\n`, 'utf8')
  await fs.rename(temp, target)
  return {
    tableName: 'map_project_snapshot',
    file: path.relative(dataRoot, target),
    rows: 1,
    backup: backup ? path.relative(dataRoot, backup) : undefined,
    warnings: ['编辑器工程快照，不参与 Luban 导表'],
  }
}

async function backupExistingPublishedProject(target: string): Promise<string | undefined> {
  try {
    await fs.access(target)
  } catch {
    return undefined
  }

  const backupDir = resolveInside(dataRoot, '.bak', 'map-semantic-editor', timestamp())
  await fs.mkdir(backupDir, { recursive: true })
  const backup = resolveInside(backupDir, path.basename(target))
  await fs.copyFile(target, backup)
  return backup
}

async function backupExistingProject(target: string): Promise<string | undefined> {
  try {
    await fs.access(target)
  } catch {
    return undefined
  }

  const backupDir = resolveInside(projectFileRoot, '.bak', timestamp())
  await fs.mkdir(backupDir, { recursive: true })
  const backup = resolveInside(backupDir, path.basename(target))
  await fs.copyFile(target, backup)
  return backup
}

function timestamp(): string {
  const now = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}-${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`
}

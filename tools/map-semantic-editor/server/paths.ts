import path from 'node:path'
import { fileURLToPath } from 'node:url'

const serverDir = path.dirname(fileURLToPath(import.meta.url))

export const editorRoot = path.resolve(serverDir, '..')
export const projectRoot = path.resolve(editorRoot, '..', '..')
export const dataRoot = path.resolve(projectRoot, 'configs', 'Datas')
export const projectFileRoot = path.resolve(editorRoot, 'data')
export const radarRoot = path.resolve(projectRoot, 'client', 'public', 'csmaps')
export const portraitRoot = path.resolve(projectRoot, 'client', 'public', 'portraits')
export const playerCardRoot = path.resolve(projectRoot, 'client', 'public', 'player-cards')
export const teamLogoRoot = path.resolve(projectRoot, 'client', 'public', 'teams')

export function resolveInside(root: string, ...segments: string[]): string {
  const target = path.resolve(root, ...segments)
  const relative = path.relative(root, target)
  if (relative.startsWith('..') || path.isAbsolute(relative)) {
    throw new Error(`Path escapes allowed root: ${target}`)
  }
  return target
}

export function projectFilePath(name: string): string {
  if (!/^[a-zA-Z0-9_-]+\.json$/.test(name)) {
    throw new Error('Project file name must match [a-zA-Z0-9_-]+.json')
  }
  return resolveInside(projectFileRoot, name)
}

export function dataFilePath(fileName: string, root = dataRoot): string {
  if (!/^#?[A-Za-z0-9_]+\.xlsx$/.test(fileName) && !/^__[A-Za-z0-9_]+__\.xlsx$/.test(fileName)) {
    throw new Error('Luban file name must be an expected .xlsx file name')
  }
  return resolveInside(root, fileName)
}

export function dataProjectFilePath(name: string): string {
  if (!/^[a-zA-Z0-9_-]+\.json$/.test(name)) {
    throw new Error('Published project file name must match [a-zA-Z0-9_-]+.json')
  }
  return resolveInside(dataRoot, name)
}

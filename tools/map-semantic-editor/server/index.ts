import fs from 'node:fs/promises'
import http from 'node:http'
import path from 'node:path'
import { URL } from 'node:url'
import { toExportTables } from '../src/lib/exportTables'
import { parseProject } from '../src/lib/model'
import { hasBlockingIssues, validateProject } from '../src/lib/validation'
import { validateDocuments } from '../src/lib/lubanValidation'
import { readLubanSummary, writeExportTables } from './excel'
import { configAssetExists, configAssetFilePath, findReferences, hydrateIncomingDocument, listLubanTables, readEditableDocuments, readLubanTable, saveConfigImage, saveLubanTable, saveLubanTables, syncRecordId } from './configTables'
import { importLubanProject } from './importTables'
import { isGenConfigRunning, runGenConfig, runLocalUpdate } from './genConfig'
import { readProject, readPublishedProject, saveProject, savePublishedProject } from './projectFiles'
import { radarRoot, resolveInside } from './paths'

const port = Number(process.env.MAP_EDITOR_SERVICE_PORT ?? 5178)
const host = '127.0.0.1'

const server = http.createServer(async (request, response) => {
  try {
    await route(request, response)
  } catch (error) {
    json(response, 500, { ok: false, error: error instanceof Error ? error.message : String(error) })
  }
})

server.listen(port, host, () => {
  console.log(`Map semantic editor service: http://${host}:${port}`)
})

async function route(request: http.IncomingMessage, response: http.ServerResponse): Promise<void> {
  if (!request.url) {
    json(response, 404, { ok: false, error: 'Missing URL' })
    return
  }
  const url = new URL(request.url, `http://${host}:${port}`)

  if (request.method === 'GET' && url.pathname === '/api/health') {
    json(response, 200, { ok: true, service: 'map-semantic-editor', genConfigRunning: isGenConfigRunning() })
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/project') {
    const project = await readProject(url.searchParams.get('name') ?? 'de_dust2.json')
    json(response, 200, { ok: true, project })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/project/save') {
    const body = await readJson(request)
    const project = parseProject(body.project)
    const result = await saveProject(String(body.name ?? 'de_dust2.json'), project)
    json(response, 200, { ok: true, file: result.file, project: result.project })
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/luban/summary') {
    json(response, 200, { ok: true, tables: await readLubanSummary() })
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/config/tables') {
    json(response, 200, { ok: true, tables: await listLubanTables() })
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/config/table') {
    const fileName = url.searchParams.get('file')
    if (!fileName) throw new Error('缺少 table file 参数')
    json(response, 200, { ok: true, document: await readLubanTable(fileName) })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/config/table/save') {
    const body = await readJson(request)
    const document = await hydrateIncomingDocument(body.document)
    const documents = await documentsWithChanges([document])
    const issues = validateDocuments(documents)
    if (issues.some((issue) => issue.severity === 'ERROR')) {
      json(response, 422, { ok: false, error: '校验未通过，已阻止保存', issues })
      return
    }
    const result = await saveLubanTable(document)
    json(response, 200, { ok: true, result, issues })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/config/tables/save') {
    const body = await readJson(request)
    if (!Array.isArray(body.documents)) throw new Error('documents 必须是数组')
    const changed = await Promise.all(body.documents.map((document) => hydrateIncomingDocument(document)))
    const documents = await documentsWithChanges(changed)
    const issues = validateDocuments(documents)
    if (issues.some((issue) => issue.severity === 'ERROR')) {
      json(response, 422, { ok: false, error: '校验未通过，已阻止批量保存', issues })
      return
    }
    const results = await saveLubanTables(changed)
    json(response, 200, { ok: true, results, issues })
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/config/references') {
    const table = url.searchParams.get('table') ?? ''
    const id = url.searchParams.get('id') ?? ''
    if (!table || !id) throw new Error('缺少 table 或 id 参数')
    json(response, 200, { ok: true, references: await findReferences(table, id) })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/config/id-sync') {
    const body = await readJson(request)
    const result = await syncRecordId(String(body.fileName ?? ''), String(body.oldId ?? ''), String(body.newId ?? ''))
    json(response, 200, { ok: true, ...result })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/config/image') {
    const body = await readJson(request, 12 * 1024 * 1024)
    const kind = body.kind === 'portrait' ? 'portrait' : body.kind === 'team' ? 'team' : body.kind === 'player-card' ? 'player-card' : null
    if (!kind) throw new Error('图片类型必须是 portrait、team 或 player-card')
    const result = await saveConfigImage(kind, String(body.fileName ?? ''), String(body.dataBase64 ?? ''), body.overwrite === true)
    json(response, 200, { ok: true, result })
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/config/asset') {
    const assetPath = url.searchParams.get('path') ?? ''
    json(response, 200, { ok: true, exists: await configAssetExists(assetPath) })
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/config/asset-file') {
    await serveConfigAsset(url.searchParams.get('path') ?? '', response)
    return
  }

  if (request.method === 'GET' && url.pathname === '/api/luban/project') {
    const result = await readPublishedProject(url.searchParams.get('name') ?? 'de_dust2.json')
    json(response, 200, { ok: true, file: result.file, project: result.project })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/luban/import') {
    const body = await readJson(request)
    const project = parseProject(body.project)
    const result = await importLubanProject(project)
    json(response, 200, { ok: true, project: result.project, summary: result.summary, warnings: result.warnings })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/luban/write') {
    const body = await readJson(request)
    const project = parseProject(body.project)
    const issues = validateProject(project)
    if (hasBlockingIssues(issues)) {
      json(response, 422, { ok: false, issues })
      return
    }
    const result = await writeExportTables(toExportTables(project))
    result.entries.push(await savePublishedProject(project))
    json(response, 200, { ok: true, result, issues })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/gen-config') {
    const result = await runGenConfig()
    json(response, 200, { ok: result.status === 'success', result })
    return
  }

  if (request.method === 'POST' && url.pathname === '/api/update-local') {
    const result = await runLocalUpdate()
    json(response, 200, { ok: result.status === 'success', result })
    return
  }

  if (request.method === 'GET' && url.pathname.startsWith('/csmaps/')) {
    await serveRadar(url.pathname.replace('/csmaps/', ''), response)
    return
  }

  json(response, 404, { ok: false, error: 'Not found' })
}

async function serveRadar(fileName: string, response: http.ServerResponse): Promise<void> {
  const file = resolveInside(radarRoot, fileName)
  const data = await fs.readFile(file)
  const type = path.extname(file).toLowerCase() === '.webp' ? 'image/webp' : 'application/octet-stream'
  response.writeHead(200, { 'Content-Type': type, 'Cache-Control': 'no-cache' })
  response.end(data)
}

function json(response: http.ServerResponse, status: number, value: unknown): void {
  response.writeHead(status, {
    'Content-Type': 'application/json; charset=utf-8',
    'Access-Control-Allow-Origin': 'http://127.0.0.1:5177',
    'Access-Control-Allow-Headers': 'content-type',
    'Access-Control-Allow-Methods': 'GET,POST,OPTIONS',
  })
  response.end(JSON.stringify(value))
}

async function readJson(request: http.IncomingMessage, maxBytes = 2 * 1024 * 1024): Promise<Record<string, unknown>> {
  const chunks: Buffer[] = []
  let size = 0
  for await (const chunk of request) {
    const buffer = Buffer.from(chunk)
    size += buffer.length
    if (size > maxBytes) throw new Error('请求内容过大')
    chunks.push(buffer)
  }
  if (chunks.length === 0) return {}
  return JSON.parse(Buffer.concat(chunks).toString('utf8')) as Record<string, unknown>
}

async function serveConfigAsset(relativePath: string, response: http.ServerResponse): Promise<void> {
  const file = configAssetFilePath(relativePath)
  if (!file) {
    json(response, 404, { ok: false, error: '不支持的图片资源路径' })
    return
  }

  const data = await fs.readFile(file)
  const extension = path.extname(file).toLowerCase()
  const type = extension === '.png' ? 'image/png' : extension === '.webp' ? 'image/webp' : 'image/jpeg'
  response.writeHead(200, { 'Content-Type': type, 'Cache-Control': 'no-cache' })
  response.end(data)
}

async function documentsWithChanges(changed: Awaited<ReturnType<typeof hydrateIncomingDocument>>[]): Promise<Awaited<ReturnType<typeof readEditableDocuments>>> {
  const documents = await readEditableDocuments()
  const byFile = new Map(changed.map((document) => [document.fileName, document]))
  return documents.map((document) => byFile.get(document.fileName) ?? document)
}

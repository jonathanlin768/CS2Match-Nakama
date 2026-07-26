import fs from 'node:fs/promises'
import http from 'node:http'
import path from 'node:path'
import { URL } from 'node:url'
import { toExportTables } from '../src/lib/exportTables'
import { parseProject } from '../src/lib/model'
import { hasBlockingIssues, validateProject } from '../src/lib/validation'
import { readLubanSummary, writeExportTables } from './excel'
import { isGenConfigRunning, runGenConfig } from './genConfig'
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

  if (request.method === 'GET' && url.pathname === '/api/luban/project') {
    const result = await readPublishedProject(url.searchParams.get('name') ?? 'de_dust2.json')
    json(response, 200, { ok: true, file: result.file, project: result.project })
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

async function readJson(request: http.IncomingMessage): Promise<Record<string, unknown>> {
  const chunks: Buffer[] = []
  for await (const chunk of request) chunks.push(Buffer.from(chunk))
  if (chunks.length === 0) return {}
  return JSON.parse(Buffer.concat(chunks).toString('utf8')) as Record<string, unknown>
}

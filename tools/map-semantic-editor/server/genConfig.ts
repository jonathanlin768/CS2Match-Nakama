import { spawn } from 'node:child_process'
import { performance } from 'node:perf_hooks'
import { projectRoot } from './paths'

export interface GenConfigResult {
  operation: 'gen-config' | 'update-local'
  status: 'success' | 'failed' | 'timeout' | 'running'
  exitCode: number | null
  durationMs: number
  stdout: string
  stderr: string
}

let running: Promise<GenConfigResult> | null = null

export function isGenConfigRunning(): boolean {
  return running !== null
}

export function runGenConfig(timeoutMs = 5 * 60 * 1000): Promise<GenConfigResult> {
  return runScript('gen-config', 'scripts/gen-config.ps1', timeoutMs)
}

export function runLocalUpdate(timeoutMs = 30 * 60 * 1000): Promise<GenConfigResult> {
  return runScript('update-local', 'scripts/update-local-config.ps1', timeoutMs)
}

function runScript(operation: GenConfigResult['operation'], script: string, timeoutMs: number): Promise<GenConfigResult> {
  if (running) return running

  running = execute(operation, script, timeoutMs).finally(() => {
    running = null
  })

  return running
}

function execute(operation: GenConfigResult['operation'], script: string, timeoutMs: number): Promise<GenConfigResult> {
  const started = performance.now()
  const child = spawn('powershell.exe', ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', script], {
    cwd: projectRoot,
    windowsHide: true,
  })

  let stdout = ''
  let stderr = ''
  let didTimeout = false
  let settled = false

  return new Promise((resolve) => {
    function finish(result: GenConfigResult): void {
      if (settled) return
      settled = true
      clearTimeout(timer)
      resolve(result)
    }

    const timer = setTimeout(() => {
      didTimeout = true
      killProcessTree(child.pid, (message) => {
        if (message) stderr += message
        finish({
          operation,
          status: 'timeout',
          exitCode: null,
          durationMs: Math.round(performance.now() - started),
          stdout,
          stderr,
        })
      })
    }, timeoutMs)

    child.stdout.on('data', (chunk: Buffer) => {
      stdout += chunk.toString('utf8')
    })
    child.stderr.on('data', (chunk: Buffer) => {
      stderr += chunk.toString('utf8')
    })

    child.on('error', (error) => {
      finish({
        operation,
        status: 'failed',
        exitCode: null,
        durationMs: Math.round(performance.now() - started),
        stdout,
        stderr: `${stderr}${error.message}`,
      })
    })
    child.on('close', (code) => {
      finish({
        operation,
        status: didTimeout ? 'timeout' : code === 0 ? 'success' : 'failed',
        exitCode: code,
        durationMs: Math.round(performance.now() - started),
        stdout,
        stderr,
      })
    })
  })
}

function killProcessTree(pid: number | undefined, onDone: (message: string) => void): void {
  if (!pid) {
    onDone('导表超时，且无法定位子进程 PID。\n')
    return
  }

  if (process.platform !== 'win32') {
    try {
      process.kill(pid, 'SIGKILL')
      onDone('')
    } catch (error) {
      onDone(`导表超时，终止进程失败: ${error instanceof Error ? error.message : String(error)}\n`)
    }
    return
  }

  const killer = spawn('taskkill.exe', ['/PID', String(pid), '/T', '/F'], { windowsHide: true })
  let output = ''
  killer.stdout.on('data', (chunk: Buffer) => {
    output += chunk.toString('utf8')
  })
  killer.stderr.on('data', (chunk: Buffer) => {
    output += chunk.toString('utf8')
  })
  killer.on('error', (error) => {
    onDone(`导表超时，终止进程树失败: ${error.message}\n`)
  })
  killer.on('close', (code) => {
    onDone(code === 0 ? '' : `导表超时，终止进程树返回 ${code}: ${output}\n`)
  })
}

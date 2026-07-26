import type { MapNode, Point } from './model'

export function sampleNode(node: MapNode, count = 20, seed = 1337): Point[] {
  const random = mulberry32(hash(`${node.id}:${seed}`))
  const samples: Point[] = []

  for (let index = 0; index < count; index += 1) {
    if (node.shape === 'Circle' && node.radius && node.radius > 0) {
      const angle = random() * Math.PI * 2
      const distance = Math.sqrt(random()) * node.radius
      samples.push(clampPoint({ x: node.x + Math.cos(angle) * distance, y: node.y + Math.sin(angle) * distance }))
      continue
    }

    if (node.shape === 'Polygon' && node.points.length >= 3) {
      samples.push(samplePolygon(node.points, random))
      continue
    }

    const jitter = 0.008
    samples.push(clampPoint({ x: node.x + (random() - 0.5) * jitter, y: node.y + (random() - 0.5) * jitter }))
  }

  return samples
}

export function pointInPolygon(point: Point, polygon: Point[]): boolean {
  let inside = false
  for (let i = 0, j = polygon.length - 1; i < polygon.length; j = i, i += 1) {
    const pi = polygon[i]
    const pj = polygon[j]
    const intersects = pi.y > point.y !== pj.y > point.y && point.x < ((pj.x - pi.x) * (point.y - pi.y)) / (pj.y - pi.y) + pi.x
    if (intersects) inside = !inside
  }
  return inside
}

function samplePolygon(points: Point[], random: () => number): Point {
  const minX = Math.min(...points.map((point) => point.x))
  const maxX = Math.max(...points.map((point) => point.x))
  const minY = Math.min(...points.map((point) => point.y))
  const maxY = Math.max(...points.map((point) => point.y))

  for (let attempts = 0; attempts < 200; attempts += 1) {
    const candidate = {
      x: minX + random() * (maxX - minX),
      y: minY + random() * (maxY - minY),
    }
    if (pointInPolygon(candidate, points)) return candidate
  }

  const center = points.reduce((acc, point) => ({ x: acc.x + point.x, y: acc.y + point.y }), { x: 0, y: 0 })
  return clampPoint({ x: center.x / points.length, y: center.y / points.length })
}

function clampPoint(point: Point): Point {
  return {
    x: Math.min(1, Math.max(0, point.x)),
    y: Math.min(1, Math.max(0, point.y)),
  }
}

function hash(value: string): number {
  let result = 2166136261
  for (let index = 0; index < value.length; index += 1) {
    result ^= value.charCodeAt(index)
    result = Math.imul(result, 16777619)
  }
  return result >>> 0
}

function mulberry32(seed: number): () => number {
  return () => {
    let t = (seed += 0x6d2b79f5)
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}


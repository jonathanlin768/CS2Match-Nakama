import { useEffect, useMemo, useRef, useState } from 'react'
import { Circle, Group, Image as KonvaImage, Label, Layer, Line, Rect, Stage, Tag, Text } from 'react-konva'
import type { KonvaEventObject } from 'konva/lib/Node'
import type { MapNode, Point } from '../lib/model'
import { nextLineSelection, type LineSelection } from '../lib/selection'
import { useEditorStore } from '../store/editorStore'

const stagePadding = 18
const lineHitThreshold = 10
const riskMarkerColor = '#ff3d71'
const riskRangeColor = '#ffb020'

export function RadarCanvas() {
  const containerRef = useRef<HTMLDivElement>(null)
  const [size, setSize] = useState({ width: 820, height: 620 })
  const [image, setImage] = useState<HTMLImageElement | null>(null)
  const project = useEditorStore((state) => state.project)
  const tool = useEditorStore((state) => state.tool)
  const selected = useEditorStore((state) => state.selected)
  const located = useEditorStore((state) => state.located)
  const pendingNodeId = useEditorStore((state) => state.pendingNodeId)
  const routeDraft = useEditorStore((state) => state.routeDraft)
  const rangeDraft = useEditorStore((state) => state.rangeDraft)
  const samples = useEditorStore((state) => state.samples)
  const addNode = useEditorStore((state) => state.addNode)
  const moveNode = useEditorStore((state) => state.moveNode)
  const select = useEditorStore((state) => state.select)
  const handleNodeToolClick = useEditorStore((state) => state.handleNodeToolClick)
  const appendRangePoint = useEditorStore((state) => state.appendRangePoint)

  useEffect(() => {
    const nextImage = new window.Image()
    nextImage.src = project.radar_image
    nextImage.onload = () => setImage(nextImage)
  }, [project.radar_image])

  useEffect(() => {
    if (!containerRef.current) return
    const observer = new ResizeObserver(([entry]) => {
      setSize({ width: Math.max(520, entry.contentRect.width), height: Math.max(420, entry.contentRect.height) })
    })
    observer.observe(containerRef.current)
    return () => observer.disconnect()
  }, [])

  const frame = useMemo(() => {
    const length = Math.min(size.width - stagePadding * 2, size.height - stagePadding * 2)
    return {
      x: (size.width - length) / 2,
      y: (size.height - length) / 2,
      length,
    }
  }, [size])

  const pointToStage = (point: Point) => ({
    x: frame.x + point.x * frame.length,
    y: frame.y + point.y * frame.length,
  })
  const stageToPoint = (point: Point) => ({
    x: (point.x - frame.x) / frame.length,
    y: (point.y - frame.y) / frame.length,
  })
  const nodeById = (id: string) => project.nodes.find((node) => node.id === id)

  function handleStageClick(event: KonvaEventObject<MouseEvent>) {
    const pointer = event.target.getStage()?.getPointerPosition()
    if (!pointer) return
    if (!isInsideFrame(pointer, frame)) return
    const normalized = stageToPoint(pointer)

    if (rangeDraft) {
      appendRangePoint(normalized)
      return
    }

    if (tool === 'select') {
      const candidates = findLineCandidates(pointer)
      const next = nextLineSelection(candidates, selected)
      if (next) select(next)
      return
    }

    if (tool !== 'node' && tool !== 'risk') return
    if (tool === 'node') addNode(normalized)
    if (tool === 'risk') addNode(normalized, true)
  }

  function findLineCandidates(pointer: Point) {
    const candidates: LineSelection[] = []
    const seen = new Set<string>()
    const addCandidate = (candidate: LineSelection) => {
      const key = `${candidate.kind}:${candidate.id}`
      if (!seen.has(key)) {
        candidates.push(candidate)
        seen.add(key)
      }
    }

    if (project.layers.routes?.visible && !project.layers.routes.locked) {
      for (const route of project.routes) {
        for (let index = 0; index < route.nodes.length - 1; index += 1) {
          const from = nodeById(route.nodes[index])
          const to = nodeById(route.nodes[index + 1])
          if (!from || !to) continue
          if (distanceToSegment(pointer, pointToStage(from), pointToStage(to)) <= lineHitThreshold) {
            addCandidate({ kind: 'route', id: route.id })
            break
          }
        }
      }
    }

    if (project.layers.edges?.visible && !project.layers.edges.locked) {
      for (const edge of project.edges) {
        const from = nodeById(edge.from)
        const to = nodeById(edge.to)
        if (!from || !to) continue
        if (distanceToSegment(pointer, pointToStage(from), pointToStage(to)) <= lineHitThreshold) {
          addCandidate({ kind: 'edge', id: edge.id })
        }
      }
    }

    return candidates
  }

  function renderNodeMarker(node: MapNode, layer: { locked: boolean; color: string }, radius: number) {
    const point = pointToStage(node)
    const isSelected = selected?.kind === 'node' && selected.id === node.id
    const isLocated = located?.kind === 'node' && located.id === node.id
    const isPending = pendingNodeId === node.id
    const isRisk = isRiskHotspot(node)
    const fill = isRisk ? riskMarkerColor : layer.color
    const stroke = isSelected || isPending ? '#ffffff' : isRisk ? '#ffe08a' : '#111820'
    const markerRadius = isRisk ? radius + 2 : radius
    return (
      <Group
        key={node.id}
        x={point.x}
        y={point.y}
        draggable={!layer.locked && !rangeDraft}
        onClick={(event) => {
          event.cancelBubble = true
          if (rangeDraft) {
            appendRangePoint(node)
            return
          }
          handleNodeToolClick(node.id)
        }}
        onDragEnd={(event) => {
          moveNode(node.id, stageToPoint({ x: event.target.x(), y: event.target.y() }))
        }}
      >
        <Circle radius={markerRadius} fill={fill} stroke={stroke} strokeWidth={isSelected || isPending ? 3 : isRisk ? 2.4 : 1.5} />
        {isLocated ? <Circle radius={markerRadius + 8} stroke="#f8e16c" strokeWidth={3} opacity={0.95} /> : null}
        {isRisk ? <Circle radius={markerRadius + 4} stroke="#ffb020" strokeWidth={1.5} dash={[3, 3]} opacity={0.9} /> : null}
        <Label x={10} y={-21} opacity={isSelected ? 1 : 0.82}>
          <Tag fill="#101418" stroke="#3a444f" cornerRadius={3} />
          <Text text={node.id} fill="#e8edf2" fontSize={11} padding={4} />
        </Label>
      </Group>
    )
  }

  return (
    <section className="canvasPanel" ref={containerRef}>
      <Stage width={size.width} height={size.height} onClick={handleStageClick}>
        <Layer>
          <Rect x={0} y={0} width={size.width} height={size.height} fill="#121417" />
          <Rect x={frame.x} y={frame.y} width={frame.length} height={frame.length} fill="#1d2228" stroke="#303943" strokeWidth={1} />
          {image ? <KonvaImage image={image} x={frame.x} y={frame.y} width={frame.length} height={frame.length} opacity={0.86} /> : null}
          {gridLines(frame).map((line) => (
            <Line key={line.key} points={line.points} stroke="#ffffff" strokeWidth={0.4} opacity={0.12} />
          ))}
        </Layer>

        <Layer visible={project.layers.edges?.visible}>
          {project.edges.map((edge) => {
            const from = nodeById(edge.from)
            const to = nodeById(edge.to)
            if (!from || !to) return null
            const a = pointToStage(from)
            const b = pointToStage(to)
            const isSelected = selected?.kind === 'edge' && selected.id === edge.id
            const isLocated = located?.kind === 'edge' && located.id === edge.id
            return (
              <Group key={edge.id}>
                {isLocated ? <Line points={[a.x, a.y, b.x, b.y]} stroke="#f8e16c" strokeWidth={8} opacity={0.55} lineCap="round" /> : null}
                <Line points={[a.x, a.y, b.x, b.y]} stroke={project.layers.edges.color} strokeWidth={isSelected ? 4 : 2} opacity={0.9} lineCap="round" />
              </Group>
            )
          })}
        </Layer>

        <Layer visible={project.layers.visibility?.visible}>
          {project.visibility.map((visibility) => {
            const from = nodeById(visibility.from)
            const to = nodeById(visibility.to)
            if (!from || !to) return null
            const a = pointToStage(from)
            const b = pointToStage(to)
            const isLocated = located?.kind === 'visibility' && located.id === visibility.id
            return (
              <Group key={visibility.id}>
                {isLocated ? <Line points={[a.x, a.y, b.x, b.y]} stroke="#f8e16c" strokeWidth={7} opacity={0.5} lineCap="round" /> : null}
                <Line points={[a.x, a.y, b.x, b.y]} stroke={project.layers.visibility.color} strokeWidth={2} dash={[8, 6]} opacity={0.78} />
              </Group>
            )
          })}
        </Layer>

        <Layer visible={project.layers.routes?.visible}>
          {project.routes.map((route) => {
            const points = route.nodes.flatMap((id) => {
              const node = nodeById(id)
              if (!node) return []
              const point = pointToStage(node)
              return [point.x, point.y]
            })
            const isSelected = selected?.kind === 'route' && selected.id === route.id
            const isLocated = located?.kind === 'route' && located.id === route.id
            return points.length >= 4 ? (
              <Group key={route.id}>
                {isLocated ? <Line points={points} stroke="#f8e16c" strokeWidth={10} opacity={0.4} lineCap="round" lineJoin="round" /> : null}
                <Line points={points} stroke={project.layers.routes.color} strokeWidth={isSelected ? 5 : 3} opacity={0.58} lineCap="round" lineJoin="round" />
              </Group>
            ) : null
          })}
          {routeDraft.length > 0 ? <DraftRoute nodes={routeDraft.map(nodeById).filter(Boolean) as MapNode[]} pointToStage={pointToStage} /> : null}
        </Layer>

        <Layer visible={project.layers.ranges?.visible}>
          {project.nodes.map((node) => renderNodeRange(node, pointToStage, frame.length, isRiskHotspot(node) ? riskRangeColor : project.layers.ranges.color, {
            selected: selected?.kind === 'node' && selected.id === node.id,
            selectable: tool === 'select' && !rangeDraft && !project.layers.ranges.locked,
            onSelect: () => select({ kind: 'node', id: node.id }),
          }))}
          {rangeDraft ? <DraftRange draft={rangeDraft} node={nodeById(rangeDraft.nodeId)} pointToStage={pointToStage} frameLength={frame.length} /> : null}
        </Layer>

        <Layer>
          {samples.map((sample, index) => {
            const point = pointToStage(sample)
            return <Circle key={`${sample.x}-${sample.y}-${index}`} x={point.x} y={point.y} radius={3} fill="#f8e16c" stroke="#181818" strokeWidth={1} opacity={0.9} />
          })}
        </Layer>

        <Layer visible={project.layers.nodes?.visible}>
          {project.nodes.map((node) => renderNodeMarker(node, project.layers.nodes, 7))}
        </Layer>
      </Stage>
      <div className="canvasStatus">
        {rangeDraft
          ? rangeDraft.shape === 'Circle'
            ? '划定圆形范围 | 点击边界点写入 radius'
            : `划定多边形范围 | 顶点 ${rangeDraft.points.length} | Enter 完成 | Esc 取消`
          : routeDraft.length > 0
            ? `路线草稿 | 点位 ${routeDraft.length} | Enter 完成 | Esc 取消`
          : `缩放 ${Math.round(project.viewport.zoom * 100)}% | snap ${project.viewport.snap ? 'on' : 'off'} | 工具 ${tool}`}
      </div>
    </section>
  )
}

function distanceToSegment(point: Point, start: Point, end: Point) {
  const dx = end.x - start.x
  const dy = end.y - start.y
  if (dx === 0 && dy === 0) return Math.hypot(point.x - start.x, point.y - start.y)
  const t = Math.max(0, Math.min(1, ((point.x - start.x) * dx + (point.y - start.y) * dy) / (dx * dx + dy * dy)))
  return Math.hypot(point.x - (start.x + t * dx), point.y - (start.y + t * dy))
}

function renderNodeRange(
  node: MapNode,
  pointToStage: (point: Point) => Point,
  frameLength: number,
  color: string,
  interaction: { selected: boolean; selectable: boolean; onSelect: () => void },
) {
  const commonProps = {
    onClick: (event: KonvaEventObject<MouseEvent>) => {
      if (!interaction.selectable) return
      event.cancelBubble = true
      interaction.onSelect()
    },
    listening: interaction.selectable,
  }
  const stroke = interaction.selected ? '#ffffff' : color
  const strokeWidth = interaction.selected ? 3 : 2
  if (node.shape === 'Circle' && node.radius) {
    const center = pointToStage(node)
    return <Circle key={`${node.id}:range`} x={center.x} y={center.y} radius={node.radius * frameLength} fill={color} opacity={0.18} stroke={stroke} strokeWidth={strokeWidth} dash={[5, 4]} {...commonProps} />
  }
  if (node.shape === 'Polygon' && node.points.length >= 3) {
    const points = node.points.flatMap((point) => {
      const next = pointToStage(point)
      return [next.x, next.y]
    })
    return <Line key={`${node.id}:range`} points={points} closed fill={color} opacity={0.18} stroke={stroke} strokeWidth={strokeWidth} dash={[5, 4]} {...commonProps} />
  }
  return null
}

function DraftRange({ draft, node, pointToStage, frameLength }: { draft: { nodeId: string; shape: 'Circle' | 'Polygon'; points: Point[] }; node?: MapNode; pointToStage: (point: Point) => Point; frameLength: number }) {
  if (!node) return null
  const center = pointToStage(node)
  if (draft.shape === 'Circle') {
    return (
      <Group>
        <Circle x={center.x} y={center.y} radius={Math.max(node.radius ?? 0.025, 0.025) * frameLength} stroke="#f8e16c" strokeWidth={2} dash={[7, 5]} opacity={0.8} />
        <Circle x={center.x} y={center.y} radius={4} fill="#f8e16c" stroke="#111820" strokeWidth={1} />
      </Group>
    )
  }
  const points = draft.points.flatMap((point) => {
    const next = pointToStage(point)
    return [next.x, next.y]
  })
  return (
    <Group>
      {points.length >= 4 ? <Line points={points} closed={draft.points.length >= 3} stroke="#f8e16c" strokeWidth={2} dash={[7, 5]} fill="#f8e16c" opacity={draft.points.length >= 3 ? 0.22 : 0.86} lineJoin="round" /> : null}
      {draft.points.map((point, index) => {
        const next = pointToStage(point)
        return <Circle key={`${point.x}-${point.y}-${index}`} x={next.x} y={next.y} radius={4} fill="#f8e16c" stroke="#111820" strokeWidth={1} />
      })}
    </Group>
  )
}

function DraftRoute({ nodes, pointToStage }: { nodes: MapNode[]; pointToStage: (point: Point) => Point }) {
  const points = nodes.flatMap((node) => {
    const point = pointToStage(node)
    return [point.x, point.y]
  })
  return points.length >= 2 ? <Line points={points} stroke="#ffffff" strokeWidth={2} dash={[4, 4]} opacity={0.7} /> : null
}

function gridLines(frame: { x: number; y: number; length: number }) {
  const lines: { key: string; points: number[] }[] = []
  for (let index = 1; index < 10; index += 1) {
    const pos = frame.x + (frame.length * index) / 10
    lines.push({ key: `v-${index}`, points: [pos, frame.y, pos, frame.y + frame.length] })
    lines.push({ key: `h-${index}`, points: [frame.x, pos, frame.x + frame.length, pos] })
  }
  return lines
}

function isInsideFrame(point: Point, frame: { x: number; y: number; length: number }) {
  return point.x >= frame.x && point.x <= frame.x + frame.length && point.y >= frame.y && point.y <= frame.y + frame.length
}

function isRiskHotspot(node: MapNode) {
  return node.area_usages.includes('Risk')
}

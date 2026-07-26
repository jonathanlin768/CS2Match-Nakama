import { describe, expect, it } from 'vitest'
import { nextLineSelection, type LineSelection } from '../src/lib/selection'

describe('line selection cycling', () => {
  const candidates: LineSelection[] = [
    { kind: 'route', id: 'ROUTE_A_LONG' },
    { kind: 'edge', id: 'EDGE_LONG' },
  ]

  it('selects the top-most line candidate first', () => {
    expect(nextLineSelection(candidates, null)).toEqual({ kind: 'route', id: 'ROUTE_A_LONG' })
  })

  it('cycles from route to edge on repeated clicks', () => {
    expect(nextLineSelection(candidates, { kind: 'route', id: 'ROUTE_A_LONG' })).toEqual({ kind: 'edge', id: 'EDGE_LONG' })
  })

  it('cycles back to route after the edge candidate', () => {
    expect(nextLineSelection(candidates, { kind: 'edge', id: 'EDGE_LONG' })).toEqual({ kind: 'route', id: 'ROUTE_A_LONG' })
  })
})

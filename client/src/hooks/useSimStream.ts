import { useState, useCallback } from "react";
import type {
  SimPlayerState,
  SimKillEvent,
  SimC4Event,
  SimTickMsg,
  SimEndMsg,
} from "@/types/sim";

export interface SimStreamState {
  connected: boolean;
  simStarted: boolean;
  tickMsg: SimTickMsg | null;
  killEvents: SimKillEvent[];
  c4Events: SimC4Event[];
  roundEnd: SimEndMsg | null;
  players: SimPlayerState[];
}

// Stub hook — returns empty state, no WebSocket connection
export function useSimStream() {
  const [state] = useState<SimStreamState>({
    connected: false,
    simStarted: false,
    tickMsg: null,
    killEvents: [],
    c4Events: [],
    roundEnd: null,
    players: [],
  });

  // No-op: does not create WebSocket or start simulation
  const startSim = useCallback((_tactic: string, _setup: string, _seed: number) => {
    // stub: do nothing
  }, []);

  const reset = useCallback(() => {
    // stub: do nothing
  }, []);

  return { ...state, startSim, reset };
}

export interface SimPlayerState {
  id: string;
  name: string;
  team: "T" | "CT";
  x: number;
  y: number;
  angle: number;
  alive: boolean;
  hp: number;
  armor: number;
  isC4Carrier?: boolean;
}

export interface SimKillEvent {
  tick: number;
  attackerId: string;
  attackerName: string;
  victimId: string;
  victimName: string;
  weapon: string;
  headshot: boolean;
  deathX: number;
  deathY: number;
}

export interface SimC4Event {
  tick: number;
  type: "plant_started" | "plant_finished" | "defuse_started" | "defuse_finished" | "exploded";
  playerId: string;
  site?: string;
  x?: number;
  y?: number;
}

export interface SimTickMsg {
  tick: number;
  maxTick: number;
  players: SimPlayerState[];
  c4State?: "carried" | "planted" | "exploded";
  c4Pos?: { x: number; y: number };
}

export interface SimStartMsg {
  tTactic: string;
  ctSetup: string;
  players: SimPlayerState[];
  c4CarrierId: string;
}

export interface SimEndMsg {
  tick: number;
  winner: "T" | "CT";
  reason: string;
  stats: SimPlayerStats[];
}

export interface SimPlayerStats {
  id: string;
  kills: number;
  deaths: number;
}

export interface WsMessage {
  msgId: number;
  msgBody: unknown;
}

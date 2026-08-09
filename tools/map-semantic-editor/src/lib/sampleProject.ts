import type { MapProject } from './model'

export function createSampleProject(): MapProject {
  const map_id = 'de_dust2'

  return {
    schema_version: 1,
    map_id,
    name: 'Dust2',
    version: 'v1',
    radar_image: '/csmaps/de_dust2_radar_trans.webp',
    coordinate_normalization: { min_x: 0, max_x: 1, min_y: 0, max_y: 1 },
    layers: {
      nodes: { visible: true, locked: false, color: '#1f9dff' },
      ranges: { visible: true, locked: false, color: '#39c77f' },
      edges: { visible: true, locked: false, color: '#f0a62b' },
      visibility: { visible: true, locked: false, color: '#9a6dff' },
      routes: { visible: true, locked: false, color: '#f24f6a' },
    },
    viewport: { zoom: 1, offset_x: 0, offset_y: 0, snap: true },
    nodes: [
      node(map_id, 'T_SPAWN', 'T Spawn', 0.18, 0.77, 'Spawn', 'T', 'None'),
      node(map_id, 'LONG_DOOR', 'Long Door', 0.37, 0.55, 'Connector', 'Both', 'A'),
      node(map_id, 'A_LONG', 'A Long', 0.55, 0.39, 'Lane', 'Both', 'A'),
      {
        ...node(map_id, 'A_SITE', 'A Site', 0.67, 0.27, 'Site', 'CT', 'A'),
        area_usages: ['KillSample', 'Plant', 'Control'],
        shape: 'Polygon',
        points: [
          { x: 0.61, y: 0.21 },
          { x: 0.72, y: 0.2 },
          { x: 0.76, y: 0.3 },
          { x: 0.66, y: 0.35 },
          { x: 0.59, y: 0.3 },
        ],
      },
      {
        ...node(map_id, 'A_LONG_CROSS', 'A Long Cross', 0.58, 0.34, 'Cover', 'Both', 'A'),
        area_usages: ['Risk'],
        shape: 'Circle',
        radius: 0.035,
      },
      node(map_id, 'CATWALK', 'Catwalk', 0.51, 0.5, 'Lane', 'Both', 'A'),
      node(map_id, 'B_TUNNEL', 'B Tunnel', 0.39, 0.72, 'Lane', 'T', 'B'),
      {
        ...node(map_id, 'B_SITE', 'B Site', 0.65, 0.7, 'Site', 'CT', 'B'),
        area_usages: ['KillSample', 'Plant'],
        shape: 'Circle',
        radius: 0.055,
      },
    ],
    edges: [
      edge('EDGE_T_LONG', 'T_SPAWN', 'LONG_DOOR', ['A_LONG_CROSS']),
      edge('EDGE_LONG_A', 'LONG_DOOR', 'A_LONG', ['A_LONG_CROSS']),
      edge('EDGE_A_SITE', 'A_LONG', 'A_SITE', []),
      edge('EDGE_T_CAT', 'T_SPAWN', 'CATWALK', []),
      edge('EDGE_CAT_A', 'CATWALK', 'A_SITE', []),
      edge('EDGE_T_B', 'T_SPAWN', 'B_TUNNEL', []),
      edge('EDGE_B_SITE', 'B_TUNNEL', 'B_SITE', []),
    ],
    visibility: [
      {
        id: 'VIS_CAT_A',
        from: 'CATWALK',
        to: 'A_SITE',
        visible: true,
        range: 'Mid',
        angle_advantage: 'T',
        elevation: 'HighToLow',
        cover_modifier: -8,
        exposure_modifier: 12,
      },
    ],
    routes: [
      {
        id: 'D2_A_LONG',
        name: 'A Long Execute',
        side: 'T',
        target_site: 'A',
        nodes: ['T_SPAWN', 'LONG_DOOR', 'A_LONG', 'A_SITE'],
        min_players: 2,
        max_players: 5,
        style_tags: ['LongRange', 'SiteEntry'],
      },
    ],
    route_templates: [
      {
        id: 'TPL_A_LONG_DEFAULT',
        map_id,
        side: 'T',
        target_site: 'A',
        tempo: 'Default',
        recommended_min: 2,
        recommended_max: 5,
        required_roles: ['Entry'],
        key_attributes: 'entry=10;aim=6',
        route_ids: ['D2_T_A_LONG'],
        route_allocations: { D2_T_A_LONG: 4 },
        scenario_ids: ['SCN_A_LONG_ENTRY'],
        map_tag_ids: ['D2_A_LONG_RANGE'],
        common_ct_setup_ids: ['CT_2A_1Mid_2B'],
        success_next_phase: 'SiteEntry',
        failure_fallbacks: ['D2_CAT_SPLIT'],
      },
    ],
    scenarios: [
      {
        id: 'SCN_A_LONG_ENTRY',
        route: 'A_Long',
        phase: 'SiteEntry',
        range: 'Long',
        site: 'A',
        tempo: 'SlowDefault',
        posture: 'T_Executing',
        utility_context: 'Even',
        map_tag_ids: ['D2_A_LONG_RANGE'],
        base_time_cost: 12,
        base_weight: 10,
      },
    ],
    map_tags: [
      {
        id: 'D2_A_LONG_RANGE',
        map_id,
        category: 'Range',
        value: 'LongRange',
        side: 'Both',
        weight: 12,
        reason_code: 'long_range_duel',
        description: 'A Long long-range duel pressure.',
      },
    ],
    encounter_modifiers: [
      {
        id: 'MOD_A_LONG_AIM',
        scenario_id: 'SCN_A_LONG_ENTRY',
        factor: 'LongRange',
        side: 'Both',
        attribute: 'aim',
        weight: 8,
        reason_code: 'aim_long_range',
      },
    ],
    combat_consts: [
      combatConst('RoundTimeLimit', 'Time', 'Int', '115', 's', 'Round time limit.'),
      combatConst('BombExplodeTime', 'Bomb', 'Int', '40', 's', 'Bomb explosion countdown.'),
      combatConst('BasePlantTime', 'Bomb', 'Int', '4', 's', 'Base plant time.'),
      combatConst('BaseDefuseTime', 'Bomb', 'Int', '10', 's', 'Base defuse time.'),
      combatConst('MaxEncounterPulses', 'Combat', 'Int', '3', 'count', 'Maximum encounter pulses.'),
      combatConst('CombatScale', 'Combat', 'Float', '12', 'score', 'Combat score scale.'),
    ],
  }
}

function node(
  map_id: string,
  id: string,
  name: string,
  x: number,
  y: number,
  node_type: MapProject['nodes'][number]['node_type'],
  default_side: MapProject['nodes'][number]['default_side'],
  site: MapProject['nodes'][number]['site'],
): MapProject['nodes'][number] {
  return {
    id,
    map_id,
    name,
    zone: site === 'None' ? 'Mid' : `${site} Area`,
    site,
    node_type,
    default_side,
    x,
    y,
    floor: 'Ground',
    area_usages: [],
    shape: 'None',
    radius: null,
    points: [],
  }
}

function edge(id: string, from: string, to: string, risk_points: string[]): MapProject['edges'][number] {
  return {
    id,
    from,
    to,
    base_time: 8,
    stamina_cost: 5,
    risk: risk_points.length ? 22 : 10,
    noise: 12,
    risk_points,
    intercept_nodes: [],
    bidirectional: true,
  }
}

function combatConst(
  key: string,
  category: string,
  value_type: MapProject['combat_consts'][number]['value_type'],
  value: string,
  unit: string,
  description: string,
): MapProject['combat_consts'][number] {
  return {
    key,
    category,
    value_type,
    value,
    min_value: '',
    max_value: '',
    unit,
    description,
  }
}

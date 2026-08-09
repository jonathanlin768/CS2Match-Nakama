# Dust2 causal config inventory

This inventory is the stable-ID and reference plan for tasks 2.1-2.11. Generated JSON is not a source of truth; every row below is authored in `configs/Datas/#*.xlsx` and exported with Luban.

## Baseline gap

| Table | Current | Required first-version coverage |
|---|---:|---|
| RouteTemplate | 1 T placeholder | 6 T strategies, 3 independent CT setups, side/routes/allocations/CT priors |
| Route | 1 T route | T attack/default/rotate routes and CT hold/reinforce/retake routes |
| Scenario | 1 A SiteEntry | OpeningDuel, MidControl, SiteEntry, Retake and BombResolution for A/B/Mid |
| MapTag | 1 A-long tag | range, hold angle, site pressure, timing, rotate risk and post-plant tags |
| EncounterModifier | 1 modifier | ten normalized attribute weights for every scenario plus semantic modifiers |
| MapNode | 10 nodes | all 15 core nodes plus named risk/intercept anchors |
| MapEdge | 7 T-oriented edges | connected bidirectional graph for both spawns, attacks, rotates, pickup and retakes |
| Visibility | 1 relation | key A-long, short, mid and B-site threat lines |
| CombatConst | 39 simulation keys + 1 debug key | 64 simulation keys + existing debug key; 25 required keys missing |

## Stable RouteTemplate IDs

T strategies:

- `A_Long_Rush`
- `A_Short_Split`
- `B_Tunnel_Explode`
- `Mid_To_B`
- `Default_Pick`
- `Fake_A_Go_B`

Independent CT setups:

- `CT_2A_1Mid_2B`
- `CT_3A_1Mid_1B`
- `CT_1A_1Mid_3B`

Every T template references all three setup IDs only through `common_ct_setup_ids`. CT setup selection remains independent and uses `DefaultCTSetupTemplateID=CT_2A_1Mid_2B` only after deterministic retries fail.

## Stable Route IDs

T routes:

- `D2_T_A_LONG`
- `D2_T_A_SHORT`
- `D2_T_B_TUNNEL`
- `D2_T_MID_B`
- `D2_T_MID_CONTROL`
- `D2_T_A_FAKE_PRESSURE`
- `D2_T_ROTATE_A_TO_B`
- `D2_T_ROTATE_B_TO_A`

CT routes:

- `D2_CT_A_LONG_HOLD`
- `D2_CT_A_SHORT_HOLD`
- `D2_CT_B_HOLD`
- `D2_CT_MID_HOLD`
- `D2_CT_REINFORCE_A`
- `D2_CT_REINFORCE_B`
- `D2_CT_RETAKE_A_FROM_B`
- `D2_CT_RETAKE_B_FROM_A`

Template allocations always total five and each allocation stays within its referenced Route min/max.

## Stable Scenario IDs

- `SCN_A_LONG_OPENING`, `SCN_A_SHORT_OPENING`, `SCN_B_TUNNEL_OPENING`
- `SCN_MID_CONTROL`
- `SCN_A_SITE_ENTRY`, `SCN_A_SHORT_ENTRY`, `SCN_B_SITE_ENTRY`, `SCN_MID_TO_B_ENTRY`
- `SCN_A_RETAKE`, `SCN_B_RETAKE`
- `SCN_A_BOMB_RESOLUTION`, `SCN_B_BOMB_RESOLUTION`

Each scenario receives exactly ten `ScenarioWeight` modifier rows for `aim`, `reaction`, `positioning`, `awareness`, `teamplay`, `utility`, `composure`, `mobility`, `endurance`, and `discipline`. Integer percentages sum to 100 and the adapter normalizes them to 1.0.

## Stable MapTag IDs

- `D2_LONG_RANGE`, `D2_CLOSE_RANGE`, `D2_CT_HOLD_ANGLE`
- `D2_SITE_PRESSURE_A`, `D2_SITE_PRESSURE_B`, `D2_MID_CONTROL`
- `D2_ROTATE_RISK`, `D2_POSTPLANT_A`, `D2_POSTPLANT_B`
- `D2_FAST_TIMING`, `D2_LOW_TIME`, `D2_T_EXECUTE`

## Core nodes and risk anchors

Core nodes: `T_SPAWN`, `LONG_DOOR`, `A_LONG`, `PIT`, `A_SITE`, `MID`, `CATWALK`, `SHORT`, `B_UPPER`, `B_TUNNEL`, `B_SITE`, `CT_SPAWN`, `CT_MID`, `B_DOOR`, `CAR`.

Risk/intercept anchors: `A_LONG_CROSS`, `MID_DOOR_CROSS`, `SHORT_RAMP`, `B_TUNNEL_EXIT`, `B_DOOR_CROSS`.

## Missing CombatConst keys to add

- Time/action: `MinPickupTime`, `MaxPickupTime`, `DecisionDelay`, `MinCombatDuration`, `MaxCombatDuration`, `PulseFireWindow`
- Combat probability: `MinDamagePotential`, `MaxDamagePotential`, `MinExposureModifier`, `MaxExposureModifier`
- Strategy defaults: `DefaultStrategyTemplateID`, `DefaultCTSetupTemplateID`
- Intel: `CommunicationDelay`, `SoundIntelMinConfidence`, `SoundIntelMaxConfidence`, `DeathIntelMaxConfidence`, `MinIntelTTL`, `MaxIntelTTL`
- Utility and guards: `UtilityBudget`, `MaxStateTransitions`, `MaxScheduledActions`, `MaxEffectsPerTimestamp`, `MaxNoOpTransitions`, `MaxRotationsPerTeam`, `MaxRoundTimeline`

`TargetHPFactor`, forced winner, target terminal and target survivor keys are intentionally absent.

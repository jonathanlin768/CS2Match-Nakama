// Package matchengine simulates complete matches from self-contained input
// snapshots and returns authoritative reports. Business callers use
// Service.Simulate; configuration loading, RPCs and persistence belong to the
// calling business package.
//
// Public models are grouped in input.go, report.go and map_config.go. Mutable
// round state, actions, scheduling and resolvers are package implementation
// details. Each simulation owns its match and round state. Configuration
// snapshots are read-only and random draws derive from stable seed identities.
//
// CalibrateRounds, RoundInput, StrategyMemory and CalibrationSummary form a
// separate offline calibration API. They execute the same causal round engine
// and are not alternate business endpoints for truncated matches.
//
// See README.md for the execution chain, file responsibilities, state ownership
// and known implementation gaps.
package matchengine

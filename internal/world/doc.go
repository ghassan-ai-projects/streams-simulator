// Package world is the seeded discrete-event world core. It holds hidden
// state, dynamics, faults and effectors, and produces the native
// sim-event-v0.1 stream. It is strictly single-goroutine; concurrency lives
// only in the sinks, behind a bounded channel with ordered writes.
//
// A run is a pure function of (sim_version, domain_digest, adapter_digest,
// seed, command_log, sink). The world honors its half: nothing here reads a
// wall clock, iterates a map to produce output, or draws from a shared
// generator.
//
// The package is a facade: the rules live in its private domain layer.
package world

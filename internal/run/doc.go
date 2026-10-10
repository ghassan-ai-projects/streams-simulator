// Package run wires the pipeline world -> perturbation -> adapter -> sink ->
// ledger into one deterministic entity, records every mutation in a command
// log, and can collapse the whole run into a run artifact that replay
// reproduces byte-for-byte with no server running.
//
// A run is a pure function of (sim_version, domain_digest, adapter_digest,
// seed, command_log, sink).
//
// The package is a facade: the orchestration lives in its private app layer.
package run

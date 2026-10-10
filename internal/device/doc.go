// Package device is a wire-faithful emulator of the serial device at the far
// end of the Agentic Stream serial-effector boundary (Real-World Sensor HIL-0).
//
// It speaks the device wire contract — command / receipt / result / state
// records — owned by Agentic Stream (see contract/SOURCE.md), enforces the
// safety envelope a real firmware+gateway would (boot identity, expiry, target
// and operation allowlist, parameter bounds, idempotency), and drives a plant
// so that state and verification reflect what the world actually did rather
// than what a command asked for. It is a test instrument: the effect a command
// has on the plant is the ground truth, and an acknowledgement is never proof
// of that effect.
//
// This package owns no transport framing beyond one-record-per-line NDJSON;
// raw serial bytes, reconnection, and device identity remain a gateway concern.
//
// The package is a facade: the rules live in its private domain layer and
// the unix-socket transport in its uds edge.
package device

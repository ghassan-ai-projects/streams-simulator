// Package deviceworld binds the device emulator's Plant seam to the Streams
// Simulator world, so the world is the effector oracle behind the wire loop.
//
// With this plant, a device command does not "energize" an output because a
// heuristic says duty>0; it energizes because the world's effector actually
// applied a physical effect. The world's failure modes then become the device's
// observed truth: a confirmed-no-effect or interlock refusal surfaces as
// energized=false — the desired≠observed case — while a silent-no-effect fools
// the device's self-reported state exactly as it would fool a real confirmation
// channel, and only the independent process reading (Value) betrays it.
//
// The package is a facade: the rules live in its private domain layer.
package deviceworld

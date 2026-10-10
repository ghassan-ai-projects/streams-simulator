// Package refconsumer is the reference consumer: a deliberately simple
// moving-window detector that closes the loop against the simulator the way
// a real consumer would. It is config-driven (threshold, window, which
// effector to actuate) and carries no domain knowledge of its own — the
// same binary serves as a baseline for any domain. It never reads the
// ledger, never reads ground truth, and treats every actuator ack as
// provisional.
//
// The package is a facade: the rules live in its private domain layer and the
// MCP operator client in its mcpclient edge.
package refconsumer

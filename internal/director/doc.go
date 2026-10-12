// Package director runs worlds on behalf of a surface. It owns the world
// registry and identities, the world commands (advance, fault, perturb,
// environment), the run lifecycle (begin, end, seal and reveal truth, score,
// verify, audit) and the operator view a capability token resolves to: the
// only door the system under test has, which sees no truth, no faults and no
// hidden state. Callers pass typed values; the module knows nothing of any
// protocol, so an MCP server, a CLI or a test drives it the same way.
//
// The package is a facade: the use cases live in its private app layer, the
// refusal taxonomy and operator view in its domain layer, and the
// capability-token entropy in its capability edge.
package director

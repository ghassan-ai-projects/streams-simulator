// Package mcp implements one MCP server with two roles: director (build,
// drive, perturb and audit worlds; holds the truth) and operator (act on a
// world as a consumer of it; sees no truth, no faults, no hidden state).
// Every tool is domain-agnostic: the binary contains no effector name, no
// consumer name, no consumer schema.
//
// The package is a facade over protocol only: tool definitions, schemas,
// argument decoding and SDK registration live in its protocol layer, the
// stdio session and operator HTTP endpoint in its transport edge. The
// director's use cases (worlds, runs, truth, scoring, the operator view) live
// in the director module, which mcp reaches only through its facade.
package mcp

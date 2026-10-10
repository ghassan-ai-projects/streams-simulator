// Package mcp implements one MCP server with two roles: director (build,
// drive, perturb and audit worlds; holds the truth) and operator (act on a
// world as a consumer of it; sees no truth, no faults, no hidden state).
// Every tool is domain-agnostic: the binary contains no effector name, no
// consumer name, no consumer schema.
//
// The package is a facade: the use cases live in its private app layer, the
// protocol wiring (tools, schemas, SDK registration) in its protocol layer, and
// the capability-token entropy in its capability edge.
package mcp

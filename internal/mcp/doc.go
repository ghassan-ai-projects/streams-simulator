// Package mcp implements one MCP server with two roles: director (build,
// drive, perturb and audit worlds; holds the truth) and operator (act on a
// world as a consumer of it; sees no truth, no faults, no hidden state).
// Every tool is domain-agnostic: the binary contains no effector name, no
// consumer name, no consumer schema.
//
// The package is a facade: the surface lives in its private app layer.
package mcp

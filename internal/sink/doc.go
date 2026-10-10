// Package sink delivers rendered records. Three sinks: inproc (a buffer),
// file (byte-reproducible), and http-push (stepped: deterministic).
// Evidence leaves on a sink; MCP never carries it.
//
// The package is a facade: the contract and the in-memory sink live in its
// domain layer, the file and HTTP sinks in edge packages.
package sink

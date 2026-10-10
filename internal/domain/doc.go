// Package domain loads, validates and compiles domain specs — the data
// files that define whole simulated worlds. The binary contains no domain
// behavior: every domain in the catalog loads through this one path, and a
// domain that needs a code branch in the binary is a bug in the simulator's
// design, not a missing feature.
//
// The package is a facade: the rules live in its private domain layer and
// file reading in its files edge.
package domain

// Package lint implements the domain plausibility check: a nominal day of a
// domain simulated with no faults and no effector calls, sampled on a fixed
// grid, and judged by a few rules so a composed plant that validates but
// behaves absurdly is caught before use.
//
// The package is a facade: the rules live in its private domain layer.
package lint

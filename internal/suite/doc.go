// Package suite generates graded scenario suites: declared negative-class
// fraction, randomized onset (including pre-degraded starts), perturbation
// coverage, and a generate-audit-regenerate loop that keeps only
// non-trivial scenarios. A domain that cannot produce non-trivial scenarios
// at the declared prevalence reports a terminal state rather than an empty
// directory (G-06).
//
// The package is a facade: the rules live in its private domain layer.
package suite

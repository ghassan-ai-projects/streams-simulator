// Package audit implements the trivial-baseline audit: every candidate
// scenario runs against a panel of one-line detectors fitted with hindsight
// on the scenario itself — deliberately unfair to the scenario. If a
// detector tuned on the answer still cannot separate the fault from its
// control at >= 0.9 balanced accuracy, the scenario is non_trivial and may
// enter the graded suite. Trivial scenarios stay as mechanism regression
// fixtures but never count as evidence that reasoning helped.
//
// The package is a facade: the rules live in its private domain layer.
package audit

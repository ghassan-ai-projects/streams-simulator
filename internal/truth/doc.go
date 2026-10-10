// Package truth computes generative ground truth: the three onset
// timestamps, the sealed label records, and the observability solver that
// makes detection latency measured against a real reference rather than a
// guess. Ground truth is reachable only under the director role; the
// operator view contains no reference to it.
//
// The package is a facade: the rules live in its private domain layer.
package truth

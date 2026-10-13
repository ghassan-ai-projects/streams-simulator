// Package journal is the run journal: two append-only JSON Lines streams
// written while a run happens. The observer stream holds what a consumer could
// know (the run's identity, every delivered observation, clock advances,
// effector acknowledgements, consumer reports); the director stream holds what
// makes the run true (commands, delivery rows, world state, effector outcomes).
// Truth-bearing records never enter the observer stream.
//
// Both streams are pure functions of the run: no wall-clock read, keys in
// sorted order. Each stream has a SHA-256 digest over the lines replay can
// reproduce, which the run artifact records and verify checks. A journal cut at
// any byte reads up to its last complete line.
//
// The package is a facade: the vocabulary and rules live in its private domain
// layer and the files in its files edge. The contract is run-journal-v0.1.
package journal

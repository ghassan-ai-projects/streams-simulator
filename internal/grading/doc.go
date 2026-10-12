// Package grading grades a finished run offline from the files it published:
// it loads the run artifact, the consumer's verdict, the sealed label, the
// delivery ledger and the state history, replays the artifact to recover the
// effector calls, and scores the evidence with the same policies the director
// scores online. A replay that does not reproduce the recorded trace is
// refused rather than graded.
//
// The package is a facade: the decoding and evidence rules live in its private
// domain layer, the workflow in its app layer and the file access in its files
// edge. It composes run and score, which do not know each other.
package grading

package score

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/score/internal/domain"

// Scorecard is the per-run scoring output (the only artifact anyone outside
// the project reads, so it is JSON-stable).
type Scorecard = layer.Scorecard

// InstrumentMetrics are the simulator grading itself: no consumer involved.
type InstrumentMetrics = layer.InstrumentMetrics

// ConsumerMetrics are computed from the submitted verdict against sealed
// truth and the ledger.
type ConsumerMetrics = layer.ConsumerMetrics

// LoopMetrics need actuation; only measurable because the loop closes.
type LoopMetrics = layer.LoopMetrics

// JudgmentMetrics need a model; reported per scenario, never gated in CI.
type JudgmentMetrics = layer.JudgmentMetrics

// Evidence is everything the scorer reads about one run: the submitted
// verdict, the delivery ledger, the director-side effector log, the applied
// perturbations, the emitted count and the hidden-state history, plus the
// domain whose faults give the recovery levels. The scorer never reaches
// back into the run that produced it.
type Evidence = layer.Evidence

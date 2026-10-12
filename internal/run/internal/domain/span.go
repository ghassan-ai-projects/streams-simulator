package domain

import "math"

// MaxRunSpanNS bounds how far a run's clock may move past its start: 30 days
// of simulated time, ten times the longest shipped profile. A run keeps its
// trace, ledger and history in memory as it goes, so an unbounded span is
// unbounded memory; the bound makes every run, however it is driven, finite.
const MaxRunSpanNS = int64(30*24*3600) * 1_000_000_000

// ExceedsRunSpan reports whether a clock at toNS is further than MaxRunSpanNS
// past a run that started at startNS. A start so late that the bound passes
// the largest time has no time beyond its span, so no sum overflows.
func ExceedsRunSpan(startNS, toNS int64) bool {
	if startNS > math.MaxInt64-MaxRunSpanNS {
		return false
	}
	return toNS > startNS+MaxRunSpanNS
}

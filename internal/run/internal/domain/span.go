package domain

import "math"

const MaxRunSpanNS = int64(30*24*3600) * 1_000_000_000

func ExceedsRunSpan(startNS, toNS int64) bool {
	if startNS > math.MaxInt64-MaxRunSpanNS {
		return false
	}
	return toNS > startNS+MaxRunSpanNS
}

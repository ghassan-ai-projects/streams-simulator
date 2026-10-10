//go:build simdet

// Package wall is the wall-clock seam of the simulator. Under the simdet
// build tag the wall clock is unavailable: Now returns the zero time. The
// deterministic test suite builds with -tags simdet, so no test can
// accidentally depend on a wall clock (see DEFERRED P-01: most clock reads
// still bypass this seam).
package wall

import "time"

// Now is unavailable under simdet. Any caller that reaches here has a
// determinism bug.
func Now() time.Time {
	return time.Time{}
}

// Package clock is the wall-clock edge of the run module: the two timestamps
// a run records (artifact creation and unblinding) are read here, so the
// orchestration and rules stay free of clock reads.
package clock

import "time"

// Stamp returns the current wall time as RFC 3339 with nanoseconds, UTC.
func Stamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

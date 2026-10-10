// Package clock is the wall-clock edge of the run module: the two timestamps
// a run records (artifact creation and unblinding) are read here, through the
// declared wall seam, so the orchestration and rules stay free of clock reads
// and a -tags simdet build records the zero time instead of the real one.
package clock

import (
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/wall"
)

// Stamp returns the current wall time as RFC 3339 with nanoseconds, UTC.
func Stamp() string {
	return wall.Now().UTC().Format(time.RFC3339Nano)
}

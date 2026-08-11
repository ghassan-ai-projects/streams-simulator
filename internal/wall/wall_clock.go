//go:build !simdet

package wall

import "time"

// Now returns the wall clock. World-model code never calls this: only the
// http-push wall sub-mode (real-time delivery) and the CLI's ack-latency
// padding may, and the deterministic suite is built with -tags simdet so it
// cannot link them.
func Now() time.Time {
	return time.Now()
}

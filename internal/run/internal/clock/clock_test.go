//go:build !simdet

package clock

import (
	"testing"
	"time"
)

func TestStampIsTheCurrentUTCTime(t *testing.T) {
	t.Parallel()
	stamp, err := time.Parse(time.RFC3339Nano, Stamp())
	if err != nil {
		t.Fatalf("stamp is not RFC 3339: %v", err)
	}
	if delta := time.Since(stamp); delta < 0 || delta > time.Minute {
		t.Fatalf("stamp is %v away from now", delta)
	}
}

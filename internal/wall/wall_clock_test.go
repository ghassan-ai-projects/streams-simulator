//go:build !simdet

package wall

import (
	"testing"
	"time"
)

func TestNowReadsTheWallClock(t *testing.T) {
	t.Parallel()
	before := time.Now()
	got := Now()
	if got.Before(before) || time.Since(got) > time.Minute {
		t.Fatalf("Now() = %v, want a reading taken after %v", got, before)
	}
}

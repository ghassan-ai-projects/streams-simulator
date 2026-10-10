package clock

import (
	"testing"
	"time"
)

func TestStampIsUTCRFC3339Nano(t *testing.T) {
	t.Parallel()
	parsed, err := time.Parse(time.RFC3339Nano, Stamp())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Location() != time.UTC {
		t.Fatalf("location = %v", parsed.Location())
	}
}

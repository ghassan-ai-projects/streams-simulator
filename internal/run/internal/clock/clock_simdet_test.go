//go:build simdet

package clock

import "testing"

func TestStampIsTheZeroTimeUnderTheDeterministicBuild(t *testing.T) {
	t.Parallel()
	if got := Stamp(); got != "0001-01-01T00:00:00Z" {
		t.Fatalf("Stamp() = %q, want the zero time under simdet", got)
	}
}

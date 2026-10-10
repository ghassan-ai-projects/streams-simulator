//go:build simdet

package wall

import "testing"

func TestNowIsUnavailableUnderTheDeterministicBuild(t *testing.T) {
	t.Parallel()
	if got := Now(); !got.IsZero() {
		t.Fatalf("Now() = %v, want the zero time under simdet", got)
	}
}

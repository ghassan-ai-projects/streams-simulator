package testsupport

import "testing"

func SkipUnderRace(tb testing.TB, reason string) {
	tb.Helper()
	if raceEnabled {
		tb.Skip("single-goroutine simulation, nothing for the race detector (" + reason + "); runs in the non-race pass")
	}
}

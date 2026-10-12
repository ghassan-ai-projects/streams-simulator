package testsupport

import "testing"

// SkipUnderRace skips a test when the binary is built with -race. It is for
// one kind of test only: a deterministic, single-goroutine simulation run (a
// world generated, audited or scored from start to finish) that shares nothing
// with another goroutine. The race detector slows such a run by an order of
// magnitude and has nothing to find in it; the same test still runs, with its
// full assertions, in the ordinary (non-race) pass of `make test-ci`, which is
// also where coverage is measured. A test that starts a goroutine, takes a
// lock another goroutine contends for, or touches a socket must stay under
// -race: test/architecture refuses a SkipUnderRace call in a test that starts
// a goroutine, and the reason must say what the test computes.
func SkipUnderRace(tb testing.TB, reason string) {
	tb.Helper()
	if raceEnabled {
		tb.Skip("single-goroutine simulation, nothing for the race detector (" + reason + "); runs in the non-race pass")
	}
}

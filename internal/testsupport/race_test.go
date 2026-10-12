package testsupport

import "testing"

func TestSkipUnderRaceSkipsExactlyWhenTheBinaryIsRaceInstrumented(t *testing.T) {
	t.Parallel()
	reachedAfterSkip := false
	// The parent's cleanup runs after its parallel subtests have finished.
	t.Cleanup(func() {
		if reachedAfterSkip == raceEnabled {
			t.Errorf("reached after SkipUnderRace = %v, race enabled = %v: the test must run unless -race", reachedAfterSkip, raceEnabled)
		}
	})
	t.Run("skippable", func(t *testing.T) {
		t.Parallel()
		SkipUnderRace(t, "pure computation")
		reachedAfterSkip = true
	})
}

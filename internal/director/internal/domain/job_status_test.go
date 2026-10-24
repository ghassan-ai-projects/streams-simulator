package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestACallOutcomeKeepsTheReasonOnlyForARefusal(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		ack  *model.EffectorAcknowledgement
		want JobCallOutcome
	}{
		"accepted": {&model.EffectorAcknowledgement{Accepted: true, EffectETANS: 90, Reason: "ignored"},
			JobCallOutcome{AtS: 5, EntityID: "e", Effector: "f", Accepted: true, EffectETANS: 90}},
		"refused with a reason": {&model.EffectorAcknowledgement{Reason: "interlock_refused"},
			JobCallOutcome{AtS: 5, EntityID: "e", Effector: "f", Reason: "interlock_refused"}},
		"refused without a reason": {&model.EffectorAcknowledgement{},
			JobCallOutcome{AtS: 5, EntityID: "e", Effector: "f", Reason: "no reason given"}},
		"no acknowledgement": {nil, JobCallOutcome{AtS: 5, EntityID: "e", Effector: "f", Reason: "no reason given"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := NewJobCallOutcome(5, "e", "f", tc.ack); got != tc.want {
				t.Fatalf("outcome = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestOnlyRunningAndStoppingJobsAreActive(t *testing.T) {
	t.Parallel()
	active := map[JobState]bool{JobRunning: true, JobStopping: true, JobCompleted: false, JobStopped: false, JobFailed: false}
	for state, want := range active {
		if JobActive(state) != want {
			t.Fatalf("%s active = %v, want %v", state, JobActive(state), want)
		}
	}
}

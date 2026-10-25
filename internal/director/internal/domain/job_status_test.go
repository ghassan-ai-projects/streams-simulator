package domain

import (
	"fmt"
	"reflect"
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

func TestOnlyRunningPausedAndStoppingJobsAreActive(t *testing.T) {
	t.Parallel()
	active := map[JobState]bool{JobRunning: true, JobPaused: true, JobStopping: true, JobCompleted: false, JobStopped: false, JobFailed: false}
	for state, want := range active {
		if JobActive(state) != want {
			t.Fatalf("%s active = %v, want %v", state, JobActive(state), want)
		}
	}
}

func TestAnImmediateCallOutcomeIsFlaggedAndKeepsTheTimeItWasApplied(t *testing.T) {
	t.Parallel()
	got := NewImmediateCallOutcome(42, "e", "f", &model.EffectorAcknowledgement{Accepted: true, EffectETANS: 7})
	want := JobCallOutcome{AtS: 42, EntityID: "e", Effector: "f", Accepted: true, EffectETANS: 7, Immediate: true}
	if got != want {
		t.Fatalf("outcome = %+v, want %+v", got, want)
	}
}

func TestANewJobStatusEchoesTheNamePurposeAndTheDisturbancesOfThePlan(t *testing.T) {
	t.Parallel()
	request := recoveryJob()
	request.Name, request.Purpose, request.Explore = "Loss and recovery", "Does the pond recover?", true
	status := NewJobStatus("w-1", "r-1", request, planStart)
	if status.Name != request.Name || status.Purpose != request.Purpose || status.State != JobRunning || !status.Explore {
		t.Fatalf("status = %+v, want the plan's name, purpose and mode of a running job", status)
	}
	if status.RequestedEndNS != planStart+100e9 || status.ReachedNS != planStart || len(status.Calls) != 0 {
		t.Fatalf("status times = %+v", status)
	}
	if len(status.Faults) != 3 || status.Faults[1].Fault != "early" || *status.Faults[1].UntilS != 35 || status.Faults[1].FaultID != "" {
		t.Fatalf("faults = %+v, want the planned faults, none injected yet", status.Faults)
	}
	if len(status.Perturbations) != 2 || status.Perturbations[0].Kind != "drop" || *status.Perturbations[0].UntilS != 70 {
		t.Fatalf("perturbations = %+v", status.Perturbations)
	}
}

func TestAStatusCloneSharesNothingWithItsSource(t *testing.T) {
	t.Parallel()
	request := recoveryJob()
	request.Explore = true
	status := NewJobStatus("w-1", "r-1", request, planStart)
	status.Calls = append(status.Calls, JobCallOutcome{Effector: "e"})
	clone := JobStatusView(status)
	status.Calls[0].Effector, status.Faults[0].FaultID, status.Perturbations[0].Kind = "x", "f-9", "x"
	if clone.Calls[0].Effector != "e" || clone.Faults[0].FaultID != "" || clone.Perturbations[0].Kind != "drop" {
		t.Fatalf("clone = %+v changed with its source", clone)
	}
}

func TestAJobWithoutDisturbancesHasEmptyListsNotNulls(t *testing.T) {
	t.Parallel()
	status := NewJobStatus("w-1", "r-1", JobRequest{Domain: "d", DurationS: 10}, planStart)
	if status.Faults == nil || status.Perturbations == nil || status.Calls == nil {
		t.Fatalf("status = %+v, want empty lists", status)
	}
}

func TestAnEvaluateJobsPlanIsHiddenUntilTheJobHasEndedAndAnExploreJobsNever(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{true, false} {
		for _, state := range []JobState{JobRunning, JobPaused, JobStopping, JobCompleted, JobStopped, JobFailed} {
			t.Run(fmt.Sprintf("explore=%v/%s", explore, state), func(t *testing.T) {
				t.Parallel()
				request := recoveryJob()
				request.Explore = explore
				status := NewJobStatus("w-1", "r-1", request, planStart)
				status.State, status.Faults[0].FaultID = state, "f-2"
				status.Calls = append(status.Calls, JobCallOutcome{Effector: "e"})
				view := JobStatusView(status)
				hidden := !explore && JobActive(state)
				if hidden != (len(view.Faults) == 0 && len(view.Perturbations) == 0) || view.Faults == nil || view.Perturbations == nil {
					t.Fatalf("view = %+v, want the plan hidden = %v and the lists never null", view, hidden)
				}
				if !hidden && (len(view.Faults) != 3 || len(view.Perturbations) != 2 || view.Faults[0].FaultID != "f-2") {
					t.Fatalf("view = %+v, want the whole plan", view)
				}
				if len(view.Calls) != 1 || view.State != state || view.Explore != explore || len(status.Faults) != 3 {
					t.Fatalf("view = %+v, want the calls, state and mode kept and the source untouched", view)
				}
			})
		}
	}
}

func TestAnActiveEvaluateJobCarriesNoFailureText(t *testing.T) {
	t.Parallel()
	status := NewJobStatus("w-1", "r-1", JobRequest{Domain: "d", DurationS: 10}, planStart)
	status.Failure = "fault canary_fault refused"
	if view := JobStatusView(status); view.Failure != "" {
		t.Fatalf("failure = %q, want none while the blind holds", view.Failure)
	}
	status.State = JobFailed
	if view := JobStatusView(status); view.Failure == "" {
		t.Fatal("a failed job lost its failure text")
	}
}

func calledJob() JobRequest {
	request := recoveryJob()
	request.Calls = []JobCall{
		{EntityID: "p-2", Effector: "late_call", Args: map[string]any{"level": 1}, AtS: 90},
		{EntityID: "p-1", Effector: "early_call", AtS: 5},
	}
	return request
}

func TestANewJobStatusListsThePlannedEffectorCallsInPlanOrderWithArgumentsNeverNull(t *testing.T) {
	t.Parallel()
	status := NewJobStatus("w-1", "r-1", calledJob(), planStart)
	want := []JobPlannedCall{
		{EntityID: "p-2", Effector: "late_call", Args: map[string]any{"level": 1}, AtS: 90},
		{EntityID: "p-1", Effector: "early_call", Args: map[string]any{}, AtS: 5},
	}
	if !reflect.DeepEqual(status.PlannedCalls, want) {
		t.Fatalf("planned calls = %+v, want %+v", status.PlannedCalls, want)
	}
	if none := NewJobStatus("w-1", "r-1", JobRequest{Domain: "d", DurationS: 10}, planStart); none.PlannedCalls == nil || len(none.PlannedCalls) != 0 {
		t.Fatalf("planned calls = %#v, want an empty list, not null", none.PlannedCalls)
	}
}

func TestPlannedCallsAreTheOperatorsOwnCommandsSoEveryModeAndStateShowsThem(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{true, false} {
		for _, state := range []JobState{JobRunning, JobPaused, JobStopping, JobCompleted, JobStopped, JobFailed} {
			t.Run(fmt.Sprintf("explore=%v/%s", explore, state), func(t *testing.T) {
				t.Parallel()
				request := calledJob()
				request.Explore = explore
				status := NewJobStatus("w-1", "r-1", request, planStart)
				status.State = state
				view := JobStatusView(status)
				if len(view.PlannedCalls) != 2 || view.PlannedCalls[0].Effector != "late_call" {
					t.Fatalf("view = %+v, want both planned calls in plan order", view.PlannedCalls)
				}
			})
		}
	}
}

func TestAStatusViewsPlannedCallsShareNothingWithTheSource(t *testing.T) {
	t.Parallel()
	status := NewJobStatus("w-1", "r-1", calledJob(), planStart)
	view := JobStatusView(status)
	status.PlannedCalls[0].Effector, status.PlannedCalls[0].Args["level"] = "x", 9
	if view.PlannedCalls[0].Effector != "late_call" || view.PlannedCalls[0].Args["level"] != 1 {
		t.Fatalf("view = %+v changed with its source", view.PlannedCalls)
	}
}

func TestAnImmediateFaultIsAFlaggedViewAtTheSecondItWasInjectedWithItsId(t *testing.T) {
	t.Parallel()
	got := NewImmediateFault("p-1", "probe_fouling", "f-4", 240)
	want := JobFaultView{JobFault: JobFault{EntityID: "p-1", Fault: "probe_fouling", AtS: 240}, FaultID: "f-4", Immediate: true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fault = %+v, want %+v", got, want)
	}
}

func TestAnEvaluateJobHidesImmediateFaultsAndClearTimesWithTheRestOfItsPlanWhileActive(t *testing.T) {
	t.Parallel()
	cleared := int64(35)
	status := NewJobStatus("w-1", "r-1", calledJob(), planStart)
	status.Faults[1].ClearedAtS = &cleared
	status.Faults = append(status.Faults, NewImmediateFault("p-1", "canary_fault", "f-9", 12))
	if view := JobStatusView(status); len(view.Faults) != 0 || len(view.PlannedCalls) != 2 {
		t.Fatalf("view = %+v, want no faults and the operator's calls", view)
	}
	status.Explore = true
	view := JobStatusView(status)
	if len(view.Faults) != 4 || !view.Faults[3].Immediate || *view.Faults[1].ClearedAtS != 35 {
		t.Fatalf("view = %+v, want the explore job's immediate fault and clear time", view.Faults)
	}
	status.Faults[1].ClearedAtS = nil
	if view.Faults[1].ClearedAtS == nil || *view.Faults[1].ClearedAtS != 35 {
		t.Fatal("the clone shares a clear time with its source")
	}
}

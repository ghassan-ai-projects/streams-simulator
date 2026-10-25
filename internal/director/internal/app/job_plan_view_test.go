package app

import (
	"encoding/json"
	"strings"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

func TestEveryModeShowsThePlannedEffectorCallsBeforeAndAfterTheyRan(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{false, true} {
		t.Run(map[bool]string{false: "evaluate", true: "explore"}[explore], func(t *testing.T) {
			t.Parallel()
			pacer := newGatedPacer()
			d := pacedBy(shippedDirector(t, t.Context()), pacer)
			started := startJob(t, d, blindPondJob(explore))
			want := rules.JobPlannedCall{EntityID: pondEntity, Effector: "start_aerator", Args: aeratorArgs, AtS: 1800}
			status, _ := d.JobStatus(started.WorldID)
			if len(started.PlannedCalls) != 1 || len(status.PlannedCalls) != 1 || len(status.Calls) != 0 || status.PlannedCalls[0].AtS != want.AtS ||
				status.PlannedCalls[0].Effector != want.Effector || status.PlannedCalls[0].Args["pond_id"] != pondEntity {
				t.Fatalf("status = %+v, want the one planned call listed before it ran", status)
			}
			pacer.release(1000)
			if done := awaitJob(t, d, started.WorldID); len(done.PlannedCalls) != 1 || len(done.Calls) != 1 {
				t.Fatalf("status = %+v, want the planned call kept next to its outcome", done)
			}
		})
	}
}

func TestAnEvaluateJobsRawStatusNamesItsCallsButNeverItsFaultsOrDeliveryWhileActive(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	id, active := activeStatuses(t, d, pacer, blindPondJob(false))
	for _, o := range active {
		raw, err := json.Marshal(o.status)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		if !strings.Contains(text, `"planned_calls":[{"entity_id":"site-a/pond-1","effector":"start_aerator"`) {
			t.Fatalf("%s: %s, want the operator's own call listed", o.operation, text)
		}
		for _, canary := range []string{"aerator_failure", `"kind":"drop"`, `"immediate":true,"fault`, "cleared_at_s", `"until_s":1200`} {
			if strings.Contains(text, canary) {
				t.Fatalf("%s: %s leaks %q while the Evaluate job is active", o.operation, text, canary)
			}
		}
	}
	terminalStatuses(t, d, pacer, id)
}

func TestAFaultInjectedNowIsListedAfterThePlannedOnesAtItsSecondWithItsId(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := pausedAt(t, d, pacer, blindPondJob(true), 2)
	res, err := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling")
	if err != nil {
		t.Fatal(err)
	}
	want := rules.NewImmediateFault(pondEntity, "do_probe_fouling", res.FaultID, 240)
	if len(res.Faults) != 2 || res.Faults[1] != want || res.Faults[0].Immediate || res.Faults[0].Fault != "aerator_failure" {
		t.Fatalf("faults = %+v, want the planned fault then %+v", res.Faults, want)
	}
	finish(t, d, pacer, started.WorldID)
	if status, _ := d.JobStatus(started.WorldID); len(status.Faults) != 2 || status.Faults[1] != want {
		t.Fatalf("faults = %+v, want the immediate fault kept to the end", status.Faults)
	}
}

func TestAPlannedClearIsShownAtTheSecondItRanAndOnlyForTheFaultItCleared(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := pausedAt(t, d, pacer, blindPondJob(true), 2)
	if _, err := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling"); err != nil {
		t.Fatal(err)
	}
	if status, _ := d.JobStatus(started.WorldID); status.Faults[0].ClearedAtS != nil {
		t.Fatalf("faults = %+v, want no clear time before the clear ran", status.Faults)
	}
	finish(t, d, pacer, started.WorldID)
	status, _ := d.JobStatus(started.WorldID)
	cleared := status.Faults[0].ClearedAtS
	if cleared == nil || *cleared != 1200 || *status.Faults[0].UntilS != 1200 || status.Faults[1].ClearedAtS != nil {
		t.Fatalf("faults = %+v, want only the planned fault cleared, at 1200 s", status.Faults)
	}
}

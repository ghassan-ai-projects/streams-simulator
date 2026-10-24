package app

import (
	"encoding/json"
	"strings"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

type observed struct {
	operation string
	status    rules.JobStatus
}

func blindPondJob(explore bool) rules.JobRequest {
	req := recoveryPondJob()
	req.Explore = explore
	return req
}

func activeStatuses(t *testing.T, d *Director, pacer *gatedPacer, req rules.JobRequest) (string, []observed) {
	t.Helper()
	started := startJob(t, d, req)
	id := started.WorldID
	<-pacer.asked
	seen := []observed{{"start", started}}
	record := func(operation string, status rules.JobStatus, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", operation, err)
		}
		seen = append(seen, observed{operation, status})
	}
	status, err := d.JobStatus(id)
	record("status", status, err)
	record("list", d.ListJobs()[0], nil)
	status, err = d.PauseJob(id)
	record("pause", status, err)
	status, err = d.StepJob(id)
	record("step", status, err)
	status, err = d.SetJobRate(id, 30)
	record("rate", status, err)
	immediate, err := d.JobInvokeEffector(id, pondEntity, "start_aerator", aeratorArgs)
	record("invoke_effector", immediate.JobStatus, err)
	status, err = d.ResumeJob(id)
	record("resume", status, err)
	status, err = d.StopJob(id)
	record("stop", status, err)
	return id, seen
}

func terminalStatuses(t *testing.T, d *Director, pacer *gatedPacer, id string) []observed {
	t.Helper()
	pacer.release(1000)
	finished := awaitJob(t, d, id)
	stopped, err := d.StopJob(id)
	if err != nil {
		t.Fatal(err)
	}
	return []observed{{"status", finished}, {"list", d.ListJobs()[0]}, {"stop", stopped}}
}

func plannedDisturbancesShown(t *testing.T, status rules.JobStatus) bool {
	t.Helper()
	raw, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	fault, perturbation := strings.Contains(text, "aerator_failure"), strings.Contains(text, `"kind":"drop"`)
	if fault != perturbation || !strings.Contains(text, `"faults":[`) || !strings.Contains(text, `"perturbations":[`) {
		t.Fatalf("status = %s, want the plan shown or hidden as a whole, with both lists always present", text)
	}
	return fault
}

func TestEveryJobStatusOfAnEvaluateJobHidesItsPlanWhileActiveAndShowsItOnceEnded(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{false, true} {
		t.Run(map[bool]string{false: "evaluate", true: "explore"}[explore], func(t *testing.T) {
			t.Parallel()
			pacer := newGatedPacer()
			d := pacedBy(shippedDirector(t, t.Context()), pacer)
			id, active := activeStatuses(t, d, pacer, blindPondJob(explore))
			for _, o := range active {
				if shown := plannedDisturbancesShown(t, o.status); shown != explore || !rules.JobActive(o.status.State) {
					t.Fatalf("%s (%s) shows the plan = %v, want %v while active", o.operation, o.status.State, shown, explore)
				}
			}
			for _, o := range terminalStatuses(t, d, pacer, id) {
				if !plannedDisturbancesShown(t, o.status) || rules.JobActive(o.status.State) {
					t.Fatalf("%s (%s) hides the plan, want it shown once the job ended", o.operation, o.status.State)
				}
			}
		})
	}
}

func TestAnExploreJobTakesAHandFaultAndShowsItsIdWhileAnEvaluateJobRefusesIt(t *testing.T) {
	t.Parallel()
	for _, explore := range []bool{false, true} {
		pacer := newGatedPacer()
		d := pacedBy(shippedDirector(t, t.Context()), pacer)
		started := pausedAt(t, d, pacer, blindPondJob(explore), 1)
		res, err := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling")
		if explore {
			if err != nil || res.FaultID == "" || !plannedDisturbancesShown(t, res.JobStatus) {
				t.Fatalf("explore: %+v, %v, want the fault and the plan shown", res, err)
			}
		} else {
			requireToolCode(t, err, rules.CodeEvaluateBlind)
			raw, _ := json.Marshal(res)
			if res.FaultID != "" || strings.Contains(string(raw), "do_probe_fouling") {
				t.Fatalf("evaluate: %s, want nothing of the refused fault", raw)
			}
			if status, _ := d.JobStatus(started.WorldID); status.State != rules.JobPaused {
				t.Fatalf("status = %+v, want the refused command to leave the job paused", status)
			}
		}
		pacer.release(1000)
		d.stopJobAndWait(started.WorldID)
	}
}

func TestAnEvaluateJobFailureStaysOutOfTheActiveStatus(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, blindPondJob(false))
	j := d.jobs.find(started.WorldID)
	j.mu.Lock()
	j.status.Failure = "fault aerator_failure refused"
	j.mu.Unlock()
	if status, _ := d.JobStatus(started.WorldID); status.Failure != "" {
		t.Fatalf("failure = %q, want none while the job is active", status.Failure)
	}
	pacer.release(1000)
	d.stopJobAndWait(started.WorldID)
}

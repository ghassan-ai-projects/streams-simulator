package app

import (
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func secs(n int64) *int64 { return &n }

func recoveryPondJob() rules.JobRequest {
	req := pondJob()
	req.Name, req.Purpose = "Aerator loss and recovery", "Does dissolved oxygen recover once the aerator is back?"
	req.Faults[0].UntilS = secs(1200)
	req.Perturbations = []rules.JobPerturbation{{Kind: "drop", Params: map[string]any{"rate": 0.3}, AtS: 300, UntilS: secs(2400)}}
	return req
}

func commandsOf(art *model.RunArtifact, op string) []model.Command {
	var found []model.Command
	for _, command := range art.CommandLog {
		if command.Op == op {
			found = append(found, command)
		}
	}
	return found
}

func TestARecoveredFaultAndADeliveryConditionAreRunRecordedNamedAndReplayable(t *testing.T) {
	t.Parallel()
	d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
	started := startJob(t, d, recoveryPondJob())
	status := awaitJob(t, d, started.WorldID)
	if status.State != rules.JobCompleted || status.Name != "Aerator loss and recovery" || status.Purpose == "" {
		t.Fatalf("status = %+v, want a completed job that keeps its name and purpose", status)
	}
	if len(status.Faults) != 1 || status.Faults[0].FaultID != "f-0" || *status.Faults[0].UntilS != 1200 || len(status.Perturbations) != 1 {
		t.Fatalf("status plan = %+v %+v, want the injected fault with its clear time and the perturbation", status.Faults, status.Perturbations)
	}
	art := requireReplaysFromTheRunFile(t, d, started.WorldID)
	if art.Plan == nil || art.Plan.Name != status.Name || art.Plan.Purpose != status.Purpose {
		t.Fatalf("artifact plan = %+v, want the name and purpose recorded", art.Plan)
	}
	clears := commandsOf(art, model.OpFaultClear)
	if len(clears) != 1 || clears[0].Args["fault_id"] != "f-0" || clears[0].AtNS != started.StartNS+1200e9 {
		t.Fatalf("clears = %+v, want f-0 cleared at 1200 s", clears)
	}
	if applied := commandsOf(art, model.OpPerturbApply); len(applied) != 1 || art.Counts.DroppedByDesign+art.Counts.Perturbed == 0 {
		t.Fatalf("applied %+v, counts %+v, want the drop applied and felt", applied, art.Counts)
	}
	requireVerifies(t, d, started.WorldID)
}

func TestAnImmediateFaultDoesNotMakeAPlannedClearHitTheWrongFault(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	req := recoveryPondJob()
	req.Explore = true
	started := pausedAt(t, d, pacer, req, 2)
	immediate, err := d.JobInjectFault(started.WorldID, pondEntity, "do_probe_fouling")
	if err != nil || immediate.FaultID != "f-0" {
		t.Fatalf("JobInjectFault = %+v, %v, want the first fault of the world", immediate, err)
	}
	art := finish(t, d, pacer, started.WorldID)
	clears := commandsOf(art, model.OpFaultClear)
	if len(clears) != 1 || clears[0].Args["fault_id"] != "f-1" {
		t.Fatalf("clears = %+v, want the planned fault, now f-1, cleared and the immediate f-0 left alone", clears)
	}
	if status, _ := d.JobStatus(started.WorldID); status.Faults[0].FaultID != "f-1" {
		t.Fatalf("faults = %+v, want the planned fault shown under the id the world gave it", status.Faults)
	}
}

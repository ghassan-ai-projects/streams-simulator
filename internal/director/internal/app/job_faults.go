package app

import (
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type faultLedger struct {
	order    []int
	planned  []string
	injected int
	given    map[string]string
}

func newFaultLedger(faults []rules.JobFault) *faultLedger {
	return &faultLedger{order: rules.InjectionOrder(faults), planned: rules.FaultIDs(faults), given: map[string]string{}}
}

func (ledger *faultLedger) record(faultID string) (index int, ok bool) {
	if ledger.injected >= len(ledger.order) {
		return 0, false
	}
	index = ledger.order[ledger.injected]
	ledger.injected++
	if faultID != ledger.planned[index] {
		ledger.given[ledger.planned[index]] = faultID
	}
	return index, true
}

func (ledger *faultLedger) retarget(plannedID string) string {
	if given, moved := ledger.given[plannedID]; moved {
		return given
	}
	return plannedID
}

func (j *job) noteInjection(faultID string) {
	if index, ok := j.faults.record(faultID); ok {
		j.status.Faults[index].FaultID = faultID
	}
}

func (j *job) retargetClear(command model.ScenarioCommand) {
	if command.Op != model.OpFaultClear {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	planned, _ := command.Args["fault_id"].(string)
	command.Args["fault_id"] = j.faults.retarget(planned)
}

func (j *job) noteClear(command model.ScenarioCommand) {
	faultID, _ := command.Args["fault_id"].(string)
	clearedAtS := (timeArgument(command, "at_ns") - j.status.StartNS) / nsPerSecond
	for i := range j.status.Faults {
		if j.status.Faults[i].FaultID == faultID {
			j.status.Faults[i].ClearedAtS = &clearedAtS
		}
	}
}

package app

import (
	"reflect"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (r *Run) journalFaultTruth() {
	r.sealFaultTruth(r.World.FaultIntervals())
}

func (r *Run) journalFinalFaultTruth() {
	r.sealFaultTruth(r.World.FinalFaultIntervals())
}

func (r *Run) sealFaultTruth(faults []model.FaultInterval) {
	if len(faults) == 0 || reflect.DeepEqual(faults, r.sealedFaults) {
		return
	}
	r.sealedFaults = faults
	r.journal.FaultTruth(r.journalNow(), faults)
	r.flushJournal()
}

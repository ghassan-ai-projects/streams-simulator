package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (l *Library) checkInstrument(i *Instrument) error {
	var issues issueList
	l.checkInstrumentUnits(i, &issues)
	if i.Range != nil && i.Range.Min >= i.Range.Max {
		issues.addf("/range", "range min %v must be below max %v", i.Range.Min, i.Range.Max)
	}
	checkInstrumentFaults(i, &issues)
	checkInstrumentStates(i, &issues)
	return issues.violation(model.ContractInstrument, i.File)
}

func (l *Library) checkInstrumentUnits(i *Instrument, issues *issueList) {
	if _, known := l.vocabulary.UnitFamilies[i.UnitFamily]; !known {
		issues.addf("/unit_family", "unit family %q is not in the vocabulary", i.UnitFamily)
	} else if !l.vocabulary.hasUnit(i.UnitFamily, i.Unit) {
		issues.addf("/unit", "unit %q is not a unit of family %q", i.Unit, i.UnitFamily)
	}
	if _, known := l.vocabulary.TagLetters[i.Tag]; !known {
		issues.addf("/tag", "tag %q is not in the vocabulary", i.Tag)
	} else if !l.vocabulary.tagMeasures(i.Tag, i.UnitFamily) {
		issues.addf("/tag", "tag %q does not measure unit family %q", i.Tag, i.UnitFamily)
	}
}

func checkInstrumentFaults(i *Instrument, issues *issueList) {
	ids := make([]string, len(i.Faults))
	for index, fault := range i.Faults {
		ids[index] = fault.ID
	}
	checkDistinct(ids, "faults", "id", "fault", issues)
}

func checkInstrumentStates(i *Instrument, issues *issueList) {
	if index, repeated := model.RepeatedStateCode(i.States); repeated {
		issues.addf(fmt.Sprintf("/states/%d/value", index), "state value %d is declared twice", i.States[index].Value)
	}
}

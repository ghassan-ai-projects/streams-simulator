package app

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestANonFiniteReadingIsNeverCarriedOnTheWire(t *testing.T) {
	t.Parallel()
	d := newDevices(runOf(t, testsupport.ThermalBoardDomain()), nil, nil)
	at := model.FormatTime(model.DefaultStartTimeNS)
	for _, value := range []any{math.NaN(), math.Inf(1), math.Inf(-1)} {
		records := d.observe(model.SimEvent{EntityID: "zone-01", Channel: "temperature", Value: value, Quality: "good", EventTime: at})
		for _, record := range records {
			if record.Type == "DDATA" {
				t.Fatalf("a %v reading was carried: %+v", value, record)
			}
		}
	}
	records := d.observe(model.SimEvent{EntityID: "zone-01", Channel: "temperature", Value: 24.0, Quality: "good", EventTime: at})
	if len(records) != 1 || records[0].Type != "DDATA" {
		t.Fatalf("a finite reading = %+v, want one DDATA after the birth already sent", records)
	}
}

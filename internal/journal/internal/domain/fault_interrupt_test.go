package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestFaultTruthIsADirectorRecordCarryingTheIntervals(t *testing.T) {
	t.Parallel()
	cleared := int64(250)
	by := "operator"
	j := NewMemory(true)
	j.FaultTruth(200, []model.FaultInterval{{FaultID: "f-1", EntityID: "e-1", Fault: "stuck", InjectedAtNS: 100, OnsetNS: 120, ClearedNS: &cleared, ClearedBy: &by}})
	raw, err := j.Bytes(StreamDirector)
	if err != nil {
		t.Fatal(err)
	}
	records, err := Read(raw)
	if err != nil || len(records) != 1 {
		t.Fatalf("Read = %+v (%v), want one director record", records, err)
	}
	record := records[0]
	faults, ok := record.Fields["faults"].([]any)
	if record.Type != TypeFaultTruth || record.AtNS != 200 || !ok || len(faults) != 1 {
		t.Fatalf("record = %+v, want a fault.truth record at 200 with one interval", record)
	}
	interval, ok := faults[0].(map[string]any)
	if !ok || interval["fault_id"] != "f-1" || interval["fault"] != "stuck" || interval["cleared_by"] != "operator" {
		t.Fatalf("interval = %+v, want fault f-1 stuck and cleared by the operator", faults[0])
	}
	if !Replayable([]byte(`{"type":"fault.truth"}`)) {
		t.Fatal("fault truth is a director record: it belongs to the replayable stream")
	}
}

func TestInterruptIsAnObserverRecordWithItsFieldsAndTheObserverStreamOnly(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	j.Interrupt(9, map[string]any{"event": "control", "op": "release"})
	observer, err := j.Bytes(StreamObserver)
	if err != nil {
		t.Fatal(err)
	}
	records, err := Read(observer)
	if err != nil || len(records) != 1 || records[0].Type != TypeInterrupt || records[0].Fields["op"] != "release" {
		t.Fatalf("observer = %+v (%v), want one interrupt record with op release", records, err)
	}
	director, err := j.Bytes(StreamDirector)
	if err != nil || len(director) != 0 {
		t.Fatalf("director = %q (%v), want no interrupt on the director stream", director, err)
	}
}

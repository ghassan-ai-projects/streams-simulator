package journal

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestFaultTruthIsSealedInTheDirectorStreamOnly(t *testing.T) {
	t.Parallel()
	by, cleared := model.ClearedByCommand, int64(9)
	j := NewMemory(true)
	j.FaultTruth(5, []model.FaultInterval{{FaultID: "f-0", EntityID: "e-1", Fault: "f1", InjectedAtNS: 1, OnsetNS: 2, ClearedNS: &cleared, ClearedBy: &by}})
	for stream, want := range map[string]int{StreamDirector: 1, StreamObserver: 0} {
		raw, err := j.Bytes(stream)
		if err != nil {
			t.Fatal(err)
		}
		records, err := Read(raw)
		if err != nil || len(records) != want || (want == 1 && records[0].Type != TypeFaultTruth) {
			t.Fatalf("%s stream = %+v (%v), want %d fault.truth records", stream, records, err, want)
		}
	}
}

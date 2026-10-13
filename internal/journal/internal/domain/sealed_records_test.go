package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestKPITruthGoesToTheDirectorAndFramingToTheObserver(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	j.KPITruth(3, &model.KPITruth{ToNS: 3})
	j.Framing(3, FramingPreamble, "{}")
	for stream, want := range map[string]string{StreamDirector: TypeKPITruth, StreamObserver: TypeObservation} {
		raw, err := j.Bytes(stream)
		if err != nil {
			t.Fatal(err)
		}
		records, err := Read(raw)
		if err != nil || len(records) != 1 || records[0].Type != want {
			t.Fatalf("%s stream = %+v (%v), want one %s record", stream, records, err, want)
		}
	}
}

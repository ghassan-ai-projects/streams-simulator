package journal

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestKPITruthIsSealedInTheDirectorStreamAndFileDigestsCoverSessionRecords(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	j.KPITruth(7, &model.KPITruth{FromNS: 0, ToNS: 7, Entities: []model.EntityKPIs{{EntityID: "e-1"}}})
	j.Pacing(7, 0, 3, false)
	digests, err := j.Close(7, false, "")
	if err != nil {
		t.Fatal(err)
	}
	for stream, file := range map[string]string{StreamObserver: digests.ObserverFile, StreamDirector: digests.DirectorFile} {
		raw, err := j.Bytes(stream)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		if file != "sha256:"+hex.EncodeToString(sum[:]) {
			t.Errorf("%s: file digest %s is not the digest of every line written", stream, file)
		}
	}
	director, err := j.Bytes(StreamDirector)
	if err != nil {
		t.Fatal(err)
	}
	records, err := Read(director)
	if err != nil || records[0].Type != TypeKPITruth || digests.ObserverFile == digests.Observer || digests.DirectorFile != digests.Director {
		t.Fatalf("director records %+v (%v): want kpi.truth first, an observer file digest that also covers the pacing record, and a director one equal to its stream digest (no session record)", records, err)
	}
	if j.Digests().ObserverFile != digests.ObserverFile {
		t.Fatal("Digests after the close disagree with the close")
	}
}

package domain

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

// sampleJournal writes one record of every type the writers produce.
func sampleJournal(t *testing.T) *Journal {
	t.Helper()
	j := NewMemory(true)
	digest := "sha256:" + strings.Repeat("a", 64)
	j.Opened(Opened{RunID: "r-1", SimVersion: "0.2.0", Domain: model.ArtifactRef{ID: "d", Version: "1", Digest: digest},
		Adapter: model.ArtifactRef{ID: "a", Version: "1", Digest: digest}, Seed: 7, StartTimeNS: 100, EntityIDs: []string{"e-1"}, Mode: ModeEvaluate})
	line := `{"seq":3,"entity_id":"e-1","channel":"temp","value":1.5}`
	j.State(110, model.StateSnapshot{Seq: 0, TimeNS: 110, Entity: "e-1", States: map[string]float64{"temp_true": 1.5}})
	j.Delivery(110, model.LedgerRecord{DeliveryID: 1, Seq: 0, Delivered: true, DeliveryReason: model.DeliveryOK})
	j.Observation(110, line)
	j.Malformed(111, `{"seq":4,"broken":`)
	j.Command(120, model.Command{Seq: 0, AtNS: 120, Op: model.OpClockAdvance, Args: map[string]any{"to_ns": 120}})
	j.ClockAdvanced(120, 120)
	call := Call{Effector: "set", EntityID: "e-1", CommandID: "c-1"}
	j.Outcome(121, call, map[string]any{"accepted": true, "simulated": true, "world_id": "w-1", "command_id": "c-1"}, "")
	j.Acknowledged(121, call, &model.EffectorAcknowledgement{Accepted: true, Simulated: true, WorldID: "w-1", CommandID: "c-1"})
	j.Outcome(122, call, nil, "the effector declined")
	j.ConsumerReported(120, true)
	j.Unblinded(130, ReasonRevealed)
	return j
}

func TestJournalRecordsConformToTheContract(t *testing.T) {
	t.Parallel()
	var doc any
	decoder := json.NewDecoder(bytes.NewReader(schemas.RunJournal()))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	schema, err := jsonschema.Compile(doc)
	if err != nil {
		t.Fatal(err)
	}
	j := sampleJournal(t)
	if _, err := j.Close(140, false, ""); err != nil {
		t.Fatal(err)
	}
	for _, stream := range []string{StreamObserver, StreamDirector} {
		raw, _ := j.Bytes(stream)
		checkLinesAgainstSchema(t, schema, raw)
	}
}

func checkLinesAgainstSchema(t *testing.T, schema *jsonschema.Schema, raw []byte) {
	t.Helper()
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte{'\n'}) {
		var value any
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			t.Fatalf("line %s: %v", line, err)
		}
		if errs := schema.Validate(value); len(errs) > 0 {
			t.Errorf("line %s violates run-journal-v0.1: %s", line, errs[0])
		}
	}
}

func TestJournalKeepsTruthOutOfTheObserverStream(t *testing.T) {
	t.Parallel()
	j := sampleJournal(t)
	observer, _ := j.Bytes(StreamObserver)
	director, _ := j.Bytes(StreamDirector)
	records, err := Read(observer)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range records {
		switch rec.Type {
		case TypeCommand, TypeDelivery, TypeState, TypeEffectorOutcome, TypeRunUnblinded:
			t.Errorf("observer stream holds the truth-bearing record %s", rec.Type)
		}
	}
	if bytes.Contains(observer, []byte(`"seed"`)) || !bytes.Contains(director, []byte(`"seed":7`)) {
		t.Error("the seed belongs to the director stream only")
	}
	for _, truth := range []string{"temp_true", "the effector declined", "revealed", "delivery_reason"} {
		if bytes.Contains(observer, []byte(truth)) || !bytes.Contains(director, []byte(truth)) {
			t.Errorf("%q must be in the director stream only", truth)
		}
	}
}

func TestJournalCloseReturnsBothDigestsAndClosesBothStreams(t *testing.T) {
	t.Parallel()
	j := sampleJournal(t)
	digests, err := j.Close(140, true, "boom")
	if err != nil {
		t.Fatal(err)
	}
	observer, _ := j.Bytes(StreamObserver)
	director, _ := j.Bytes(StreamDirector)
	if digests.Observer != Digest(observer) || digests.Director != Digest(director) || digests != j.Digests() {
		t.Fatalf("digests %+v do not match the bytes", digests)
	}
	if !bytes.Contains(director, []byte(`"error":"boom"`)) || bytes.Contains(observer, []byte("boom")) {
		t.Fatal("the run error belongs to the director stream only")
	}
}

func TestJournalBytesRefusesAnUnknownStream(t *testing.T) {
	t.Parallel()
	if _, err := NewMemory(true).Bytes("sealed"); err == nil || !strings.Contains(err.Error(), `unknown stream "sealed"`) {
		t.Fatalf("Bytes = %v", err)
	}
}

func TestJournalFlushReportsAStreamFailure(t *testing.T) {
	t.Parallel()
	sink := &memorySink{flushErr: errFlush}
	j := New(NewStream(StreamObserver, sink, false), NewStream(StreamDirector, nil, false))
	if err := j.Flush(); err == nil || !strings.Contains(err.Error(), "observer stream") {
		t.Fatalf("Flush = %v, want the observer stream failure", err)
	}
}

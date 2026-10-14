package journal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestFramingIsAnObservationMarkedWithItsKind(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	j.Framing(5, FramingPreamble, `{"record_type":"runtime_config"}`)
	j.Framing(9, FramingPostamble, `{"record_type":"trace_end"}`)
	raw, err := j.Bytes(StreamObserver)
	if err != nil {
		t.Fatal(err)
	}
	records, err := Read(raw)
	if err != nil || len(records) != 2 || records[0].Type != TypeObservation || records[0].Fields["framing"] != FramingPreamble ||
		records[1].Fields["framing"] != FramingPostamble {
		t.Fatalf("Read = %+v (%v), want a preamble and a postamble observation", records, err)
	}
}

func TestNewMemoryKeepsDigestsAndBytesWithoutFiles(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	j.ClockAdvanced(5, 5)
	j.Command(5, model.Command{AtNS: 5, Op: model.OpClockAdvance, Args: map[string]any{"to_ns": 5}})
	digests, err := j.Close(5, false, "")
	if err != nil {
		t.Fatal(err)
	}
	observer, err := j.Bytes(StreamObserver)
	if err != nil || Digest(observer) != digests.Observer || digests != j.Digests() {
		t.Fatalf("digests %+v, observer %q (%v)", digests, observer, err)
	}
	records, err := Read(observer)
	if err != nil || len(records) != 2 || records[1].Type != "run.closed" {
		t.Fatalf("Read = %+v (%v)", records, err)
	}
}

func TestClosedIncompleteReadsTheClosingRecordThroughTheFacade(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	if _, err := j.Close(5, true, "sink failed"); err != nil {
		t.Fatal(err)
	}
	raw, err := j.Bytes(StreamObserver)
	if err != nil {
		t.Fatal(err)
	}
	if incomplete, closed := ClosedIncomplete(raw); !closed || !incomplete {
		t.Fatalf("ClosedIncomplete = %v, %v; want a closing record that says incomplete", incomplete, closed)
	}
}

func TestNewInDirWritesTheTwoJournalFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	j, err := NewInDir(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	j.Unblinded(1, ReasonExplore)
	if _, err := j.Close(2, false, ""); err != nil {
		t.Fatal(err)
	}
	for file, want := range map[string]string{ObserverFile: `"run.closed"`, DirectorFile: `"run.unblinded"`} {
		raw, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil || !strings.Contains(string(raw), want) {
			t.Fatalf("%s = %q (%v), want %s", file, raw, err, want)
		}
	}
}

func TestNewInDirReportsADirectoryItCannotCreate(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewInDir(filepath.Join(blocker, "run"), false); err == nil || !strings.Contains(err.Error(), "journal:") {
		t.Fatalf("NewInDir under a file = %v", err)
	}
}

func TestEveryRecordVerbReachesItsStream(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	j.Opened(Opened{RunID: "r-1", Mode: ModeExplore})
	j.Observation(1, "line")
	j.Malformed(2, "broken")
	j.Acknowledged(3, Call{Effector: "e", EntityID: "x", CommandID: "c"}, &model.EffectorAcknowledgement{CommandID: "c"})
	j.ConsumerReported(3, false)
	j.Delivery(4, model.LedgerRecord{DeliveryID: 1})
	j.State(4, model.StateSnapshot{Entity: "x"})
	j.Outcome(5, Call{CommandID: "c"}, nil, "declined")
	j.Wire(6, "node", map[string]any{"type": "NSYNC"})
	j.Pacing(6, 1, 0, false)
	if err := j.Flush(); err != nil {
		t.Fatal(err)
	}
	observer, _ := j.Bytes(StreamObserver)
	director, _ := j.Bytes(StreamDirector)
	if got, _ := Read(observer); len(got) != 7 {
		t.Fatalf("observer records = %d, want 7", len(got))
	}
	if got, _ := Read(director); len(got) != 4 {
		t.Fatalf("director records = %d, want 4", len(got))
	}
	if first := FirstDifference(observer, observer); first != -1 {
		t.Fatalf("FirstDifference of a journal with itself = %d", first)
	}
}

func TestDecodeLineAndReplayableAreReachableThroughTheFacade(t *testing.T) {
	t.Parallel()
	record, err := DecodeLine([]byte(`{"stream":"observer","seq":0,"type":"run.opened","at_ns":0}`), 1)
	if err != nil || record.Type != TypeRunOpened {
		t.Fatalf("record = %+v (%v), want the opening", record, err)
	}
	if _, err := DecodeLine([]byte(`nope`), 4); err == nil || !strings.HasPrefix(err.Error(), "journal: line 4: ") {
		t.Fatalf("err = %v, want line 4 named", err)
	}
	if !Replayable([]byte(`{"type":"observation"}`)) || Replayable([]byte(`{"type":"consumer.reported"}`)) {
		t.Fatal("Replayable must keep observations and skip session records")
	}
}

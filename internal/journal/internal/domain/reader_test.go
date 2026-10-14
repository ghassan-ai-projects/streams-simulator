package domain

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadRecoversEveryRecordAndFieldOfAJournal(t *testing.T) {
	t.Parallel()
	j := sampleJournal(t)
	director, _ := j.Bytes(StreamDirector)
	records, err := Read(director)
	if err != nil || len(records) == 0 {
		t.Fatalf("Read = %d records, %v", len(records), err)
	}
	if records[0].Type != TypeRunOpened || records[0].Fields["run_id"] != "r-1" {
		t.Fatalf("first record = %+v, want the director's run.opened", records[0])
	}
	first := records[1]
	if first.Stream != StreamDirector || first.Type != TypeState || first.Seq != 1 || first.AtNS != 110 {
		t.Fatalf("first record = %+v", first)
	}
	if _, ok := first.Fields["state"]; !ok || first.Fields["stream"] != nil {
		t.Fatalf("fields = %v, want the type's fields without the envelope", first.Fields)
	}
}

// A crash can cut the file at any byte: the reader reads up to the last
// complete line, and the digest is that of the complete lines.
func TestReadStopsAtTheLastCompleteLineWhereverTheFileIsCut(t *testing.T) {
	t.Parallel()
	j := sampleJournal(t)
	for _, stream := range []string{StreamObserver, StreamDirector} {
		raw, _ := j.Bytes(stream)
		for cut := 0; cut <= len(raw); cut++ {
			prefix := raw[:cut]
			complete := bytes.Count(prefix, []byte{'\n'})
			records, err := Read(prefix)
			if err != nil || len(records) != complete {
				t.Fatalf("%s cut at %d: %d records, err %v, want %d complete lines", stream, cut, len(records), err, complete)
			}
			whole := prefix[:bytes.LastIndexByte(prefix, '\n')+1]
			if Digest(prefix) != Digest(whole) {
				t.Fatalf("%s cut at %d: the torn line changed the digest", stream, cut)
			}
		}
	}
}

func TestReadNamesTheLineThatIsNotARecord(t *testing.T) {
	t.Parallel()
	good := `{"at_ns":1,"seq":0,"stream":"observer","type":"clock.advanced","to_ns":1}` + "\n"
	cases := map[string]string{
		"not json":            "oops\n",
		"not an object":       "[1]\n",
		"no stream":           `{"at_ns":1,"seq":0,"type":"x"}` + "\n",
		"no type":             `{"at_ns":1,"seq":0,"stream":"observer"}` + "\n",
		"fractional seq":      `{"at_ns":1,"seq":0.5,"stream":"observer","type":"x"}` + "\n",
		"seq is not a number": `{"at_ns":1,"seq":"0","stream":"observer","type":"x"}` + "\n",
		"missing at_ns":       `{"seq":0,"stream":"observer","type":"x"}` + "\n",
	}
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			records, err := Read([]byte(good + bad))
			if err == nil || !strings.HasPrefix(err.Error(), "journal: line 2: ") || len(records) != 1 {
				t.Fatalf("Read = %d records, %v, want the first record and a line 2 error", len(records), err)
			}
		})
	}
}

func TestFirstDifferenceNamesTheFirstReplayableRecordThatDiffers(t *testing.T) {
	t.Parallel()
	a := "{\"type\":\"a\"}\n{\"type\":\"b\"}\n{\"type\":\"consumer.reported\"}\n{\"type\":\"c\"}\n"
	cases := []struct {
		name, other string
		want        int
	}{
		{"identical", a, -1},
		{"a session record is not a difference", "{\"type\":\"a\"}\n{\"type\":\"b\"}\n{\"type\":\"c\"}\n", -1},
		{"changed record", "{\"type\":\"a\"}\n{\"type\":\"B\"}\n{\"type\":\"c\"}\n", 1},
		{"shorter", "{\"type\":\"a\"}\n", 1},
		{"longer", a + "{\"type\":\"d\"}\n", 3},
	}
	for _, tc := range cases {
		if got := FirstDifference([]byte(a), []byte(tc.other)); got != tc.want {
			t.Errorf("%s: FirstDifference = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestDecodeLineNamesItsNumberAndReplayableSkipsSessionRecords(t *testing.T) {
	t.Parallel()
	if _, err := DecodeLine([]byte(`nope`), 7); err == nil || !strings.HasPrefix(err.Error(), "journal: line 7: ") {
		t.Fatalf("err = %v, want line 7 named", err)
	}
	if rec, err := DecodeLine([]byte(`{"stream":"observer","seq":0,"type":"run.opened","at_ns":0}`), 1); err != nil || rec.Type != TypeRunOpened {
		t.Fatalf("record = %+v (%v), want the opening", rec, err)
	}
	for line, want := range map[string]bool{
		`{"type":"observation"}`:       true,
		`{"type":"consumer.reported"}`: false,
		`{"type":"wire"}`:              false,
		`not json`:                     true,
	} {
		if got := Replayable([]byte(line)); got != want {
			t.Errorf("Replayable(%s) = %v, want %v", line, got, want)
		}
	}
}

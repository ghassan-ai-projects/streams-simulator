package domain

import "testing"

func TestClosedIncompleteReadsTheClosingRecordOfEitherStream(t *testing.T) {
	t.Parallel()
	for _, incomplete := range []bool{false, true} {
		j := NewMemory(true)
		if _, err := j.Close(10, incomplete, "sink failed"); err != nil {
			t.Fatal(err)
		}
		for _, stream := range []string{StreamObserver, StreamDirector} {
			raw, _ := j.Bytes(stream)
			if got, closed := ClosedIncomplete(raw); !closed || got != incomplete {
				t.Errorf("%s of a run closed incomplete=%v: got %v, closed %v", stream, incomplete, got, closed)
			}
		}
	}
}

func TestClosedIncompleteFindsNoClosingRecordInAnOpenOrTornJournal(t *testing.T) {
	t.Parallel()
	closing := []byte(`{"stream":"observer","seq":1,"at_ns":1,"type":"run.closed","incomplete":true}`)
	cases := map[string][]byte{
		"empty":                nil,
		"a record, no newline": closing,
		"another record last":  []byte(`{"type":"state","incomplete":true}` + "\n"),
		"not JSON last":        []byte("not json\n"),
		"no incomplete field":  []byte(`{"type":"run.closed"}` + "\n"),
		"closing record torn":  append(append([]byte(`{"type":"state"}`+"\n"), closing[:20]...), '\n'),
	}
	for name, raw := range cases {
		if got, closed := ClosedIncomplete(raw); closed || got {
			t.Errorf("%s: got incomplete %v, closed %v; want no closing record", name, got, closed)
		}
	}
	if got, closed := ClosedIncomplete(append(append([]byte{}, closing...), '\n')); !closed || !got {
		t.Errorf("a whole closing line: got %v, %v", got, closed)
	}
}

package domain

import (
	"bytes"
	"testing"
)

// FuzzJournalRead holds the reader to its promises on arbitrary bytes: it
// never panics, returns a prefix of the complete lines, and its digest does
// not depend on a torn last line.
func FuzzJournalRead(f *testing.F) {
	f.Add([]byte(`{"at_ns":1,"seq":0,"stream":"observer","type":"clock.advanced","to_ns":1}` + "\n"))
	f.Add([]byte("{\"at_ns\":1,\"seq\":0,\"stream\":\"observer\",\"type\":\"run.opened\"}\n{\"at_ns\":"))
	f.Add([]byte("\n\n[]\nnull\n{}\n"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		records, _ := Read(raw)
		if len(records) > bytes.Count(raw, []byte{'\n'}) {
			t.Fatalf("%d records from %d complete lines", len(records), bytes.Count(raw, []byte{'\n'}))
		}
		whole := raw[:bytes.LastIndexByte(raw, '\n')+1]
		if Digest(raw) != Digest(whole) || FirstDifference(raw, whole) != -1 {
			t.Fatal("a torn last line must not change the digest")
		}
	})
}

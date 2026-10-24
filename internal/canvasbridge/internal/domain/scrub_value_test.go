package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestScrubValueScrubsEveryDecodedStringAndKeepsTheRest(t *testing.T) {
	t.Parallel()
	seed := uint64(18446744073709551615)
	listing := Listing{Name: "w", Runs: []RunEntry{{WorldID: "c-1", Seed: &seed, IncompleteReason: "open /R&D/ws/a: <gone>", Plan: []byte(`{"p":"/R&D/ws/b"}`)}}}
	got, err := ScrubValue(listing, []string{"/R&D/ws"})
	if err != nil {
		t.Fatal(err)
	}
	want := listing
	want.Runs = []RunEntry{{WorldID: "c-1", Seed: &seed, IncompleteReason: "open a: <gone>", Plan: []byte(`{"p":"b"}`)}}
	if !reflect.DeepEqual(got.Runs, want.Runs) || got.Name != "w" {
		t.Fatalf("ScrubValue = %+v, want %+v", got, want)
	}
}

func TestScrubValueWithoutRootsReturnsTheValueUnchanged(t *testing.T) {
	t.Parallel()
	in := Listing{Name: "w", Domains: []Definition{{Path: "a.json", ID: "a", Instances: 2}}}
	got, err := ScrubValue(in, nil)
	if err != nil || !reflect.DeepEqual(got, in) {
		t.Fatalf("ScrubValue = %+v, %v, want %+v", got, err, in)
	}
}

func TestScrubValueReportsAValueThatCannotBeEncoded(t *testing.T) {
	t.Parallel()
	var unsupported *json.UnsupportedTypeError
	if _, err := ScrubValue(make(chan int), nil); !errors.As(err, &unsupported) {
		t.Fatalf("ScrubValue of a channel = %v, want an unsupported type error", err)
	}
}

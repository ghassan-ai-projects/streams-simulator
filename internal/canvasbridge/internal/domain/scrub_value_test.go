package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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

type lopsided struct{}

var errLopsidedDecode = errors.New("lopsided cannot be decoded")

func (lopsided) MarshalJSON() ([]byte, error) { return []byte(`"/ws/a"`), nil }

func (*lopsided) UnmarshalJSON([]byte) error { return errLopsidedDecode }

func TestScrubValueReportsAScrubbedDocumentThatCannotBeDecodedBack(t *testing.T) {
	t.Parallel()
	if _, err := ScrubValue(lopsided{}, []string{"/ws"}); !errors.Is(err, errLopsidedDecode) {
		t.Fatalf("ScrubValue = %v, want the decode failure of the target type", err)
	}
}

func TestScrubValueReportsATreeTooDeepToDecode(t *testing.T) {
	t.Parallel()
	var nested any = "/ws/a"
	for range 10001 {
		nested = []any{nested}
	}
	var syntax *json.SyntaxError
	if _, err := ScrubValue(nested, []string{"/ws"}); !errors.As(err, &syntax) || !strings.Contains(err.Error(), "canvasbridge: scrub") {
		t.Fatalf("ScrubValue of a 10001 deep value = %v, want a wrapped syntax error", err)
	}
}

func TestScrubValueReachesEveryStringInNestedMapsAndSlicesAndKeepsOtherLeaves(t *testing.T) {
	t.Parallel()
	in := map[string]any{
		"a": []any{"/ws/x", 1.5, true, nil, map[string]any{"b": "/ws/y", "n": 7.0}},
		"c": map[string]any{"d": []any{[]any{"/ws/z"}}},
		"e": "no path",
	}
	want := map[string]any{
		"a": []any{"x", 1.5, true, nil, map[string]any{"b": "y", "n": 7.0}},
		"c": map[string]any{"d": []any{[]any{"z"}}},
		"e": "no path",
	}
	got, err := ScrubValue(in, []string{"/ws"})
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("ScrubValue = %v, %v, want %v", got, err, want)
	}
}

package model

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestAlarmActivationJSONPreservesNanosecondsBeyondFloatingPointPrecision(t *testing.T) {
	t.Parallel()
	want := []int64{1767225740000000001, 1767225740000000002}
	raw, err := json.Marshal(AlarmKPI{Activations: len(want), ActivationTimesNS: want})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	var got []int64
	if err := json.Unmarshal(fields["activation_times_ns"], &got); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("wire activation instants = %v; want exact nanoseconds %v", got, want)
	}
}

package domain

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
)

func TestAnInvalidRecordKeepsItsSeqTypeAndCommandID(t *testing.T) {
	t.Parallel()
	in, ok := Decode([]byte(`{"type":"DCMD","seq":7,"command_id":"c-9","device":5}`))
	if !ok || in.Invalid == nil || in.Seq != 7 || in.Type != "DCMD" || in.CommandID != "c-9" || in.Device != "" {
		t.Fatalf("decoded %+v", in)
	}
	for _, line := range []string{"null", "not json", "[1]"} {
		if _, ok := Decode([]byte(line)); ok {
			t.Fatalf("%q decoded as a record", line)
		}
	}
}

func TestAValidRecordIsReadIntoItsTypedFields(t *testing.T) {
	t.Parallel()
	in, ok := Decode([]byte(`{"type":"DDATA","protocol_version":1,"session_id":"s-1","seq":2,"lane":"data","device":"P1","boot_id":"b",
		"historical":true,"metrics":[{"alias":1,"value":42.5,"quality":"good","event_time_ns":1,"time_source":"device"}],"future_field":1}`))
	if !ok || in.Invalid != nil || in.Device != "P1" || !in.Historical || len(in.Metrics) != 1 || in.Metrics[0].Alias != 1 ||
		in.Metrics[0].Value != json.Number("42.5") {
		t.Fatalf("decoded %+v, %v", in.Message, in.Invalid)
	}
}

func TestSequencesAreReadStrictlyAndNeverWrap(t *testing.T) {
	t.Parallel()
	for value, want := range map[any]bool{json.Number("7"): true, json.Number("7.0"): true, json.Number("-1"): false,
		json.Number("1.5"): false, json.Number("1e300"): false, 3.0: true, -2.0: false, "x": false} {
		if _, ok := SeqOf(value); ok != want {
			t.Errorf("SeqOf(%v) ok = %v, want %v", value, ok, want)
		}
	}
	var in Receiver
	if in.Accept(math.MaxUint64-1) != Gap || in.Accept(math.MaxUint64) != Gap || in.Accept(math.MaxUint64) != Gap || in.Accept(5) != Duplicate {
		t.Fatal("the receiver wrapped past the largest seq")
	}
	if WireID("") != "d" {
		t.Fatalf("WireID of nothing = %q", WireID(""))
	}
}

func TestARecordOverItsSizeOrFromAnotherSessionIsInvalid(t *testing.T) {
	t.Parallel()
	padding := strings.Repeat("x", MaxRecordBytes)
	in, ok := Decode([]byte(`{"type":"NCMD","protocol_version":1,"session_id":"s-1","seq":0,"lane":"control","command":"rebirth","pad":"` + padding + `"}`))
	if !ok || !errors.Is(in.Invalid, ErrTooLarge) {
		t.Fatalf("an oversized NCMD = %v", in.Invalid)
	}
	in, _ = Decode([]byte(`{"type":"NCMD","protocol_version":1,"session_id":"s-1","seq":0,"lane":"control","command":"rebirth"}`))
	if in.Within("s-1", 1); in.Invalid != nil {
		t.Fatalf("an in-session record = %v", in.Invalid)
	}
	for session, version := range map[string]int{"s-2": 1, "s-1": 2} {
		other := in
		if other.Within(session, version); !errors.Is(other.Invalid, ErrNotSession) {
			t.Fatalf("a record for %s v%d = %v", session, version, other.Invalid)
		}
	}
}

func TestTheHostsVersionMustBeTheHighestBothSupport(t *testing.T) {
	t.Parallel()
	if _, err := Negotiate(Message{Online: true, ProtocolVersion: 1, ProtocolMin: 1, ProtocolMax: 2}, 2); !errors.Is(err, ErrRefused) {
		t.Fatalf("a lower common version = %v, want a refusal", err)
	}
	if version, err := Negotiate(Message{Online: true, ProtocolVersion: 2, ProtocolMin: 1, ProtocolMax: 2}, 2); err != nil || version != 2 {
		t.Fatalf("the highest common version = %d, %v", version, err)
	}
}

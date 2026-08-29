package device

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t)})
	out := d.ApplyCommand(validCommand(t, nil))
	for name, record := range map[string]map[string]any{
		"receipt": out.Receipt,
		"result":  out.Result,
		"state":   d.State(),
	} {
		frame, err := EncodeRecord(record)
		if err != nil {
			t.Fatalf("encode %s: %v", name, err)
		}
		if !bytes.HasSuffix(frame, []byte("\n")) {
			t.Fatalf("%s frame must end in a newline", name)
		}
		if _, err := DecodeRecord(frame); err != nil {
			t.Fatalf("decode round-tripped %s: %v", name, err)
		}
	}
}

func TestDecodeFailsClosed(t *testing.T) {
	command, err := EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string][]byte{
		"empty":            {},
		"not-an-object":    []byte("[1,2,3]\n"),
		"unknown-type":     []byte(`{"message_type":"telemetry"}` + "\n"),
		"missing-type":     []byte(`{"protocol_version":1}` + "\n"),
		"trailing-json":    append(append([]byte{}, bytes.TrimRight(command, "\n")...), []byte(" {}\n")...),
		"oversize":         []byte(`{"message_type":"state","x":"` + strings.Repeat("z", maxFrameBytes) + `"}`),
		"schema-violation": []byte(`{"message_type":"command","protocol_version":1}` + "\n"),
	}
	for name, frame := range cases {
		frame := frame
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeRecord(frame); err == nil {
				t.Fatalf("%s must fail closed, but decoded", name)
			}
		})
	}
}

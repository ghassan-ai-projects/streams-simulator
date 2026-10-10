package domain

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	command, err := EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		frame []byte
		want  string
	}{
		"empty":            {[]byte{}, "frame is empty"},
		"not-an-object":    {[]byte("[1,2,3]\n"), "decode frame:"},
		"unknown-type":     {[]byte(`{"message_type":"telemetry"}` + "\n"), `unsupported message_type "telemetry"`},
		"missing-type":     {[]byte(`{"protocol_version":1}` + "\n"), "message_type is required"},
		"trailing-json":    {append(append([]byte{}, bytes.TrimRight(command, "\n")...), []byte(" {}\n")...), "trailing JSON"},
		"oversize":         {[]byte(`{"message_type":"state","x":"` + strings.Repeat("z", MaxFrameBytes) + `"}`), "frame exceeds"},
		"schema-violation": {[]byte(`{"message_type":"command","protocol_version":1}` + "\n"), `missing required property "command_id"`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeRecord(tc.frame); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s: err = %v, want %q", name, err, tc.want)
			}
		})
	}
}

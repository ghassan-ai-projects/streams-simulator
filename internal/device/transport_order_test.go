package device

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"testing"
)

type recordedConnection struct {
	io.Reader
	io.Writer
}

func TestConnectionPreservesMalformedQueryAndCommandFrameOrder(t *testing.T) {
	command, err := EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	input := append([]byte("\ninvalid-json\n"+QueryStateControl+"\n"), command...)
	var output bytes.Buffer
	conn := &recordedConnection{Reader: bytes.NewReader(input), Writer: &output}
	d := New(Config{Capabilities: testCaps(t)})
	if err := ServeConn(conn, d); err != nil {
		t.Fatal(err)
	}
	var types []string
	for _, line := range bytes.Split(bytes.TrimSpace(output.Bytes()), []byte{'\n'}) {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatal(err)
		}
		types = append(types, record["message_type"].(string))
	}
	want := []string{"state", "receipt", "result", "state", "receipt", "result"}
	if !reflect.DeepEqual(types, want) {
		t.Fatalf("frames=%v, want %v", types, want)
	}
	if d.AcceptedCommandCount() != 1 {
		t.Fatal("malformed/query frames consumed a scheduled command")
	}
}

func TestAdmissionPreservesBootBeforeFreshnessBeforeTarget(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t), Clock: func() int64 { return 1000 }})
	command := validCommand(t, nil)
	command["expected_boot_id"] = "wrong"
	command["not_before_mono_us"] = float64(2000)
	command["target"] = "wrong"
	if got := d.admit(command, 1000); got != "wrong_boot" {
		t.Fatalf("first rejection=%q", got)
	}
	command["expected_boot_id"] = d.BootID()
	if got := d.admit(command, 1000); got != "not_ready" {
		t.Fatalf("freshness rejection=%q", got)
	}
	command["not_before_mono_us"] = float64(0)
	if got := d.admit(command, 1000); got != "wrong_target" {
		t.Fatalf("target rejection=%q", got)
	}
}

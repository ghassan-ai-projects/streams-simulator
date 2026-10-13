package device_test

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
)

type recordedConnection struct {
	io.Reader
	io.Writer
}

func TestConnectionPreservesMalformedQueryAndCommandFrameOrder(t *testing.T) {
	t.Parallel()
	command, err := device.EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	input := append([]byte("\ninvalid-json\n"+device.QueryStateControl+"\n"), command...)
	var output bytes.Buffer
	conn := &recordedConnection{Reader: bytes.NewReader(input), Writer: &output}
	stale, err := device.ParseFaultSpec("stale@1")
	if err != nil {
		t.Fatal(err)
	}
	d := device.New(device.Config{Capabilities: testCaps(t), FaultSchedule: []device.FaultInjection{stale}})
	if err := device.ServeConn(conn, d); err != nil {
		t.Fatal(err)
	}
	var types []string
	var receipt map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(output.Bytes()), []byte{'\n'}) {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatal(err)
		}
		types = append(types, record["message_type"].(string))
		if record["message_type"] == "receipt" {
			receipt = record
		}
	}
	want := []string{"state", "receipt", "result", "state", "receipt", "result"}
	if !reflect.DeepEqual(types, want) {
		t.Fatalf("frames=%v, want %v", types, want)
	}
	if receipt["reject_code"] != "not_ready" {
		t.Fatalf("malformed/query frames consumed a scheduled command: %v", receipt)
	}
}

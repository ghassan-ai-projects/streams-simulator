package device_test

import (
	"bytes"
	"encoding/json"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"io"
	"reflect"
	"testing"
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
	d := device.New(device.Config{Capabilities: testCaps(t)})
	if err := device.ServeConn(conn, d); err != nil {
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

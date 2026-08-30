package device

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestServeConnLoop drives the emulator over an in-memory connection exactly as
// a gateway link would: read the opening state, send a command, read the receipt
// and result.
func TestServeConnLoop(t *testing.T) {
	client, server := net.Pipe()
	d := New(Config{Capabilities: testCaps(t)})
	done := make(chan error, 1)
	go func() { done <- ServeConn(server, d) }()

	_ = client.SetDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewReader(client)

	state := readRecord(t, reader)
	if state["message_type"] != "state" {
		t.Fatalf("first frame must be state, got %v", state["message_type"])
	}

	cmd, err := EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(cmd); err != nil {
		t.Fatal(err)
	}

	receipt := readRecord(t, reader)
	if receipt["message_type"] != "receipt" || receipt["accepted"] != true {
		t.Fatalf("expected an accepting receipt, got %v", receipt)
	}

	// Execution truth is read back via a query_state control, not an unsolicited
	// result frame — mirroring the effector's QueryState.
	if _, err := client.Write([]byte(QueryStateControl + "\n")); err != nil {
		t.Fatal(err)
	}
	refreshed := readRecord(t, reader)
	if refreshed["message_type"] != "state" {
		t.Fatalf("query_state must return a state record, got %v", refreshed["message_type"])
	}
	output, ok := refreshed["current_output"].(map[string]any)
	if !ok || output["energized"] != true {
		t.Fatalf("state after an accepted fan command must show energized output, got %v", refreshed["current_output"])
	}

	_ = client.Close()
	if err := <-done; err != nil {
		t.Fatalf("ServeConn returned error: %v", err)
	}
}

func TestListenRefusesRegularFilePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "device.sock")
	if err := os.WriteFile(path, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(path, New(Config{Capabilities: testCaps(t)})); err == nil {
		t.Fatal("device listener must refuse a regular file path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("regular file must remain recoverable: %v", err)
	}
	if string(data) != "keep me" {
		t.Fatalf("regular file was modified: %q", data)
	}
}

func readRecord(t *testing.T, reader *bufio.Reader) map[string]any {
	t.Helper()
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	record, err := DecodeRecord(line)
	if err != nil {
		t.Fatalf("decode frame %q: %v", string(line), err)
	}
	return record
}

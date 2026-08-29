package device

import (
	"bufio"
	"net"
	"testing"
	"time"
)

// TestServeConnLoop drives the emulator over an in-memory connection exactly as
// a gateway link would: read the opening state, send a command, read the receipt
// and result.
func TestServeConnLoop(t *testing.T) {
	client, server := net.Pipe()
	d := New(Config{})
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
	result := readRecord(t, reader)
	if result["message_type"] != "result" || result["status"] != "executed" {
		t.Fatalf("expected an executed result, got %v", result)
	}

	_ = client.Close()
	if err := <-done; err != nil {
		t.Fatalf("ServeConn returned error: %v", err)
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

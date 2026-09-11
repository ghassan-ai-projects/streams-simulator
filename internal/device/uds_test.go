package device

import (
	"bufio"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
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
	result := readRecord(t, reader)
	if result["message_type"] != "result" || result["status"] != "executed" {
		t.Fatalf("expected an executed result, got %v", result)
	}

	// Execution truth is still read back via a query_state control; the result is
	// device-reported terminal status, not independent physical confirmation.
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

func TestScheduledDuplicateReplaysReceiptWithoutSecondPlantEffect(t *testing.T) {
	client, server := net.Pipe()
	plant := &countingPlant{}
	d := New(Config{
		Capabilities:  testCaps(t),
		Plant:         plant,
		FaultSchedule: []FaultInjection{{Name: FaultDuplicate, AcceptedCommand: 1}},
	})
	done := make(chan error, 1)
	go func() { done <- ServeConn(server, d) }()

	_ = client.SetDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewReader(client)
	readRecord(t, reader)
	command, err := EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(command); err != nil {
		t.Fatal(err)
	}
	receipt := readRecord(t, reader)
	if receipt["message_type"] != "receipt" || receipt["accepted"] != true {
		t.Fatalf("duplicate delivery must emit one accepted receipt: %v", receipt)
	}
	result := readRecord(t, reader)
	if result["message_type"] != "result" || result["status"] != "executed" {
		t.Fatalf("duplicate delivery must emit one executed result: %v", result)
	}
	if plant.applyCalls != 1 {
		t.Fatalf("duplicate delivery must apply the plant once, got %d calls", plant.applyCalls)
	}
	if _, err := client.Write([]byte(QueryStateControl + "\n")); err != nil {
		t.Fatal(err)
	}
	state := readRecord(t, reader)
	if state["message_type"] != "state" {
		t.Fatalf("duplicate delivery must not leave a second receipt queued; query_state got %v", state)
	}

	_ = client.Close()
	if err := <-done; err != nil {
		t.Fatalf("ServeConn returned error: %v", err)
	}
}

func TestAckLostRetryReplaysReceiptWithoutSecondPlantEffect(t *testing.T) {
	plant := &countingPlant{}
	d := New(Config{
		Capabilities:  testCaps(t),
		Plant:         plant,
		FaultSchedule: []FaultInjection{{Name: FaultAckLost, AcceptedCommand: 1}},
	})
	tmpDir, err := os.MkdirTemp("/tmp", "ss-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("remove temporary device directory: %v", err)
		}
	})
	path := filepath.Join(tmpDir, "device.sock")
	listener, err := Listen(path, d)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Errorf("close device listener: %v", err)
		}
	})
	var dialer net.Dialer
	client, err := dialer.DialContext(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewReader(client)
	readRecord(t, reader)
	command, err := EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(command); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for d.AcceptedCommandCount() != 1 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if d.AcceptedCommandCount() != 1 {
		t.Fatal("ack_lost command was not admitted")
	}
	_ = client.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if _, err := reader.ReadBytes('\n'); err == nil {
		t.Fatal("ack_lost first delivery must not emit a receipt")
	} else {
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf("expected bounded read timeout before retry, got %v", err)
		}
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := client.Write(command); err != nil {
		t.Fatal(err)
	}
	receipt := readRecord(t, reader)
	if receipt["message_type"] != "receipt" || receipt["accepted"] != true {
		t.Fatalf("retry must replay an accepted receipt: %v", receipt)
	}
	result := readRecord(t, reader)
	if result["message_type"] != "result" || result["status"] != "executed" {
		t.Fatalf("retry must replay an executed result: %v", result)
	}
	if plant.applyCalls != 1 {
		t.Fatalf("ack_lost retry must apply the plant once, got %d calls", plant.applyCalls)
	}
	if _, err := client.Write([]byte(QueryStateControl + "\n")); err != nil {
		t.Fatal(err)
	}
	state := readRecord(t, reader)
	if state["message_type"] != "state" {
		t.Fatalf("query_state after ack_lost retry must return state, got %v", state)
	}

	_ = client.Close()
}

func TestSafeStopAckLostRetryReplaysReceiptAndResult(t *testing.T) {
	plant := &countingPlant{}
	d := New(Config{
		Capabilities:  testCaps(t),
		Plant:         plant,
		FaultSchedule: []FaultInjection{{Name: FaultAckLost, AcceptedCommand: 2}},
	})
	client, server := net.Pipe()
	done := make(chan error, 1)
	go func() { done <- ServeConn(server, d) }()
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewReader(client)
	readRecord(t, reader)

	normal, err := EncodeRecord(validCommand(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(normal); err != nil {
		t.Fatal(err)
	}
	readRecord(t, reader)
	readRecord(t, reader)

	safeStop := validCommand(t, func(c map[string]any) {
		c["command_id"] = "safe-stop/fan-01"
		c["idempotency_key"] = "sha256:" + strings.Repeat("e", 64)
		c["operation"] = "safe_stop"
		c["parameters"] = map[string]any{}
		c["expires_after_ms"] = float64(1000)
	})
	frame, err := EncodeRecord(safeStop)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(frame); err != nil {
		t.Fatal(err)
	}
	_ = client.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if _, err := reader.ReadBytes('\n'); err == nil {
		t.Fatal("ack_lost safe-stop must not emit a receipt")
	} else {
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf("expected bounded safe-stop read timeout, got %v", err)
		}
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := client.Write(frame); err != nil {
		t.Fatal(err)
	}
	receipt := readRecord(t, reader)
	result := readRecord(t, reader)
	if receipt["accepted"] != true || result["status"] != "safe_state" {
		t.Fatalf("safe-stop retry receipt=%v result=%v", receipt, result)
	}
	if plant.applyCalls != 1 || plant.safeStopCalls != 1 {
		t.Fatalf("safe-stop retry applied more than once: apply=%d safe_stop=%d", plant.applyCalls, plant.safeStopCalls)
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

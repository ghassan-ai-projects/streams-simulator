package acceptance

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// deviceTranscriptGolden pins the device emulator's wire behavior for the
// real-world-sensor experiment (governance/CANVAS_QUALITY_BAR.md §4,
// device-path gate). Only the device's monotonic clock readings are masked.
const deviceTranscriptGolden = "testdata/device-serve-transcript.jsonl"

var monotonicReading = regexp.MustCompile(`"([a-z_]*mono_us)":[0-9]+`)

// deviceStep is one request sent to the device and the number of records the
// device answers it with.
type deviceStep struct {
	request string
	answers int
}

func TestDeviceServeTranscriptIsPinned(t *testing.T) {
	t.Parallel()
	conn := startDeviceServe(t)
	command := conformanceCommand(t, "valid", "command.json")
	steps := []deviceStep{
		{command, 2},
		{command, 2},
		{conformanceCommand(t, "invalid", "command-missing-operation.json"), 2},
		{wrongBootCommand(command), 2},
		{`{"message_type":"query_state"}`, 1},
	}
	got := deviceTranscript(t, conn, steps)
	want, err := os.ReadFile(deviceTranscriptGolden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("device serve transcript drifted from %s:\n--- got\n%s--- want\n%s", deviceTranscriptGolden, got, want)
	}
}

// startDeviceServe runs `streamsim device serve` with the experiment's thermal
// capabilities bound to the cold-chain world and returns a connection to it.
func startDeviceServe(t *testing.T) net.Conn {
	t.Helper()
	root := testsupport.RepositoryRoot()
	socket := testsupport.SocketPath(t)
	// #nosec G204 -- the test runs the freshly built binary with fixed args
	cmd := exec.CommandContext(t.Context(), buildBinary(t, root, ""), "device", "serve", "--socket", socket,
		"--capabilities", testsupport.DeviceCapabilities(), "--world", testsupport.DeviceWorldBindings(),
		"--world-domain", testsupport.Domain("cold-chain-transit"))
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start device serve: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Wait() })
	awaitListening(t, stderr)
	var dialer net.Dialer
	conn, err := dialer.DialContext(t.Context(), "unix", socket)
	if err != nil {
		t.Fatalf("connect to device serve: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// deviceStartTimeout bounds how long the device may take to listen.
const deviceStartTimeout = 10 * time.Second

// awaitListening blocks until the device announces its socket on stderr.
func awaitListening(t *testing.T, stderr io.Reader) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), deviceStartTimeout)
	defer cancel()
	listening := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "listening on") {
				listening <- true
				return
			}
		}
		listening <- false
	}()
	select {
	case ok := <-listening:
		if !ok {
			t.Fatal("device serve exited before listening")
		}
	case <-ctx.Done():
		t.Fatalf("device serve did not announce its socket within %s", deviceStartTimeout)
	}
}

// deviceTranscript reads the device's opening record, sends every step and
// returns all records with monotonic readings masked, one per line.
func deviceTranscript(t *testing.T, conn net.Conn, steps []deviceStep) string {
	t.Helper()
	reader := bufio.NewReader(conn)
	var transcript strings.Builder
	transcript.WriteString(maskedRecord(t, reader))
	for _, step := range steps {
		if _, err := conn.Write([]byte(step.request + "\n")); err != nil {
			t.Fatalf("send %s: %v", step.request, err)
		}
		for range step.answers {
			transcript.WriteString(maskedRecord(t, reader))
		}
	}
	return transcript.String()
}

func maskedRecord(t *testing.T, reader *bufio.Reader) string {
	t.Helper()
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read device record: %v", err)
	}
	return monotonicReading.ReplaceAllString(line, `"$1":"MASKED"`)
}

// conformanceCommand is a vendored device-wire command fixture made due at
// once, as one line.
func conformanceCommand(t *testing.T, validity, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(testsupport.DeviceConformance(validity, name)))
	if err != nil {
		t.Fatal(err)
	}
	var command map[string]any
	if err := json.Unmarshal(raw, &command); err != nil {
		t.Fatal(err)
	}
	command["not_before_mono_us"] = 0
	line, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	return string(line)
}

// wrongBootCommand is a new command (own id and idempotency key) addressed to
// a boot the device is not running.
func wrongBootCommand(command string) string {
	return strings.NewReplacer(`"expected_boot_id":"boot-A"`, `"expected_boot_id":"boot-Z"`,
		`"command_id":"cmd-01"`, `"command_id":"cmd-02"`,
		`"idempotency_key":"sha256:aaaa`, `"idempotency_key":"sha256:dddd`).Replace(command)
}

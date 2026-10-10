package uds

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type pipeConn struct {
	io.Reader
	io.Writer
}

func serve(t *testing.T, input string, faults WireFaults) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	d := domain.New(domain.Config{})
	if err := ServeConnWithFaults(&pipeConn{Reader: strings.NewReader(input), Writer: &out}, d, faults); err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("frame %q: %v", scanner.Text(), err)
		}
		records = append(records, record)
	}
	return records
}

func TestSessionOpensWithStateAndAnswersAQueryWithAFreshState(t *testing.T) {
	t.Parallel()
	records := serve(t, QueryStateControl+"\n", WireFaults{})
	if len(records) != 2 || records[0]["message_type"] != "state" || records[1]["message_type"] != "state" {
		t.Fatalf("records = %v", records)
	}
}

func TestMalformedFrameGetsAnOrderedRejectionPair(t *testing.T) {
	t.Parallel()
	records := serve(t, "not json\n", WireFaults{})
	if len(records) != 3 || records[1]["message_type"] != "receipt" || records[2]["message_type"] != "result" {
		t.Fatalf("records = %v", records)
	}
	if records[1]["accepted"] != false {
		t.Fatalf("a malformed command must be rejected: %v", records[1])
	}
}

func TestSessionAppliesTheWireFaultPlanToOutboundFrames(t *testing.T) {
	t.Parallel()
	records := serve(t, QueryStateControl+"\n", WireFaults{Drop: map[int]bool{1: true}})
	if len(records) != 1 || records[0]["message_type"] != "state" {
		t.Fatalf("dropping the first outbound frame leaves one state record, got %v", records)
	}
}

func TestListenRefusesAPathThatIsARegularFileAndReplacesAStaleSocket(t *testing.T) {
	t.Parallel()
	sock := testsupport.SocketPath(t)
	dir := filepath.Dir(sock)
	regular := filepath.Join(dir, "regular")
	if err := os.WriteFile(regular, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(regular, domain.New(domain.Config{})); err == nil || !strings.Contains(err.Error(), "device: refusing to remove non-socket path") {
		t.Fatal("a regular file at the socket path must be refused")
	}
	for i := 0; i < 2; i++ {
		listener, err := Listen(sock, domain.New(domain.Config{}))
		if err != nil {
			t.Fatalf("listen %d: %v", i, err)
		}
		if err := listener.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func capabilities(t *testing.T) *domain.Capabilities {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "thermal_capability_catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	caps, err := domain.LoadCapabilities(data)
	if err != nil {
		t.Fatal(err)
	}
	return caps
}

func commandFrame(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "contract", "conformance", "v1", "valid", "command.json"))
	if err != nil {
		t.Fatal(err)
	}
	var command map[string]any
	if err := json.Unmarshal(raw, &command); err != nil {
		t.Fatal(err)
	}
	command["not_before_mono_us"] = float64(0)
	frame, err := domain.EncodeRecord(command)
	if err != nil {
		t.Fatal(err)
	}
	return string(frame)
}

func serveDevice(t *testing.T, d *domain.Device, input string) []string {
	t.Helper()
	var out bytes.Buffer
	if err := ServeConn(&pipeConn{Reader: strings.NewReader(input), Writer: &out}, d); err != nil && !errors.Is(err, ErrInjectedDisconnect) {
		t.Fatal(err)
	}
	var types []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		types = append(types, record["message_type"].(string))
	}
	return types
}

func TestAcceptedCommandIsAnsweredWithReceiptThenResult(t *testing.T) {
	t.Parallel()
	d := domain.New(domain.Config{Capabilities: capabilities(t)})
	got := serveDevice(t, d, commandFrame(t))
	if strings.Join(got, ",") != "state,receipt,result" {
		t.Fatalf("frames = %v", got)
	}
}

func TestAckLostWithholdsTheReceiptButTheDeviceStillAppliesTheCommand(t *testing.T) {
	t.Parallel()
	d := domain.New(domain.Config{Capabilities: capabilities(t)})
	d.SetFaults(domain.Faults{AckLost: true})
	got := serveDevice(t, d, commandFrame(t))
	if strings.Join(got, ",") != "state" || d.AcceptedCommandCount() != 1 {
		t.Fatalf("frames = %v, accepted = %d", got, d.AcceptedCommandCount())
	}
}

func TestInjectedDisconnectEndsTheSessionAfterTheExchange(t *testing.T) {
	t.Parallel()
	d := domain.New(domain.Config{
		Capabilities:  capabilities(t),
		FaultSchedule: []domain.FaultInjection{{Name: domain.FaultDisconnect, AcceptedCommand: 1}},
	})
	var out bytes.Buffer
	err := ServeConn(&pipeConn{Reader: strings.NewReader(commandFrame(t)), Writer: &out}, d)
	if !errors.Is(err, ErrInjectedDisconnect) {
		t.Fatalf("err = %v, want ErrInjectedDisconnect", err)
	}
}

package plantwire

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func hostSocket(t *testing.T) (string, <-chan []string) {
	t.Helper()
	path := testsupport.SocketPath(t)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	received := make(chan []string, 1)
	go serveHost(listener, received)
	return path, received
}

func serveHost(listener net.Listener, received chan<- []string) {
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	var types []string
	seq := 0
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		var record map[string]any
		_ = json.Unmarshal(scanner.Bytes(), &record)
		types = append(types, record["type"].(string))
		if answer := hostAnswer(record, seq); answer != nil {
			raw, _ := json.Marshal(answer)
			_, _ = conn.Write(append(raw, '\n'))
			seq++
		}
	}
	received <- types
}

func hostAnswer(record map[string]any, seq int) map[string]any {
	envelope := map[string]any{"protocol_version": 1, "session_id": record["session_id"], "seq": seq, "lane": "control"}
	switch record["type"] {
	case "NBIRTH":
		envelope["type"], envelope["host_id"], envelope["online"], envelope["heartbeat"] = "STATE", "host-a", true, 0
		envelope["heartbeat_interval_ms"], envelope["authority_epoch"], envelope["protocol_min"], envelope["protocol_max"] = 1000, 1, 1, 1
		return envelope
	case "NSYNC":
		envelope["type"], envelope["through_seq"] = "HSYNC", record["seq"]
		return envelope
	}
	return nil
}

func pondRun(t *testing.T) *run.Run {
	t.Helper()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(context.Background(), run.Config{Domain: spec, Adapter: a, Seed: 3, SinkName: model.SinkInproc, TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestANodeServesAWholeSessionOverThePrivateSocket(t *testing.T) {
	t.Parallel()
	path, received := hostSocket(t)
	options := Options{Environment: "simulated", ProtocolMax: 1, ContractVersion: "1.0.0", SessionID: "s-1", SliceNS: 60e9, UntilNS: model.DefaultStartTimeNS + 180e9}
	if err := Serve(t.Context(), pondRun(t), path, 5*time.Second, options); err != nil {
		t.Fatal(err)
	}
	types := <-received
	if types[0] != "NBIRTH" || types[len(types)-1] != "NDEATH" || !strings.Contains(strings.Join(types, " "), "DDATA") {
		t.Fatalf("host received %v", types)
	}
}

func TestANodeRefusesASharedSocketAndAnnouncesThePinnedContract(t *testing.T) {
	t.Parallel()
	path, _ := hostSocket(t)
	if err := os.Chmod(filepath.Dir(path), 0o755); err != nil { //nolint:gosec // the test shares the directory on purpose
		t.Fatal(err)
	}
	if err := Serve(t.Context(), pondRun(t), path, time.Second, Options{}); !errors.Is(err, ErrSocketNotPrivate) {
		t.Fatalf("serve = %v, want a refusal", err)
	}
	pinned, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "DIGEST"))
	if err != nil || ContractDigest() != strings.TrimSpace(string(pinned)) {
		t.Fatalf("contract digest %s, pinned %s (%v)", ContractDigest(), pinned, err)
	}
}

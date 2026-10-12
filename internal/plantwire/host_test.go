package plantwire

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const hostPolicy = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

type pairing struct {
	report HostReport
	node   error
	host   error
	run    *run.Run
}

func setpointScript() []Scripted {
	return []Scripted{{AtNS: model.DefaultStartTimeNS + 60e9, Device: "P1", Command: "set_speed", Args: map[string]any{"pct": 80, "lease_ms": 600000},
		Verify: &Verify{Metric: "speed_setpoint", Equals: 80, Tolerance: 0.5, WithinNS: 180e9}}}
}

func pumpRun(t *testing.T) *run.Run {
	t.Helper()
	journalDir := t.TempDir()
	spec, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(context.Background(), run.Config{Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, JournalDir: journalDir})
	if err != nil {
		t.Fatal(err)
	}
	journalDirs.Store(r, journalDir)
	return r
}

var journalDirs sync.Map

func pair(t *testing.T, hostOptions HostOptions, delay time.Duration, prepare func(*run.Run)) pairing {
	t.Helper()
	path := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "h.sock")
	ready, hosted := make(chan struct{}), make(chan pairing, 1)
	hostOptions.Ready = func() { close(ready) }
	go func() {
		report, err := Host(t.Context(), path, delay, hostOptions)
		hosted <- pairing{report: report, host: err}
	}()
	<-ready
	r := pumpRun(t)
	if prepare != nil {
		prepare(r)
	}
	options := Options{Environment: "simulated", ProtocolMax: 1, ContractVersion: "1.0.0", SessionID: "s-1", SliceNS: 60e9,
		UntilNS: model.DefaultStartTimeNS + 300e9, Policies: []string{hostPolicy}, SafeStop: []string{"safe_stop"}}
	nodeErr := Serve(t.Context(), r, path, 5*time.Second, options)
	result := <-hosted
	result.node, result.run = nodeErr, r
	return result
}

func referenceOptions() HostOptions {
	return HostOptions{Environment: "simulated", ProtocolMax: 1, Epoch: 1, Policy: hostPolicy, Script: setpointScript()}
}

func TestASlowReferenceHostIsNeverOutrunAndVerifiesItsCommand(t *testing.T) {
	t.Parallel()
	result := pair(t, referenceOptions(), 30*time.Millisecond, nil)
	if result.node != nil || result.host != nil {
		t.Fatalf("node %v, host %v", result.node, result.host)
	}
	status := result.report.Commands["ref-1"]
	if status == nil || status.Result != "succeeded" || status.Verified == nil || !*status.Verified {
		t.Fatalf("command status %+v", status)
	}
	for _, pacing := range journalRecords(t, result.run, "pacing") {
		if waited, _ := pacing["waited_ms"].(json.Number).Int64(); pacing["missed"] == true || waited < 30 {
			t.Fatalf("a barrier was outrun: %v", pacing)
		}
	}
}

func TestTheReferenceHostCatchesALyingResult(t *testing.T) {
	t.Parallel()
	result := pair(t, referenceOptions(), 0, func(r *run.Run) { r.SetFailureMode(world.ModeSilentNoEffect) })
	if status := result.report.Commands["ref-1"]; status.Result != "succeeded" || status.Verified == nil || *status.Verified {
		t.Fatalf("a silent no-effect command was not caught: %+v", status)
	}
}

func TestTheReferenceHostRefusesAnotherEnvironmentAndQuarantinesTheUncommissioned(t *testing.T) {
	t.Parallel()
	wrong := referenceOptions()
	wrong.Environment = "hardware"
	if result := pair(t, wrong, 0, nil); !errors.Is(result.node, ErrRefused) || result.report.Refused != "environment_mismatch" {
		t.Fatalf("node %v, report %+v", result.node, result.report)
	}
	partial := referenceOptions()
	partial.Inventory = map[string]string{"P1": "pump"}
	if result := pair(t, partial, 0, nil); len(result.report.Quarantined) != 1 || result.report.Quarantined[0] != "P2" {
		t.Fatalf("quarantined %v", result.report.Quarantined)
	}
}

func TestTwoRunsWithTheSameHostGiveTheSameWire(t *testing.T) {
	t.Parallel()
	wires := func() []byte {
		result := pair(t, referenceOptions(), 0, nil)
		var out bytes.Buffer
		for _, wire := range journalRecords(t, result.run, "wire") {
			raw, _ := json.Marshal(wire)
			out.Write(raw)
		}
		return out.Bytes()
	}
	if first, second := wires(), wires(); len(first) == 0 || !bytes.Equal(first, second) {
		t.Fatalf("the wire records differ (%d and %d bytes)", len(first), len(second))
	}
}

func journalRecords(t *testing.T, r *run.Run, kind string) []map[string]any {
	t.Helper()
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	dir, _ := journalDirs.Load(r)
	raw, err := os.ReadFile(filepath.Join(dir.(string), "journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.UseNumber()
		var record map[string]any
		if err := decoder.Decode(&record); err == nil && record["type"] == kind {
			records = append(records, record)
		}
	}
	return records
}

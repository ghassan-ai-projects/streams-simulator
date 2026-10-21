package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	twinStartNS  = int64(1767225600000000000)
	twinTickNS   = int64(1_000_000_000)
	twinLeaseAt  = twinStartNS + 60*twinTickNS
	twinEndNS    = twinLeaseAt + 600*twinTickNS
	twinLevel    = int64(30)
	twinOpenTemp = 34.0
)

// twinLease is one set_cooling the host sent, with the receipt the board sent back.
type twinLease struct {
	ID, Status, Reason string
	AtNS               int64
}

// twinClosedForm is the thermal bench's own equation: with the fan leased at 30 % (duty 300)
// the zone settles at 22 + 12 - 0.02*300 = 28 degC with a 180 s time constant.
func twinClosedForm(at int64) float64 {
	if at < twinLeaseAt {
		return twinOpenTemp
	}
	seconds := float64(at-twinLeaseAt) / 1e9
	return 28 + (twinOpenTemp-28)*math.Exp(-seconds/180)
}

// twinRenewals is the fan lease renewed every 5 s for the ten minutes after it is first sent.
func twinRenewals(status string) []twinLease {
	var leases []twinLease
	for at := twinLeaseAt; at <= twinEndNS; at += 5 * twinTickNS {
		leases = append(leases, twinLease{ID: fmt.Sprintf("c-%d", len(leases)+1), AtNS: at, Status: status})
	}
	return leases
}

func twinBirths() []map[string]any {
	return []map[string]any{
		{"type": "DBIRTH", "device": "zone-01", "kind": "thermal_sensor", "board": "mega-01", "birth_time_ns": twinStartNS,
			"metrics": []any{
				map[string]any{"name": "temperature", "alias": 1, "value": twinOpenTemp, "quality": "good"},
				map[string]any{"name": "humidity", "alias": 2, "value": 42.0, "quality": "good"},
			}},
		{"type": "DBIRTH", "device": "fan-01", "kind": "fan", "board": "mega-01", "birth_time_ns": twinStartNS,
			"metrics": []any{map[string]any{"name": "level_percent", "alias": 1, "value": 0, "quality": "good"}}},
	}
}

func twinTick(at int64, temperature float64) map[string]any {
	return map[string]any{"type": "DDATA", "device": "zone-01", "historical": false, "metrics": []any{
		map[string]any{"alias": 1, "value": temperature, "quality": "good", "event_time_ns": at}}}
}

func twinReceipts(lease twinLease) []map[string]any {
	if lease.Status == "" {
		return nil
	}
	records := []map[string]any{{"type": "DRECEIPT", "command_id": lease.ID, "status": lease.Status, "reason": lease.Reason}}
	if lease.Status == "accepted" {
		records = append(records, map[string]any{"type": "DRESULT", "command_id": lease.ID, "reported_outcome": "succeeded"})
	}
	return records
}

func twinCommand(t *testing.T, lease twinLease) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", "DCMD.json"))
	if err != nil {
		t.Fatal(err)
	}
	record := twinDecode(t, raw)
	record["command_id"], record["device"], record["command"] = lease.ID, "fan-01", "set_cooling"
	record["args"] = map[string]any{"level_percent": twinLevel}
	return record
}

// twinRecords lays the board's field trace out in time order: each tick's sample,
// then the receipts of the leases sent at that tick, so a receipt is stamped with it.
func twinRecords(t *testing.T, offset func(int64) float64, leases []twinLease) (field, host []map[string]any) {
	t.Helper()
	field = twinBirths()
	for at := twinStartNS; at <= twinEndNS; at += twinTickNS {
		field = append(field, twinTick(at, twinClosedForm(at)+offset(at)))
		for _, lease := range leases {
			if lease.AtNS == at {
				field = append(field, twinReceipts(lease)...)
			}
		}
	}
	for _, lease := range leases {
		host = append(host, twinCommand(t, lease))
	}
	return field, host
}

func noOffset(int64) float64 { return 0 }

// writeTwinFolder writes a run folder the way the relay does: field-trace.jsonl, journal.jsonl
// with the host's commands, and interrupts.jsonl when interrupts are given.
func writeTwinFolder(t *testing.T, field, host []map[string]any, interrupts ...map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	writeFieldTrace(t, dir, field)
	writeHostJournal(t, dir, host)
	if len(interrupts) > 0 {
		writeInterrupts(t, dir, interrupts)
	}
	return dir
}

func writeFieldTrace(t *testing.T, dir string, records []map[string]any) string {
	t.Helper()
	trace, err := run.NewFieldTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	return writeEvidence(t, trace, records)
}

type evidenceLog interface {
	Add(map[string]any) error
	Close() error
	Digest() string
}

func writeEvidence(t *testing.T, log evidenceLog, records []map[string]any) string {
	t.Helper()
	for _, record := range records {
		if err := log.Add(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	return log.Digest()
}

func writeHostJournal(t *testing.T, dir string, host []map[string]any) {
	t.Helper()
	wire, err := run.NewWireTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range host {
		wire.Wire(domain.TraceHost, record)
	}
	if err := wire.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeInterrupts(t *testing.T, dir string, records []map[string]any) string {
	t.Helper()
	log, err := run.NewInterruptLog(dir)
	if err != nil {
		t.Fatal(err)
	}
	return writeEvidence(t, log, records)
}

// twinArtifactFolder adds a run.json that records the field trace's digest, as a
// finished run does; its journal.jsonl is the observer journal of that run.
func twinArtifactFolder(t *testing.T, field []map[string]any, recordTrace bool, interrupts ...map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	spec, err := simdomain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	digest := writeFieldTrace(t, dir, field)
	if recordTrace {
		r.SetFieldTrace(digest)
	}
	if len(interrupts) > 0 {
		r.SetInterruptLog(writeInterrupts(t, dir, interrupts))
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func twinOptions(t *testing.T, dir string) TwinReportOptions {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	domains := t.TempDir()
	testsupport.WriteFileIn(t, domains, "arduino-thermal.domain.json", raw)
	return TwinReportOptions{Run: dir, DomainsDir: domains, Domain: "arduino-thermal", KindsDirs: []string{testsupport.ThermalBoardKindsDir()}}
}

func twinDecode(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var record map[string]any
	if err := decoder.Decode(&record); err != nil {
		t.Fatal(err)
	}
	return record
}

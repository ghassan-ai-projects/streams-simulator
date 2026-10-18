package app

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func recordedHardwareRun(t *testing.T, reading float64) string {
	t.Helper()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	relay := startRelay(t, hostPath, out, nil)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	gateway.handshake()
	gateway.send(gateway.node.Emit(domain.DeviceData("zone-01", "boot-1", false, []map[string]any{domain.Sample(1, reading, "", 1767225600123456789)}))...)
	gateway.sync()
	relay.stop()
	started.stop()
	return out
}

func replayOptions(folder, socket string) ReplayOptions {
	return ReplayOptions{Run: folder, HostSocket: socket, SessionID: "replay-1", BarrierCap: 5 * time.Second}
}

func playTrace(t *testing.T, folder string) (ReplayStats, []map[string]any) {
	t.Helper()
	hostPath := testsupport.SocketPath(t)
	host := serveWireHost(t, hostPath, false)
	stats, err := ReplayTrace(t.Context(), replayOptions(folder, hostPath))
	if err != nil {
		t.Fatal(err)
	}
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NDEATH", "", nil) })
	return stats, host.snapshot()
}

func TestTwoReplaysOfOneFieldTraceGiveTheHostByteIdenticalStreams(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	firstStats, first := playTrace(t, folder)
	secondStats, second := playTrace(t, folder)
	if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstStats, secondStats) {
		t.Fatalf("replays differ:\n%v\n%v", first, second)
	}
	if firstStats.Records != len(first) || firstStats.Barriers != 1 || firstStats.Skipped != 0 || firstStats.Refused != 0 {
		t.Fatalf("stats = %+v over %d records, want every record counted, one barrier of the replay's own and nothing skipped", firstStats, len(first))
	}
}

func TestAReplayPlaysTheRecordedRecordsAsASimulatedNodeNeverAsHardware(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	_, played := playTrace(t, folder)
	trace, _, err := run.ReadFieldTrace(folder)
	if err != nil {
		t.Fatal(err)
	}
	birth := played[0]
	if birth["type"] != "NBIRTH" || birth["environment"] != "simulated" || birth["session_id"] != "replay-1" {
		t.Fatalf("first record = %v, want the replay's own NBIRTH, simulated", birth)
	}
	for _, entry := range birth["boards"].([]any) {
		if board := entry.(map[string]any); board["environment"] != "simulated" {
			t.Errorf("board %v announced as %v, a replay is not a board", board["board"], board["environment"])
		}
	}
	recorded := ofType(trace, "NBIRTH")[0]
	if birth["capability_digest"] != recorded["capability_digest"] || !reflect.DeepEqual(birth["kinds"], recorded["kinds"]) {
		t.Fatal("the replay changed the kinds or their digests the gateway announced")
	}
	for _, kind := range []string{"DBIRTH", "DDATA"} {
		replayed, original := ofType(played, kind), ofType(trace, kind)
		if len(replayed) != len(original) || len(original) == 0 {
			t.Fatalf("%s: replayed %d, recorded %d, want them all in the same order", kind, len(replayed), len(original))
		}
		for i := range original {
			if !reflect.DeepEqual(envelopeFree(replayed[i]), envelopeFree(original[i])) {
				t.Errorf("%s %d replayed as %v, recorded as %v", kind, i, replayed[i], original[i])
			}
		}
	}
}

func TestADifferentFieldTraceGivesADifferentDigestAndADifferentStream(t *testing.T) {
	t.Parallel()
	coldStats, cold := playTrace(t, recordedHardwareRun(t, 5))
	warmStats, warm := playTrace(t, recordedHardwareRun(t, 35))
	if coldStats.FieldTrace == warmStats.FieldTrace || reflect.DeepEqual(cold, warm) {
		t.Fatalf("digests %q and %q, streams equal %v, want two traces to differ in both", coldStats.FieldTrace, warmStats.FieldTrace, reflect.DeepEqual(cold, warm))
	}
}

func TestAReplayWontPassForHardwareToAHostThatAcceptsOnlyHardware(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	hostPath := testsupport.SocketPath(t)
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	_, err := ReplayTrace(t.Context(), replayOptions(folder, hostPath))
	if report := started.stop(); !errors.Is(err, domain.ErrRefused) || report.Refused != "environment_mismatch" {
		t.Fatalf("ReplayTrace = %v, host refused %q, want the host to refuse a node that says it is simulated", err, report.Refused)
	}
}

func TestACommandDuringAReplayIsRefusedAndReachesNothing(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	hostPath := testsupport.SocketPath(t)
	host := serveWireHost(t, hostPath, true)
	done := make(chan error, 1)
	var stats ReplayStats
	go func() {
		var err error
		stats, err = ReplayTrace(t.Context(), replayOptions(folder, hostPath))
		done <- err
	}()
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NSYNC", "", nil) })
	host.write(commandRecord("known", "fan-01", "boot-1", 1), commandRecord("unknown", "ghost-9", "boot-1", 1))
	host.release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NDEATH", "", nil) })
	receipts := map[string]map[string]any{}
	for _, receipt := range ofType(host.snapshot(), "DRECEIPT") {
		receipts[receipt["command_id"].(string)] = receipt
	}
	known, unknown := receipts["known"], receipts["unknown"]
	if stats.Refused != 2 || known["reason"] != "not_ready" || known["boot_id"] != "boot-1" || unknown["reason"] != "unknown_device" || unknown["boot_id"] != "none" {
		t.Fatalf("stats %+v receipts %v, want both commands refused: not_ready for a recorded device, unknown_device for another", stats, receipts)
	}
	if len(ofType(host.snapshot(), "DRESULT")) != 0 {
		t.Fatal("a replay reported the result of a command")
	}
}

func TestAReplayHasNoWayToReachHardware(t *testing.T) {
	t.Parallel()
	var fields []string
	for _, field := range reflect.VisibleFields(reflect.TypeFor[ReplayOptions]()) {
		fields = append(fields, field.Name)
	}
	if want := []string{"Run", "HostSocket", "SessionID", "BarrierCap"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("ReplayOptions fields = %v, want %v: no gateway, no socket to hardware", fields, want)
	}
	for _, file := range []string{"replay_trace.go", "replay_answers.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"Gateway", "gateway", "Relay"} {
			if strings.Contains(string(raw), forbidden) {
				t.Errorf("%s names %q: a replay has no path to a gateway", file, forbidden)
			}
		}
	}
}

func folderWithArtifact(t *testing.T) string {
	t.Helper()
	gateway := relayGateway(t, "s-1")
	birth, err := gateway.Birth()
	if err != nil {
		t.Fatal(err)
	}
	state := map[string]any{"type": "STATE", "lane": "control", "session_id": "h", "protocol_version": 1, "seq": 0, "host_id": "h", "online": true,
		"heartbeat": 0, "heartbeat_interval_ms": 1000, "authority_epoch": 7, "protocol_min": 1, "protocol_max": 1}
	births, err := gateway.Accept(state)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := simdomain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	trace, err := run.NewFieldTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range append([]map[string]any{birth}, births...) {
		if err := trace.Add(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	r.SetFieldTrace(trace.Digest())
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAReplayRefusesATraceThatIsNotTheOneTheArtifactRecorded(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	host := serveWireHost(t, hostPath, false)
	intact := folderWithArtifact(t)
	if stats, err := ReplayTrace(t.Context(), replayOptions(intact, hostPath)); err != nil || stats.Records != 5 {
		t.Fatalf("ReplayTrace of the untouched folder = %+v, %v, want the birth, two device births, a barrier and the death", stats, err)
	}
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NDEATH", "", nil) })
	tampered := folderWithArtifact(t)
	path := filepath.Join(tampered, model.FileFieldTrace)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, tampered, model.FileFieldTrace, []byte(strings.Replace(string(raw), "mega-01", "mega-99", 1)))
	other := testsupport.SocketPath(t)
	untouched := serveWireHost(t, other, false)
	if _, err := ReplayTrace(t.Context(), replayOptions(tampered, other)); !errors.Is(err, run.ErrFieldTraceDigest) {
		t.Fatalf("ReplayTrace of a tampered trace = %v, want ErrFieldTraceDigest", err)
	}
	if got := untouched.snapshot(); len(got) != 0 {
		t.Fatalf("the host received %d records from a replay that was refused, want none: the digest is checked before anything is played", len(got))
	}
}

func TestAReplayRefusesWhatItCannotPlay(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	full := func(folder string) ReplayOptions { return replayOptions(folder, hostPath) }
	cases := map[string]struct {
		options func(folder string) ReplayOptions
		trace   string
		want    error
	}{
		"no run folder": {func(string) ReplayOptions {
			return ReplayOptions{HostSocket: hostPath, SessionID: "r", BarrierCap: time.Second}
		}, "", ErrBadOptions},
		"no host socket": {func(folder string) ReplayOptions {
			return ReplayOptions{Run: folder, SessionID: "r", BarrierCap: time.Second}
		}, "", ErrBadOptions},
		"no barrier cap": {func(folder string) ReplayOptions {
			return ReplayOptions{Run: folder, HostSocket: hostPath, SessionID: "r"}
		}, "", ErrBadOptions},
		"an unsafe session": {func(folder string) ReplayOptions {
			return ReplayOptions{Run: folder, HostSocket: hostPath, SessionID: "a b", BarrierCap: time.Second}
		}, "", ErrBadOptions},
		"a folder with no trace":        {full, "", run.ErrFieldTraceMissing},
		"a trace of junk":               {full, `{"ordinal":0,"record":{"type":"DDATA"}}` + "\n", domain.ErrNoReplayBirth},
		"a trace that is not canonical": {full, `{"ordinal": 0, "record": {}}` + "\n", run.ErrFieldTrace},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			folder := t.TempDir()
			if tc.trace != "" {
				writeTrace(t, folder, tc.trace)
			}
			if _, err := ReplayTrace(t.Context(), tc.options(folder)); !errors.Is(err, tc.want) {
				t.Fatalf("ReplayTrace = %v, want %v", err, tc.want)
			}
		})
	}
}

func writeTrace(t *testing.T, folder, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(folder, model.FileFieldTrace), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

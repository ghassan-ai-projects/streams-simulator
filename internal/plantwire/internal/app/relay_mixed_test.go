package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var mixedPolicy = "sha256:" + strings.Repeat("b", 64)

func thermalKindFile(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.ThermalBoardKindsDir(), name))
	if err != nil {
		t.Fatal(err)
	}
	return relayKindFile(t, string(raw))
}

func mixedGateway(t *testing.T, session string) *standin.Node {
	t.Helper()
	node, err := standin.New(standin.Config{Session: session, Board: "mega-02",
		Kinds: []map[string]any{thermalKindFile(t, "thermal_sensor.kind.json"), thermalKindFile(t, "fan.kind.json")},
		Devices: []standin.Device{{Name: "zone-02", Kind: "thermal_sensor", BootID: "boot-1", Initial: map[string]any{"temperature": 21.5}},
			{Name: "fan-02", Kind: "fan", BootID: "boot-1"}},
		Script: standin.Script{Replies: []standin.Reply{{Command: "set_cooling", Receipt: "accepted", Result: "succeeded"}}}})
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func mixedOptions(t *testing.T, out string) RelayOptions {
	t.Helper()
	domains := t.TempDir()
	raw, err := os.ReadFile(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, domains, "arduino-thermal.domain.json", raw)
	return RelayOptions{Out: out, Simulated: &SimulatedOptions{DomainsDir: domains, AdaptersDir: testsupport.AdaptersDir(),
		Domain: "arduino-thermal", Adapter: "native-jsonl", Seed: 5, BarrierCap: 30 * time.Second,
		Node: Options{Environment: "simulated", ProtocolMax: 1, SessionID: "sim-1", SliceNS: 60e9, UntilNS: model.DefaultStartTimeNS + 180e9,
			Policies: []string{mixedPolicy}, KindsDirs: []string{testsupport.ThermalBoardKindsDir()}}}}
}

func startMixedRelay(t *testing.T, hostPath string, options RelayOptions) startedRelay {
	t.Helper()
	return startMixedRelayTicked(t, hostPath, options, nil)
}

func startMixedRelayTicked(t *testing.T, hostPath string, options RelayOptions, ticks <-chan time.Time) startedRelay {
	t.Helper()
	options.GatewaySocket, options.HostSocket = filepath.Join(filepath.Dir(hostPath), "g.sock"), hostPath
	ctx, cancel := context.WithCancel(t.Context())
	ready, done := make(chan struct{}), make(chan relayRun, 1)
	options.Ready = func() { close(ready) }
	go func() {
		stats, err := serveRelay(ctx, options, ticks)
		done <- relayRun{stats: stats, err: err}
	}()
	select {
	case <-ready:
	case ended := <-done:
		cancel()
		t.Fatalf("the relay did not start: %v", ended.err)
	}
	return startedRelay{gateway: options.GatewaySocket, done: done, cancel: cancel}
}

func commandRecord(id, device, boot string, now int64) domain.Record {
	args := map[string]any{"level_percent": 30}
	key := domain.Command{Device: device, Name: "set_cooling", Args: args, Boot: boot}.Content()
	return domain.Record{Type: "DCMD", Lane: domain.LaneControl, Fields: map[string]any{"command_id": id, "idempotency_key": key, "device": device,
		"command": "set_cooling", "args": args, "expected_boot_id": boot, "deadline_ns": now + 120e9, "not_before_ns": 0, "authority_epoch": 7,
		"policy_digest": mixedPolicy, "fallback_on_lease_loss": "safe_off"}}
}

func lastWorldTime(received []map[string]any) int64 {
	var now int64
	for _, record := range ofType(received, "NSYNC") {
		number, _ := record["world_time_ns"].(json.Number)
		now, _ = number.Int64()
	}
	return now
}

func recordsOf(inbound []domain.Inbound) []map[string]any {
	records := make([]map[string]any, len(inbound))
	for i, in := range inbound {
		records[i] = in.Record
	}
	return records
}

func outcomeOf(received []map[string]any, commandID string) string {
	for _, record := range ofType(received, "DRESULT") {
		if record["command_id"] == commandID {
			outcome, _ := record["reported_outcome"].(string)
			return outcome
		}
	}
	return ""
}

func announcedDevices(received []map[string]any) []string {
	var devices []string
	for _, record := range ofType(received, "DBIRTH") {
		devices = append(devices, record["device"].(string))
	}
	slices.Sort(devices)
	return devices
}

func TestAMixedSessionRoutesEachCommandToTheBoardOfItsDeviceAndRecordsTheFieldTrace(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	host := serveWireHost(t, hostPath, true)
	options := mixedOptions(t, out)
	relay := startMixedRelay(t, hostPath, options)
	gateway := dialGateway(t, relay.gateway, mixedGateway(t, "gw-1"))
	if state := gateway.handshake(); !state.Online || state.AuthorityEpoch != 7 {
		t.Fatalf("STATE = %v, want the host's own for the gateway", state.Record)
	}
	host.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "device", "fan-01") && hasRecord(r, "DBIRTH", "device", "fan-02") && hasRecord(r, "NSYNC", "", nil)
	})
	now := lastWorldTime(host.snapshot())
	host.write(commandRecord("sim-cmd", "fan-01", "mega-01-boot-1", now), commandRecord("hw-cmd", "fan-02", "boot-1", now))
	for _, command := range ofType(recordsOf(gateway.await(ofKind("DCMD"))), "DCMD") {
		gateway.send(gateway.node.Handle(command)...)
	}
	host.release()
	host.waitFor(func(r []map[string]any) bool {
		return outcomeOf(r, "sim-cmd") != "" && outcomeOf(r, "hw-cmd") != "" && hasRecord(r, "DDEATH", "device", "zone-01")
	})
	ended := relay.stop()
	received := host.snapshot()
	assertMixedBirth(t, ofType(received, "NBIRTH")[0])
	assertOneSessionOfTwoOrigins(t, received)
	assertCommandsWentToTheirOwnBoard(t, ended, received)
	assertTheRunKeptTheFieldTrace(t, out, options.Simulated.DomainsDir)
}

func assertMixedBirth(t *testing.T, birth map[string]any) {
	t.Helper()
	if birth["environment"] != "mixed" {
		t.Fatalf("NBIRTH environment = %v, want mixed", birth["environment"])
	}
	boards := map[string]any{}
	for _, entry := range birth["boards"].([]any) {
		board := entry.(map[string]any)
		boards[board["board"].(string)] = board["environment"]
	}
	if len(boards) != 2 || boards["mega-02"] != "hardware" || boards["mega-01"] != "simulated" {
		t.Fatalf("boards = %v, want the gateway's board hardware and the world's simulated", boards)
	}
	for _, entry := range birth["kinds"].([]any) {
		kind := entry.(map[string]any)
		file := map[string]string{"thermal_sensor": "thermal_sensor.kind.json", "fan": "fan.kind.json", "indicator": "indicator.kind.json"}[kind["kind"].(string)]
		digest, err := domain.KindDigest(thermalKindFile(t, file))
		if err != nil || kind["kind_digest"] != digest {
			t.Errorf("kind %v announced as %v, its kind file digests to %s (%v)", kind["kind"], kind["kind_digest"], digest, err)
		}
	}
}

func assertOneSessionOfTwoOrigins(t *testing.T, received []map[string]any) {
	t.Helper()
	for i, record := range received {
		if record["seq"] != json.Number(strconv.Itoa(i)) || record["session_id"] != "relay-h1" {
			t.Fatalf("record %d is %v seq %v of %v, want one session with the relay's own consecutive seq", i, record["type"], record["seq"], record["session_id"])
		}
	}
	want := []string{"fan-01", "fan-02", "led-01", "zone-01", "zone-02"}
	if got := announcedDevices(received); !slices.Equal(got, want) {
		t.Fatalf("devices born on the host session = %v, want %v", got, want)
	}
}

func assertCommandsWentToTheirOwnBoard(t *testing.T, ended relayRun, received []map[string]any) {
	t.Helper()
	if ended.err != nil || ended.stats.Down != 2 || ended.stats.UnknownDevice != 0 || ended.stats.ToSimulated == 0 || ended.stats.FromSimulated == 0 {
		t.Fatalf("relay %+v, want the gateway to have been sent its STATE and its own command and nothing else (2 down)", ended)
	}
	for _, id := range []string{"sim-cmd", "hw-cmd"} {
		if outcome := outcomeOf(received, id); outcome != "succeeded" {
			t.Errorf("command %s: result %q, want succeeded", id, outcome)
		}
	}
	for _, receipt := range ofType(received, "DRECEIPT") {
		boot := map[string]string{"sim-cmd": "mega-01-boot-1", "hw-cmd": "boot-1"}[receipt["command_id"].(string)]
		if receipt["status"] != "accepted" || receipt["boot_id"] != boot {
			t.Errorf("receipt %v, want it accepted by the board of the command's device", receipt)
		}
	}
}

func assertTheRunKeptTheFieldTrace(t *testing.T, out, domains string) {
	t.Helper()
	art, err := run.LoadArtifact(filepath.Join(out, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	invoked := 0
	for _, command := range art.CommandLog {
		if command.Op == model.OpEffectorInvoke {
			invoked++
		}
	}
	trace, digest, err := run.ReadFieldTrace(out)
	if err != nil || invoked != 1 || art.FieldTraceDigest == "" || art.FieldTraceDigest != digest {
		t.Fatalf("artifact %q, trace digest %q (%v), %d effector calls, want one call (the world's only) and the digest of the trace in the artifact", art.FieldTraceDigest, digest, err, invoked)
	}
	var types []string
	for _, record := range trace {
		types = append(types, record["type"].(string))
	}
	if want := []string{"NBIRTH", "DBIRTH", "DBIRTH", "DRECEIPT", "DRESULT"}; !slices.Equal(types, want) {
		t.Fatalf("field trace = %v, want what the gateway sent, in order, and nothing of the world's", types)
	}
	dirs := run.Directories{Domains: domains, Adapters: testsupport.AdaptersDir()}
	if res, err := run.ReplayFile(t.Context(), filepath.Join(out, model.FileRunArtifact), dirs); err != nil || !res.Matches {
		t.Fatalf("replay of the mixed run = %+v, %v, want the world and the field trace to verify", res, err)
	}
}

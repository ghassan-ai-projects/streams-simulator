package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func planText(host, gateway string, binding ...string) string {
	plan := `{"schema_version":"0.1","domain":"arduino-thermal","seed":3,"host_socket":"` + host + `"`
	if gateway != "" {
		plan += `,"gateway_socket":"` + gateway + `"`
	}
	if len(binding) == 3 {
		plan += `,"real":[{"instance":"` + binding[0] + `","board":"` + binding[1] + `","device":"` + binding[2] + `"}]`
	}
	return plan + "}"
}

func connectOptions(t *testing.T, plan, out string) ConnectOptions {
	t.Helper()
	domains := mixedOptions(t, out).Simulated.DomainsDir
	return ConnectOptions{Plan: strings.NewReader(plan), Source: "plan.json", DomainsDir: domains, AdaptersDir: testsupport.AdaptersDir(),
		KindsDirs: []string{testsupport.ThermalBoardKindsDir()}, Policies: []string{mixedPolicy}, Out: out, SliceNS: 60e9, BarrierCap: 30 * time.Second}
}

func TestACheckSaysWhichInstanceIsARealBoardAndWhichIsSimulated(t *testing.T) {
	t.Parallel()
	check, err := CheckConnect(connectOptions(t, planText("/tmp/h.sock", "/tmp/g.sock", "zone-01", "uno-07", "zone-01"), ""))
	if err != nil {
		t.Fatal(err)
	}
	origins := map[string]domain.PlannedInstance{}
	for _, instance := range check.Instances {
		origins[instance.Instance] = instance
	}
	zone := origins["zone-01"]
	if check.Domain != "arduino-thermal" || check.Environment != domain.EnvironmentMixed || zone.Origin != domain.PlannedReal || zone.Board != "uno-07" ||
		zone.KindDigest == "" || origins["fan-01"].Origin != domain.PlannedSimulated || len(check.ImplementsChecked) != 3 {
		t.Fatalf("check = %+v, want zone-01 real on uno-07 with its kind file's digest and the rest simulated", check)
	}
}

func TestACheckRefusesAPlanThatDoesNotHoldOrCannotBeRead(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		options func(ConnectOptions) ConnectOptions
		want    error
	}{
		"an unknown instance": {func(o ConnectOptions) ConnectOptions {
			o.Plan = strings.NewReader(planText("/tmp/h.sock", "/tmp/g.sock", "ghost", "uno-07", "zone-01"))
			return o
		}, ErrPlanRefused},
		"no kind files": {func(o ConnectOptions) ConnectOptions { o.KindsDirs = nil; return o }, domain.ErrKindFileMissing},
		"an unreadable plan": {func(o ConnectOptions) ConnectOptions {
			o.Plan = iotest.ErrReader(errors.New("disk gone"))
			return o
		}, nil},
		"a plan too large": {func(o ConnectOptions) ConnectOptions {
			o.Plan = strings.NewReader(strings.Repeat(" ", maxPlanBytes+1))
			return o
		}, ErrBadOptions},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := CheckConnect(tc.options(connectOptions(t, planText("/tmp/h.sock", ""), "")))
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || (tc.want == nil && !strings.Contains(err.Error(), "disk gone")) {
				t.Fatalf("CheckConnect = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestAPlanViolationNamesTheMemberAndTheRuleForTheCanvas(t *testing.T) {
	t.Parallel()
	_, err := CheckConnect(connectOptions(t, planText("/tmp/h.sock", "", "zone-01", "uno-07", "zone-01"), ""))
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractSessionPlan || invalid.Source != "plan.json" || invalid.Issues[0].Keyword != "gateway_required" {
		t.Fatalf("err = %v, want a session-plan violation from plan.json saying a real board needs a gateway", err)
	}
}

func TestAConnectNeedsARunFolderAndASlice(t *testing.T) {
	t.Parallel()
	options := connectOptions(t, planText("/tmp/h.sock", ""), "")
	if _, err := Connect(t.Context(), options); !errors.Is(err, ErrBadOptions) {
		t.Fatalf("Connect without --out = %v, want ErrBadOptions", err)
	}
}

type connected struct {
	result ConnectResult
	err    error
}

func startConnect(t *testing.T, options ConnectOptions, viaGateway bool) (context.CancelFunc, <-chan connected) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	ready, done := make(chan struct{}), make(chan connected, 1)
	options.Ready = func() { close(ready) }
	go func() {
		result, err := Connect(ctx, options)
		done <- connected{result: result, err: err}
	}()
	if viaGateway {
		select {
		case <-ready:
		case ended := <-done:
			cancel()
			t.Fatalf("the session did not start: %+v", ended)
		}
	}
	return cancel, done
}

func journalOfType(t *testing.T, out, recordType string) []map[string]any {
	t.Helper()
	file, err := os.Open(filepath.Join(out, "journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	var found []map[string]any
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err == nil && record["type"] == recordType {
			found = append(found, record)
		}
	}
	return found
}

func replaysWhole(t *testing.T, out, domains string) {
	t.Helper()
	dirs := run.Directories{Domains: domains, Adapters: testsupport.AdaptersDir()}
	if res, err := run.ReplayFile(t.Context(), filepath.Join(out, model.FileRunArtifact), dirs); err != nil || !res.Matches {
		t.Fatalf("replay of the stopped session = %+v, %v, want an ordinary run folder that replays", res, err)
	}
}

func TestASessionWithoutAGatewayRunsTheWorldUntilStoppedAndSavesARunThatReplays(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, true)
	options := connectOptions(t, planText(hostPath, ""), out)
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "DBIRTH", "device", "fan-01") })
	cancel()
	ended := <-done
	if ended.err != nil || !ended.result.Stopped || ended.result.Out != out || ended.result.Relay != nil {
		t.Fatalf("Connect = %+v, want a stopped node session saved in %s", ended, out)
	}
	origins := map[string]any{}
	for _, origin := range journalOfType(t, out, "origin") {
		origins[origin["instance"].(string)] = origin["environment"]
	}
	if len(origins) != 3 || origins["zone-01"] != "simulated" || origins["fan-01"] != "simulated" {
		t.Fatalf("origins = %v, want the three instances simulated", origins)
	}
	assertLiveRecording(t, out)
	replaysWhole(t, out, options.DomainsDir)
}

func benchGateway(t *testing.T, device string) *standin.Node {
	t.Helper()
	node, err := standin.New(standin.Config{Session: "bench-1", Board: "uno-07",
		Kinds:   []map[string]any{thermalKindFile(t, "thermal_sensor.kind.json")},
		Devices: []standin.Device{{Name: device, Kind: "thermal_sensor", BootID: "61", Initial: map[string]any{"temperature": 24.5}}}})
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func TestARealBoardTakesItsInstancesPlaceAndEveryBirthIsJournalledWithItsOrigin(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-2")
	gatewayPath := filepath.Join(filepath.Dir(hostPath), "g.sock")
	host := serveWireHost(t, hostPath, true)
	options := connectOptions(t, planText(hostPath, gatewayPath, "zone-01", "uno-07", "zone-01"), out)
	cancel, done := startConnect(t, options, true)
	gateway := dialGateway(t, gatewayPath, benchGateway(t, "zone-01"))
	if state := gateway.handshake(); !state.Online {
		t.Fatalf("STATE = %v, want the gateway admitted against the plan", state.Record)
	}
	host.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "board", "uno-07") && hasRecord(r, "DBIRTH", "device", "fan-01")
	})
	cancel()
	if ended := <-done; ended.err != nil || !ended.result.Stopped || ended.result.Relay == nil {
		t.Fatalf("Connect = %+v, want a stopped mixed session", ended)
	}
	assertOnlyTheBoardAnnouncesTheRealInstance(t, host.snapshot())
	assertRealOrigin(t, out)
	assertLiveRecording(t, out)
	replaysWhole(t, out, options.DomainsDir)
}

func assertOnlyTheBoardAnnouncesTheRealInstance(t *testing.T, received []map[string]any) {
	t.Helper()
	for _, birth := range ofType(received, "DBIRTH") {
		if birth["device"] == "zone-01" && birth["board"] != "uno-07" {
			t.Fatalf("the world announced zone-01 on %v, want the real board alone to announce it", birth["board"])
		}
	}
}

func assertRealOrigin(t *testing.T, out string) {
	t.Helper()
	bindings := journalOfType(t, out, "binding")
	if len(bindings) != 1 || bindings[0]["outcome"] != "matched" {
		t.Fatalf("bindings = %v, want the one gateway matched", bindings)
	}
	environments := map[string]any{}
	for _, origin := range journalOfType(t, out, "origin") {
		instance, _ := origin["instance"].(string)
		environments[instance] = origin["environment"]
		if instance == "zone-01" && (origin["board"] != "uno-07" || origin["boot_id"] != "61") {
			t.Fatalf("origin of zone-01 = %v, want the real board uno-07, boot 61", origin)
		}
	}
	if environments["zone-01"] != "hardware" || environments["fan-01"] != "simulated" || environments["led-01"] != "simulated" {
		t.Fatalf("origins = %v, want zone-01 from the gateway's hardware board and the rest from the world", environments)
	}
}

func TestAGatewayThatDoesNotHoldTheBoundDeviceIsRefusedAndTheRefusalIsJournalled(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-3")
	gatewayPath := filepath.Join(filepath.Dir(hostPath), "g.sock")
	serveWireHost(t, hostPath, true)
	cancel, done := startConnect(t, connectOptions(t, planText(hostPath, gatewayPath, "zone-01", "uno-07", "zone-01"), out), true)
	gateway := dialGateway(t, gatewayPath, benchGateway(t, "zone-09"))
	if state := gateway.handshake(); state.Online || state.Refusal != "inventory_mismatch" {
		t.Fatalf("STATE = %v, want inventory_mismatch", state.Record)
	}
	cancel()
	if ended := <-done; ended.err != nil || ended.result.Relay.Refused != "inventory_mismatch" {
		t.Fatalf("Connect = %+v, want the refusal in the relay's report", ended)
	}
	bindings := journalOfType(t, out, "binding")
	issues, _ := bindings[0]["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	if len(bindings) != 1 || bindings[0]["outcome"] != "refused" || issue["keyword"] != "device_missing" || issue["pointer"] != "/real/0/device" {
		t.Fatalf("bindings = %v, want a refused binding with device_missing at /real/0/device", bindings)
	}
}

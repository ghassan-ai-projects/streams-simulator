package app

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	modeldomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const stationPump = "site-a/pump-1"

func stationPlan(domainID, hostSocket, disturbances string) string {
	return `{"schema_version":"0.1","domain":"` + domainID + `","seed":5,"host_socket":"` + hostSocket + `","policies":["` + mixedPolicy + `"]` + disturbances + `}`
}

func stationOptions(plan, out string) ConnectOptions {
	return ConnectOptions{Plan: strings.NewReader(plan), Source: "plan.json", DomainsDir: testsupport.PumpStationDir(),
		AdaptersDir: testsupport.AdaptersDir(), Out: out, SliceNS: 60e9, BarrierCap: 30 * time.Second}
}

func stationCommand(id, boot, name string, args map[string]any, now int64) domain.Record {
	device := domain.WireID(stationPump)
	key := domain.Command{Device: device, Name: name, Args: args, Boot: boot}.Content()
	return domain.Record{Type: "DCMD", Lane: domain.LaneControl, Fields: map[string]any{"command_id": id, "idempotency_key": key, "device": device,
		"command": name, "args": args, "expected_boot_id": boot, "deadline_ns": now + 120e9, "not_before_ns": 0, "authority_epoch": 7,
		"policy_digest": mixedPolicy, "fallback_on_lease_loss": "safe_off"}}
}

type stationSession struct {
	out      string
	received []map[string]any
}

func runStationSession(t *testing.T, domainID, disturbances, command string, args map[string]any, hours float64) stationSession {
	t.Helper()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-station")
	host := serveWireHost(t, hostPath, true)
	cancel, done := startConnect(t, stationOptions(stationPlan(domainID, hostPath, disturbances), out), false)
	device := domain.WireID(stationPump)
	host.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "device", device) && hasRecord(r, "NSYNC", "", nil)
	})
	received := host.snapshot()
	if command != "" {
		host.write(stationCommand("cmd-1", bootOf(received, device), command, args, lastWorldTime(received)))
	}
	host.release()
	until := model.DefaultStartTimeNS + int64(hours*3600e9)
	host.waitFor(func(r []map[string]any) bool { return lastWorldTime(r) >= until })
	cancel()
	if ended := <-done; ended.err != nil || !ended.result.Stopped {
		t.Fatalf("Connect = %+v, want a stopped session", ended)
	}
	return stationSession{out: out, received: host.snapshot()}
}

func pumpKindDigestOf(t *testing.T, domainFile string) string {
	t.Helper()
	spec, err := modeldomain.Load(domainFile)
	if err != nil {
		t.Fatal(err)
	}
	kinds := make([]*model.Kind, len(spec.Spec.Kinds))
	for i := range spec.Spec.Kinds {
		kinds[i] = &spec.Spec.Kinds[i]
	}
	templates, err := domain.WithKindDigests(domain.KindTemplates(kinds))
	if err != nil {
		t.Fatal(err)
	}
	return templates[0].(map[string]any)["kind_digest"].(string)
}

func TestAMaintenanceCommandNamingTheFaultReachesTheWorldAndTheResponderClearsIt(t *testing.T) {
	t.Parallel()
	const wear = `,"disturbances":[{"at_s":120,"instance":"` + stationPump + `","fault":"bearing_wear"}]`
	session := runStationSession(t, "pump-station", wear, "schedule_maintenance", map[string]any{"pump_id": stationPump, "work_order": "bearing_wear"}, 2)
	if status, reason := receiptOf(session.received, "cmd-1"); status != "accepted" || outcomeOf(session.received, "cmd-1") != "succeeded" {
		t.Fatalf("command receipt %q (%s) result %q, want an executed maintenance", status, reason, outcomeOf(session.received, "cmd-1"))
	}
	if effectorCalls(t, session.out) != 1 {
		t.Fatalf("effector calls = %d, want the one maintenance", effectorCalls(t, session.out))
	}
	if !fileHolds(t, session.out, model.FileDirectorJournal, `"cleared_by":"responder"`) {
		t.Fatal("the sealed journal does not show the wear cleared by the responder")
	}
	replaysWhole(t, session.out, testsupport.PumpStationDir())
}

func TestTheNodeBirthCarriesThePumpKindAndItsDigest(t *testing.T) {
	t.Parallel()
	for domainID, file := range map[string]string{"pump-station": testsupport.PumpStationDomain(), "pump-station-probe-b": testsupport.PumpStationProbeDomain()} {
		t.Run(domainID, func(t *testing.T) {
			t.Parallel()
			session := runStationSession(t, domainID, "", "", nil, 0.05)
			births := ofType(session.received, "NBIRTH")
			if len(births) != 1 {
				t.Fatalf("%d node births, want one", len(births))
			}
			kinds, _ := births[0]["kinds"].([]any)
			pump, _ := kinds[0].(map[string]any)
			if pump["kind"] != "pump" || pump["kind_digest"] != pumpKindDigestOf(t, file) {
				t.Fatalf("born kind = %v %v, want pump with digest %s", pump["kind"], pump["kind_digest"], pumpKindDigestOf(t, file))
			}
			if !hasRecord(session.received, "DBIRTH", "device", domain.WireID(stationPump)) {
				t.Fatalf("the device %s was not born", domain.WireID(stationPump))
			}
		})
	}
}

func pumpLoads(t *testing.T, out string) []float64 {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, model.FileHistory))
	if err != nil {
		t.Fatal(err)
	}
	var loads []float64
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var snapshot model.StateSnapshot
		if err := json.Unmarshal([]byte(line), &snapshot); err != nil {
			t.Fatal(err)
		}
		if load, ok := snapshot.States["pump_load"]; ok && snapshot.Entity == stationPump {
			loads = append(loads, load)
		}
	}
	return loads
}

func TestAReduceLoadCommandMovesThePumpLoadDown(t *testing.T) {
	t.Parallel()
	session := runStationSession(t, "pump-station", "", "reduce_load", map[string]any{"pump_id": stationPump, "target_load": 0.6}, 1)
	if status, reason := receiptOf(session.received, "cmd-1"); status != "accepted" || outcomeOf(session.received, "cmd-1") != "succeeded" {
		t.Fatalf("command receipt %q (%s), want an executed reduce_load", status, reason)
	}
	loads := pumpLoads(t, session.out)
	if len(loads) < 2 || math.Abs(loads[0]-0.8) > 1e-9 || math.Abs(loads[len(loads)-1]-0.6) > 0.01 {
		t.Fatalf("pump load history of %d snapshots from %v to %v, want it from 0.8 down to 0.6", len(loads), loads[0], loads[len(loads)-1])
	}
}

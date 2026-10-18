package app

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAStandInGatewayDrivesACommandThroughTheRelayToTheReferenceHostOverSockets(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	started := startHost(t, hostPath, relayHostOptions("hardware", relayCommandStep()))
	relay := startRelay(t, hostPath, out, nil)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	if state := gateway.handshake(); !state.Online || state.AuthorityEpoch != 7 {
		t.Fatalf("STATE = %v, want the host's own, online, with epoch 7", state.Record)
	}
	answered := gateway.sync()
	if types := []string{answered[0].Type, answered[1].Type}; !reflect.DeepEqual(types, []string{"DCMD", "HSYNC"}) {
		t.Fatalf("the barrier brought %v, want the host's command and then its answer", types)
	}
	gateway.sync()
	ended, report := relay.stop(), started.stop()
	status := report.Commands["cmd-1"]
	if ended.err != nil || report.Refused != "" || status == nil || !reflect.DeepEqual(status.Receipts, []string{"accepted"}) || status.Result != "succeeded" {
		t.Fatalf("relay %+v, host report %+v, want the command accepted and succeeded through the relay", ended, report)
	}
	if stats := ended.stats; stats.GatewaySessions != 1 || stats.HostSessions != 1 || stats.Dropped != 0 {
		t.Fatalf("stats = %+v, want one gateway session, one host session and nothing dropped", ended.stats)
	}
}

func TestTheRelayJournalHoldsEveryRelayedRecordAsReceivedAndAsSent(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	started := startHost(t, hostPath, relayHostOptions("hardware", relayCommandStep()))
	relay := startRelay(t, hostPath, out, nil)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	gateway.handshake()
	gateway.sync()
	gateway.sync()
	relay.stop()
	started.stop()
	lines := wireLines(t, out)
	assertCarriedOnce(t, lines, "gateway", "node", 1)
	assertCarriedOnce(t, lines, "host", "relay", 0)
	first, second := lines[0], lines[1]
	if first["direction"] != "gateway" || recordOf(first)["type"] != "NBIRTH" || recordOf(first)["session_id"] != "s-1" ||
		second["direction"] != "node" || recordOf(second)["type"] != "NBIRTH" || recordOf(second)["session_id"] != "relay-h1" {
		t.Fatalf("first lines = %v and %v, want the gateway's NBIRTH and then the relay's own", first, second)
	}
	if first["stream"] != "observer" || first["type"] != "wire" {
		t.Fatalf("line = %v, want an observer wire session record of the run journal", first)
	}
}

func TestKindsCapturedFromTheRelayJournalCarryTheDigestsTheStandInAnnounced(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	relay := startRelay(t, hostPath, out, nil)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	gateway.handshake()
	gateway.sync()
	relay.stop()
	started.stop()
	captured, err := CaptureFile(filepath.Join(out, "journal.jsonl"), nil)
	if err != nil || len(captured.Kinds) != 2 || captured.Environment != "hardware" {
		t.Fatalf("CaptureFile = %+v, %v, want the two kinds of the hardware board", captured, err)
	}
	files := map[string]string{"thermal_sensor": relaySensorKind, "fan": relayFanKind}
	for _, kind := range captured.Kinds {
		digest, err := domain.KindDigest(relayKindFile(t, files[kind.Kind]))
		if err != nil || kind.KindDigest != digest {
			t.Errorf("kind %s captured with digest %s, its kind file digests to %s (%v)", kind.Kind, kind.KindDigest, digest, err)
		}
	}
}

func TestTheRelayKeepsTheGatewaySessionWhenTheHostDiesAndRebirthsItOnAnotherHost(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	first := startHost(t, hostPath, relayHostOptions("hardware"))
	ticks := make(chan time.Time)
	relay := startRelay(t, hostPath, "", ticks)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	gateway.handshake()
	gateway.sync()
	first.stop()
	offline := gateway.await(func(m domain.Message) bool { return m.Type == "STATE" && !m.Online })
	if last := offline[len(offline)-1]; last.AuthorityEpoch != 7 || last.Refusal != "" {
		t.Fatalf("STATE = %v, want the real host's epoch 7, offline, with no refusal", last.Record)
	}
	tick(t, ticks)
	tick(t, ticks)
	second := startHost(t, hostPath, relayHostOptions("hardware"))
	tick(t, ticks)
	asked := gateway.await(ofKind("NCMD"))
	gateway.send(mustBirth(t, gateway))
	gateway.accept(lastState(t, asked))
	gateway.sync()
	ended, report := relay.stop(), second.stop()
	if ended.err != nil || ended.stats.HostSessions != 2 || ended.stats.GatewaySessions != 1 || report.Refused != "" || len(report.Quarantined) != 0 {
		t.Fatalf("relay %+v, new host report %+v, want the second host session born from the same gateway session", ended, report)
	}
}

func mustBirth(t *testing.T, gateway *gatewayDriver) map[string]any {
	t.Helper()
	birth, err := gateway.node.Birth()
	if err != nil {
		t.Fatal(err)
	}
	return birth
}

func TestTheRelayKillsTheDevicesOfAGatewayThatDisconnectsAndKeepsTheHostSession(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	relay := startRelay(t, hostPath, out, nil)
	first := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	first.handshake()
	first.sync()
	if err := first.link.Close(); err != nil {
		t.Fatal(err)
	}
	second := dialGateway(t, relay.gateway, relayGateway(t, "s-2"))
	if state := second.handshake(); !state.Online || state.AuthorityEpoch != 7 {
		t.Fatalf("STATE = %v, want the host's own state given to the returning gateway", state.Record)
	}
	second.sync()
	ended, report := relay.stop(), started.stop()
	if ended.err != nil || ended.stats.GatewaySessions != 2 || ended.stats.HostSessions != 1 || report.Refused != "" || len(report.Quarantined) != 0 {
		t.Fatalf("relay %+v, host report %+v, want two gateway sessions on one host session", ended, report)
	}
	var kinds []string
	for _, line := range wireLines(t, out) {
		if record := recordOf(line); line["direction"] == "node" && record["type"] == "DDEATH" {
			kinds = append(kinds, record["device"].(string)+":"+record["reason"].(string))
		}
	}
	if want := []string{"fan-01:link_down", "zone-01:link_down"}; !reflect.DeepEqual(kinds, want) {
		t.Fatalf("DDEATH up = %v, want a link_down for each device of the lost gateway", kinds)
	}
}

func TestAGatewayThatSendsWhatIsTooLargeIsDroppedOrDisconnectedWhileTheHostSessionLives(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	relay := startRelay(t, hostPath, "", nil)
	first := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	first.handshake()
	oversize := first.node.Emit(domain.DeviceData("zone-01", "boot-1", false, []map[string]any{domain.Sample(1, 1.0, "", 1)}))[0]
	oversize["padding"] = strings.Repeat("x", 70<<10)
	first.send(oversize)
	first.sync()
	flood := map[string]any{"type": "DDATA", "padding": strings.Repeat("y", 2<<20)}
	_ = first.link.Send([]map[string]any{flood})
	second := dialGateway(t, relay.gateway, relayGateway(t, "s-2"))
	if state := second.handshake(); !state.Online {
		t.Fatalf("STATE = %v, want the host session to answer the next gateway", state.Record)
	}
	second.sync()
	ended, report := relay.stop(), started.stop()
	if ended.err != nil || ended.stats.Dropped != 1 || ended.stats.GatewaySessions != 2 || report.Refused != "" || len(report.Quarantined) != 0 {
		t.Fatalf("relay %+v, host report %+v, want the 70 KiB record dropped, the 2 MiB line to have cost its gateway only and no refusal", ended, report)
	}
}

func envelopeFree(record map[string]any) map[string]any {
	content := map[string]any{}
	for member, value := range record {
		switch member {
		case "type", "lane", "session_id", "protocol_version", "seq", "bd_seq", "prior_last_seq":
		default:
			content[member] = value
		}
	}
	content["type"] = record["type"]
	return content
}

func assertCarriedOnce(t *testing.T, lines []map[string]any, received, sent string, originated int) {
	t.Helper()
	carried, own, next := 0, 0, 0.0
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		switch line["direction"] {
		case received:
			carried++
			if i+1 >= len(lines) || lines[i+1]["direction"] != sent || !reflect.DeepEqual(envelopeFree(recordOf(lines[i+1])), envelopeFree(recordOf(line))) {
				t.Fatalf("%s line %d (%v) is not followed by the %s line carrying the same content", received, i, recordOf(line)["type"], sent)
			}
		case sent:
			if seq, _ := recordOf(line)["seq"].(float64); seq != next {
				t.Fatalf("%s line %d has seq %v, want the relay's own consecutive %v", sent, i, recordOf(line)["seq"], next)
			}
			next++
			own++
		}
	}
	if own != carried+originated {
		raw, _ := json.Marshal(lines)
		t.Fatalf("%s lines %d, %s lines %d, want every record carried once and %d the relay sent itself: %s", received, carried, sent, own, originated, raw)
	}
}

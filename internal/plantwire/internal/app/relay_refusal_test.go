package app

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheRelayOffersNoOptionToAddAFaultAPerturbationOrAScript(t *testing.T) {
	t.Parallel()
	var fields []string
	for _, field := range reflect.VisibleFields(reflect.TypeFor[RelayOptions]()) {
		fields = append(fields, field.Name)
	}
	if want := []string{"GatewaySocket", "HostSocket", "Out", "Simulated", "Ready", "Control"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("RelayOptions fields = %v, want %v: no perturbation, wire fault or script may act on a hardware device's records; the operator control only refuses or stops", fields, want)
	}
}

func TestTheOnlyRelayOptionThatShapesRecordsActsOnTheSimulatedWorldAlone(t *testing.T) {
	t.Parallel()
	var fields []string
	for _, field := range reflect.VisibleFields(reflect.TypeFor[SimulatedOptions]()) {
		fields = append(fields, field.Name)
	}
	if want := []string{"DomainsDir", "AdaptersDir", "Domain", "Adapter", "Seed", "BarrierCap", "Node", "Session"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("SimulatedOptions fields = %v, want %v: the node options inside it (wire faults, policies) reach the simulated session only", fields, want)
	}
}

func TestARelayNeedsTwoDifferentSockets(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	for name, options := range map[string]RelayOptions{"none": {}, "no host": {GatewaySocket: path}, "no gateway": {HostSocket: path},
		"one socket for both": {GatewaySocket: path, HostSocket: path}} {
		if _, err := serveRelay(t.Context(), options, nil); !errors.Is(err, ErrBadOptions) {
			t.Errorf("%s: serveRelay = %v, want ErrBadOptions", name, err)
		}
	}
}

func TestARelayStopsWithItsContextAndReportsWhatItCarried(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	ctx, cancel := context.WithCancel(t.Context())
	stats, err := Relay(ctx, RelayOptions{GatewaySocket: filepath.Join(filepath.Dir(hostPath), "g.sock"), HostSocket: hostPath, Ready: cancel})
	if err != nil || stats != (domain.RelayStats{}) {
		t.Fatalf("Relay = %+v, %v, want an empty report and no error", stats, err)
	}
}

func TestARelayRefusesAGatewaySocketThatIsInUseOrNotPrivate(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	hostPath := path + ".host"
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	options := RelayOptions{GatewaySocket: path, HostSocket: hostPath, Out: t.TempDir()}
	if _, err := serveRelay(t.Context(), options, nil); !errors.Is(err, uds.ErrSocketInUse) {
		t.Fatalf("serveRelay on a live socket = %v, want ErrSocketInUse", err)
	}
	if _, err := os.Stat(filepath.Join(options.Out, "journal.jsonl")); err != nil {
		t.Fatalf("the journal of the refused relay = %v, want it opened and closed", err)
	}
	private := testsupport.SocketPath(t)
	chmodIn(t, filepath.Dir(private), 0o755)
	if _, err := serveRelay(t.Context(), RelayOptions{GatewaySocket: private, HostSocket: hostPath}, nil); !errors.Is(err, uds.ErrSocketNotPrivate) {
		t.Fatalf("serveRelay in a group-readable directory = %v, want ErrSocketNotPrivate", err)
	}
}

func TestARelayRefusesAJournalFolderItCannotCreate(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "plain")
	testsupport.WriteFileIn(t, filepath.Dir(file), "plain", nil)
	hostPath := testsupport.SocketPath(t)
	_, err := serveRelay(t.Context(), RelayOptions{GatewaySocket: hostPath + ".g", HostSocket: hostPath, Out: filepath.Join(file, "run")}, nil)
	var pathError *fs.PathError
	if !errors.As(err, &pathError) {
		t.Fatalf("serveRelay = %v, want the file system's refusal to create a folder under a file", err)
	}
}

func TestARelayDoesNotDialAHostSocketThatOthersCanReachAndSaysSoToTheGateway(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", hostPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	chmodIn(t, hostPath, 0o777)
	relay := startRelay(t, hostPath, "", nil)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	state := gateway.handshake()
	ended := <-relay.done
	if state.Online || state.Refusal != "" || !errors.Is(ended.err, uds.ErrSocketNotPrivate) {
		t.Fatalf("STATE online %v refusal %q, relay %v, want an offline STATE and the relay ended on ErrSocketNotPrivate", state.Online, state.Refusal, ended.err)
	}
}

func TestAHostThatRefusesTheRelayedSessionEndsTheRelayAndTheGatewayHearsWhy(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	started := startHost(t, hostPath, relayHostOptions("simulated"))
	relay := startRelay(t, hostPath, "", nil)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	state := gateway.handshake()
	ended, report := <-relay.done, started.stop()
	if state.Online || state.Refusal != "environment_mismatch" || !errors.Is(ended.err, domain.ErrRefused) || ended.stats.HostRefused != "environment_mismatch" {
		t.Fatalf("STATE %v, relay %+v, want the host's own refusal handed to the gateway and the relay ended on it", state.Record, ended)
	}
	if report.Refused != "environment_mismatch" {
		t.Fatalf("host report %+v, want it to have refused the relayed node", report)
	}
}

func TestARelayRefusesAGatewayThatClaimsToBeSimulatedAndWaitsForAnother(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	relay := startRelay(t, hostPath, "", nil)
	config := standin.Config{Session: "s-1", Board: "mega-01", Environment: "simulated", Kinds: []map[string]any{relayKindFile(t, relaySensorKind)},
		Devices: []standin.Device{{Name: "zone-01", Kind: "thermal_sensor", BootID: "boot-1"}}}
	liar, err := standin.New(config)
	if err != nil {
		t.Fatal(err)
	}
	state := dialGateway(t, relay.gateway, liar).handshake()
	if state.Online || state.Refusal != "environment_mismatch" {
		t.Fatalf("STATE %v, want the gateway refused as a host would", state.Record)
	}
	honest := dialGateway(t, relay.gateway, relayGateway(t, "s-2")).handshake()
	ended := relay.stop()
	if honest.Online || honest.Refusal != "" || ended.err != nil || ended.stats.Refused != "environment_mismatch" || ended.stats.Reason == "" || ended.stats.GatewaySessions != 2 {
		t.Fatalf("second STATE %v, relay %+v, want the second gateway told the host is offline and the first refusal reported with its reason", honest.Record, ended)
	}
}

func TestASendThatFailsLosesTheLinkInsteadOfBeingIgnored(t *testing.T) {
	t.Parallel()
	client, server := net.Pipe()
	link := uds.NewAccepted(client, time.Second)
	t.Cleanup(func() { _ = link.Close() })
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	lost := 0
	records := []map[string]any{{"type": "DDATA"}}
	if err := deliver(link, records, "host", func() error { lost++; return nil }); err != nil || lost != 1 {
		t.Fatalf("deliver on a dead link = %v with %d losses, want the loss handled once", err, lost)
	}
	if err := deliver(nil, records, "host", func() error { lost++; return nil }); err != nil || lost != 1 {
		t.Fatalf("deliver with no link = %v with %d losses, want nothing", err, lost)
	}
	if err := deliver(link, nil, "host", func() error { lost++; return nil }); err != nil || lost != 1 {
		t.Fatalf("deliver of nothing = %v with %d losses, want nothing", err, lost)
	}
}

func chmodIn(t *testing.T, path string, mode fs.FileMode) {
	t.Helper()
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	if err := root.Chmod(filepath.Base(path), mode); err != nil {
		t.Fatal(err)
	}
}

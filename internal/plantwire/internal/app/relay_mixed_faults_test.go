package app

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheFaultsOfTheSimulatedWorldNeverTouchWhatTheGatewaySent(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	host := serveWireHost(t, hostPath, false)
	options := mixedOptions(t, out)
	options.Simulated.Node.WireFaults = []domain.WireFault{{Seq: 1, Kind: domain.FaultGarble}, {Seq: 2, Kind: domain.FaultDrop}, {Seq: 3, Kind: domain.FaultGarble}}
	relay := startMixedRelay(t, hostPath, options)
	gateway := dialGateway(t, relay.gateway, mixedGateway(t, "gw-1"))
	gateway.handshake()
	host.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "device", "fan-02") && hasRecord(r, "NSYNC", "", nil)
	})
	ended := relay.stop()
	trace, _, err := run.ReadFieldTrace(out)
	if err != nil {
		t.Fatal(err)
	}
	carried := ofType(host.snapshot(), "DBIRTH")
	for _, sent := range ofType(trace, "DBIRTH") {
		if !slices.ContainsFunc(carried, func(record map[string]any) bool { return sameDeviceBirth(record, sent) }) {
			t.Errorf("the gateway's %v did not reach the host as it was sent", sent["device"])
		}
	}
	if ended.err != nil || ended.stats.Dropped == 0 {
		t.Fatalf("relay %+v, want the world's garbled records dropped and counted while the gateway's passed", ended)
	}
}

func sameDeviceBirth(carried, sent map[string]any) bool {
	for _, member := range []string{"device", "kind", "board", "boot_id", "birth_time_ns", "metrics"} {
		a, _ := json.Marshal(carried[member])
		b, _ := json.Marshal(sent[member])
		if string(a) != string(b) {
			return false
		}
	}
	return true
}

func TestAMixedRelayRefusesWhatItCannotBuildBeforeItListens(t *testing.T) {
	t.Parallel()
	good := func(t *testing.T) RelayOptions {
		options := mixedOptions(t, t.TempDir())
		options.GatewaySocket, options.HostSocket = testsupport.SocketPath(t), testsupport.SocketPath(t)+".host"
		return options
	}
	cases := map[string]struct {
		change func(*RelayOptions)
		want   error
	}{
		"no run folder":      {func(o *RelayOptions) { o.Out = "" }, ErrBadOptions},
		"an unknown domain":  {func(o *RelayOptions) { o.Simulated.Domain = "no-such-domain" }, nil},
		"no slice":           {func(o *RelayOptions) { o.Simulated.Node.SliceNS = 0 }, ErrBadOptions},
		"a hardware world":   {func(o *RelayOptions) { o.Simulated.Node.Environment = "hardware" }, ErrBadOptions},
		"kind files missing": {func(o *RelayOptions) { o.Simulated.Node.KindsDirs = nil }, domain.ErrKindFileMissing},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options := good(t)
			tc.change(&options)
			_, err := serveRelay(t.Context(), options, nil)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("serveRelay = %v, want a refusal %v before anything listens", err, tc.want)
			}
			if _, statErr := os.Stat(options.GatewaySocket); statErr == nil {
				t.Fatal("the gateway socket was published although the relay could not be built")
			}
		})
	}
}

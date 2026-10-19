package app

import (
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func floodRecord() domain.Record {
	samples := make([]map[string]any, 500)
	for i := range samples {
		samples[i] = domain.Sample(1, 21.5, "good", twinStartNS+int64(i))
	}
	return domain.DeviceData("zone-01", "boot-1", false, samples)
}

func TestAHostThatStopsReadingDoesNotHoldTheEmergencyStopBackForLongerThanOneBoundedWrite(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	wireHost := serveWireHost(t, hostPath, false)
	port, ticks := NewControlPort(), make(chan time.Time)
	relay := startMixedRelayTicked(t, hostPath, RelayOptions{Out: out, Control: port}, ticks)
	node := relayGateway(t, "gw-1")
	gateway := dialGateway(t, relay.gateway, node)
	gateway.handshake()
	wireHost.waitFor(func(r []map[string]any) bool { return hasRecord(r, "DBIRTH", "device", "fan-01") })
	wireHost.stall()
	for range 40 {
		gateway.send(node.Emit(floodRecord())...)
	}
	started := time.Now()
	controlThroughTicks(t, port, ticks, `{"op":"estop","scope":{"kind":"device","id":"fan-01"}}`)
	stop := ofType(recordsOf(gateway.await(ofKind("DCMD"))), "DCMD")[0]
	if waited := time.Since(started); stop["device"] != "fan-01" || stop["command"] != "stop_fan" || stop["expected_boot_id"] != "boot-1" || waited >= 4*time.Second {
		t.Fatalf("stop %v reached the gateway after %v, want the stop within about one 2 s bounded host write (limit 4 s), not the 30 s a stalled host once cost", stop, waited)
	}
	relay.stop()
}

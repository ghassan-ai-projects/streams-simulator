package app

import (
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAGatewaySilentForLongerThanItsWatchdogDuringAHostOutageIsKeptAndRebornWhenTheHostReturns(t *testing.T) {
	t.Parallel()
	hostPath, ticks := testsupport.SocketPath(t), make(chan time.Time)
	first := startHost(t, hostPath, relayHostOptions("hardware"))
	relay := startRelay(t, hostPath, "", ticks)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	gateway.handshake()
	gateway.sync()
	first.stop()
	gateway.await(func(m domain.Message) bool { return m.Type == "STATE" && !m.Online })
	for _, second := range []int64{0, 31, 62} {
		tickAt(t, ticks, second)
	}
	second := startHost(t, hostPath, relayHostOptions("hardware"))
	tickAt(t, ticks, 63)
	asked := gateway.await(ofKind("NCMD"))
	gateway.send(mustBirth(t, gateway))
	gateway.accept(lastState(t, asked))
	gateway.sync()
	ended, report := relay.stop(), second.stop()
	if ended.err != nil || ended.stats.GatewaySessions != 1 || ended.stats.HostSessions != 2 || report.Refused != "" || len(report.Quarantined) != 0 {
		t.Fatalf("relay %+v, new host report %+v, want the silent gateway kept through a 62 s outage and reborn on the new host", ended, report)
	}
}

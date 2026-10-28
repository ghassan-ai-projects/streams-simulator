package uds

import (
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const beatMS = 100

func heartbeat(t *testing.T, seq, beat int) map[string]any {
	t.Helper()
	return map[string]any{"type": "STATE", "protocol_version": 1, "session_id": "s-1", "seq": seq, "lane": "control", "host_id": "host-a", "online": true,
		"heartbeat": beat, "heartbeat_interval_ms": beatMS, "authority_epoch": 1, "protocol_min": 1, "protocol_max": 1}
}

func inboundHeartbeat(t *testing.T, seq, beat int) domain.Inbound {
	t.Helper()
	raw, err := json.Marshal(heartbeat(t, seq, beat))
	if err != nil {
		t.Fatal(err)
	}
	in, ok := domain.Decode(raw)
	if !ok || in.Invalid != nil {
		t.Fatalf("heartbeat %s is no valid record: %v", raw, in.Invalid)
	}
	return in
}

func TestASimulatedNodeLosesAHostThatStopsHeartbeatingAfterThreeIntervals(t *testing.T) {
	t.Parallel()
	link := NewSimLink(time.Minute)
	defer link.Close()
	link.Deliver(inboundHeartbeat(t, 1, 1))
	started := time.Now()
	if _, met, err := link.Await(t.Context(), isKind("HSYNC")); met || !errors.Is(err, ErrHostSilent) {
		t.Fatalf("Await = %v, %v, want %v", met, err, ErrHostSilent)
	}
	if waited := time.Since(started); waited < 3*beatMS*time.Millisecond-20*time.Millisecond || waited > 5*time.Second {
		t.Fatalf("the host was declared lost after %v, want three heartbeat intervals", waited)
	}
}

func TestAHostThatKeepsHeartbeatingIsNotLostWhileItHoldsItsAnswer(t *testing.T) {
	t.Parallel()
	link := NewSimLink(5 * beatMS * time.Millisecond)
	defer link.Close()
	go func() {
		ticks := time.NewTicker(beatMS / 2 * time.Millisecond)
		defer ticks.Stop()
		for beat := 1; beat <= 12; beat++ {
			link.Deliver(inboundHeartbeat(t, beat, beat))
			select {
			case <-ticks.C:
			case <-t.Context().Done():
				return
			}
		}
	}()
	got, met, err := link.Await(t.Context(), isKind("HSYNC"))
	if met || err != nil || len(got) < 6 {
		t.Fatalf("Await = %d records, %v, %v, want a cap that expired unmet and without a loss", len(got), met, err)
	}
}

func TestSilenceBeforeTheFirstHeartbeatIsNotHostLoss(t *testing.T) {
	t.Parallel()
	link := NewSimLink(4 * beatMS * time.Millisecond)
	defer link.Close()
	link.Deliver(inboundHeartbeat(t, 1, 0))
	if _, met, err := link.Await(t.Context(), isKind("HSYNC")); met || err != nil {
		t.Fatalf("Await = %v, %v, want an unmet wait without a loss", met, err)
	}
}

func TestAHostOnTheSocketThatStopsHeartbeatingIsLostWhateverTheCap(t *testing.T) {
	t.Parallel()
	path, listener := privateListener(t)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		writeLine(t, conn, heartbeat(t, 1, 1))
		<-t.Context().Done()
	}()
	link, err := Dial(t.Context(), path, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = link.Close() }()
	if _, met, err := link.Await(t.Context(), isKind("HSYNC")); met || !errors.Is(err, ErrHostSilent) {
		t.Fatalf("Await = %v, %v, want %v", met, err, ErrHostSilent)
	}
}

func writeLine(t *testing.T, conn net.Conn, record map[string]any) {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Error(err)
		return
	}
	_, _ = conn.Write(append(raw, '\n'))
}

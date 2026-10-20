package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type hostEnd struct {
	report host.Report
	err    error
}

func TestAReferenceHostInterruptedBeforeANodeEndsCleanlyWithAnEmptyReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	report, err := ServeHost(ctx, testsupport.SocketPath(t), 0, host.Options{Environment: "simulated", Ready: cancel})
	if err != nil || report.Refused != "" || len(report.Commands) != 0 {
		t.Fatalf("ServeHost = %+v, %v, want a clean end with an empty report", report, err)
	}
}

func TestAReferenceHostWhoseNodeStopsDuringABarrierWaitEndsCleanlyWithItsReport(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	ready, ended := make(chan struct{}), make(chan hostEnd, 1)
	go func() {
		options := host.Options{Environment: "simulated", ProtocolMax: 1, Ready: func() { close(ready) }}
		report, err := ServeHost(t.Context(), hostPath, 250*time.Millisecond, options)
		ended <- hostEnd{report: report, err: err}
	}()
	<-ready
	stopNode, nodeDone := startConnect(t, connectOptions(t, planText(hostPath, ""), out), false)
	waitForJournal(t, out, `"HSYNC"`)
	stopNode()
	<-nodeDone
	select {
	case end := <-ended:
		if end.err != nil || end.report.Refused != "" || end.report.Commands == nil {
			t.Fatalf("ServeHost = %+v, %v, want a clean end with the report of an accepted session", end.report, end.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the reference host did not end after its node stopped")
	}
}

func waitForJournal(t *testing.T, out, text string) {
	t.Helper()
	giveUp := time.After(10 * time.Second)
	for {
		if raw, err := os.ReadFile(filepath.Join(out, "journal.jsonl")); err == nil && strings.Contains(string(raw), text) {
			return
		}
		select {
		case <-giveUp:
			t.Fatalf("the journal in %s never held %s", out, text)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestTheReferenceHostBeatsInWallTimeWhileItsNodeIsSilent(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	node := dialGateway(t, hostPath, relayGateway(t, "s-1"))
	if state := node.handshake(); !state.Online || domain.Seq(state.Record["heartbeat"]) != 0 {
		t.Fatalf("handshake STATE = %v, want online with heartbeat 0", state.Record)
	}
	beats := node.await(func(m domain.Message) bool { return m.Type == "STATE" })
	if beat := beats[len(beats)-1].Record; beat["online"] != true || domain.Seq(beat["heartbeat"]) != 1 || domain.Seq(beat["heartbeat_interval_ms"]) != 1000 {
		t.Fatalf("beat = %v, want the next heartbeat of an online host", beat)
	}
	if report := started.stop(); report.Refused != "" {
		t.Fatalf("report = %+v", report)
	}
}

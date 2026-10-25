package app

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestASimNodeSwallowsItsCancellationAndKeepsTheFirstFailure(t *testing.T) {
	t.Parallel()
	first, second := errors.New("first failure"), errors.New("second failure")
	node := &simNode{}
	node.complete(context.Canceled)
	node.complete(uds.ErrSimClosed)
	if node.failure != nil {
		t.Fatalf("failure after a cancellation = %v, want none", node.failure)
	}
	node.complete(first)
	node.complete(second)
	if !errors.Is(node.failure, first) || errors.Is(node.failure, second) {
		t.Fatalf("failure = %v, want the first one kept", node.failure)
	}
}

func TestASimNodeReapsItsEndOnlyOnceItHasEnded(t *testing.T) {
	t.Parallel()
	failure := errors.New("the node ended")
	done := make(chan error, 1)
	node := &simNode{done: done}
	node.reap()
	if node.done == nil || node.failure != nil {
		t.Fatal("a node still running was reaped")
	}
	done <- failure
	node.reap()
	if node.done != nil || !errors.Is(node.failure, failure) {
		t.Fatalf("done %v, failure %v, want the end taken and its failure kept", node.done, node.failure)
	}
}

func TestASimNodeStopReportsTheFailureOfAnEndedSession(t *testing.T) {
	t.Parallel()
	failure := errors.New("the session failed")
	done := make(chan error, 1)
	done <- failure
	node := &simNode{link: uds.NewSimLink(time.Second), cancel: func() {}, done: done}
	if err := node.stop(); !errors.Is(err, failure) {
		t.Fatalf("stop = %v, want %v", err, failure)
	}
	var none *simNode
	if none.stop() != nil || none.end("", "", "", nil) != nil || none.signal() != nil {
		t.Fatal("a relay without a simulated node must stop, end and signal as nothing")
	}
}

func TestASimNodeStartsItsSessionOnTheFirstRecordsDeliveredToIt(t *testing.T) {
	t.Parallel()
	begun := 0
	node := &simNode{link: uds.NewSimLink(time.Second)}
	node.begin = func() { begun, node.started = begun+1, true }
	node.deliver(nil)
	if begun != 0 {
		t.Fatal("an empty delivery started the session")
	}
	node.deliver([]map[string]any{{"type": "STATE"}})
	node.deliver([]map[string]any{{"type": "DCMD"}})
	if begun != 1 || !node.started {
		t.Fatalf("session started %d times, want once", begun)
	}
	var none *simNode
	none.deliver([]map[string]any{{"type": "STATE"}})
}

func TestAMixedRelayThatCannotListenReleasesItsSimulatedNode(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	options := mixedOptions(t, t.TempDir())
	options.GatewaySocket, options.HostSocket = path, path+".host"
	if _, err := serveRelay(t.Context(), options, nil); !errors.Is(err, uds.ErrSocketInUse) {
		t.Fatalf("serveRelay on a live socket = %v, want %v", err, uds.ErrSocketInUse)
	}
}

func TestARelayWhoseWireJournalCannotBeOpenedClosesTheFieldTraceItOpened(t *testing.T) {
	t.Parallel()
	out, hostPath := t.TempDir(), testsupport.SocketPath(t)
	if err := os.Mkdir(filepath.Join(out, model.FileObserverJournal), 0o700); err != nil {
		t.Fatal(err)
	}
	_, err := serveRelay(t.Context(), RelayOptions{GatewaySocket: hostPath + ".g", HostSocket: hostPath, Out: out}, nil)
	var pathError *fs.PathError
	if !errors.As(err, &pathError) {
		t.Fatalf("serveRelay = %v, want the file system's refusal to open a journal that is a folder", err)
	}
}

func TestATickBeforeAnyGatewayConnectsChangesNothingAndTheSessionStillRuns(t *testing.T) {
	t.Parallel()
	hostPath, ticks := testsupport.SocketPath(t), make(chan time.Time)
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	relay := startRelay(t, hostPath, t.TempDir(), ticks)
	tick(t, ticks)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	if state := gateway.handshake(); !state.Online {
		t.Fatal("a tick before the gateway connected cost it the host's STATE")
	}
	tick(t, ticks)
	gateway.sync()
	ended := relay.stop()
	started.stop()
	if ended.err != nil {
		t.Fatalf("relay ended with %v, want no error", ended.err)
	}
}

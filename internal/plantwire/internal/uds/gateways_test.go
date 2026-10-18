package uds

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func listenGateways(t *testing.T) (string, *Gateways) {
	t.Helper()
	path := testsupport.SocketPath(t)
	gateways, err := ListenGateways(t.Context(), path, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = gateways.Close() })
	return path, gateways
}

func nextLink(t *testing.T, gateways *Gateways) *Link {
	t.Helper()
	select {
	case link := <-gateways.Links():
		t.Cleanup(func() { _ = link.Close() })
		return link
	case <-t.Context().Done():
		t.Fatal("no gateway connected")
		return nil
	}
}

func TestGatewaysPublishAnOwnerOnlySocketAndHandEachConnectionOverAsALink(t *testing.T) {
	t.Parallel()
	path, gateways := listenGateways(t)
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("socket = %v, %v, want mode 0600", info, err)
	}
	conn := dialHost(t, path)
	link := nextLink(t, gateways)
	if _, err := conn.Write([]byte("{\"type\":\"NSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	if got := <-link.Incoming(); got.Type != "NSYNC" || got.Invalid == nil {
		t.Fatalf("incoming = %+v, want the NSYNC read as a record that fails its schema", got)
	}
	if err := link.Send([]map[string]any{{"type": "STATE"}}); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || line != "{\"type\":\"STATE\"}\n" {
		t.Fatalf("the gateway read %q, %v, want the record the link sent", line, err)
	}
	if err := gateways.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat after Close = %v, want the socket removed", err)
	}
}

func TestGatewaysRefuseAPathThatIsInUse(t *testing.T) {
	t.Parallel()
	path, _ := listenGateways(t)
	if _, err := ListenGateways(t.Context(), path, time.Second); !errors.Is(err, ErrSocketInUse) {
		t.Fatalf("ListenGateways on a live relay = %v, want ErrSocketInUse", err)
	}
}

func TestAGatewayLinkEndsWhenItSaysNothingInTimeAndStaysOpenOnceSettled(t *testing.T) {
	t.Parallel()
	path, gateways := listenGateways(t)
	dialHost(t, path)
	link := nextLink(t, gateways)
	if err := link.Expect(30 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	select {
	case _, open := <-link.Incoming():
		if open {
			t.Fatal("a record arrived from a gateway that sent nothing")
		}
	case <-t.Context().Done():
		t.Fatal("the deadline never ended the link")
	}
	patient := dialHost(t, path)
	settled := nextLink(t, gateways)
	if err := settled.Expect(30 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := settled.Settle(); err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)
	if _, err := patient.Write([]byte("{\"type\":\"NSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	if got, open := <-settled.Incoming(); !open || got.Type != "NSYNC" {
		t.Fatalf("incoming = %+v open %v, want the late record of a settled link", got, open)
	}
}

func TestClosingTheGatewaysHangsUpOnAConnectionNobodyTook(t *testing.T) {
	t.Parallel()
	path, gateways := listenGateways(t)
	conn := dialHost(t, path)
	if err := gateways.Close(); err != nil {
		t.Fatal(err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	_, err := bufio.NewReader(conn).ReadByte()
	if !errors.Is(err, io.EOF) && !errors.Is(err, syscall.ECONNRESET) {
		t.Fatalf("read on a connection nobody took = %v, want it hung up", err)
	}
}

func TestAClosedGatewaysHangUpOnALinkNobodyTakesAndAnOpenOneHandsItOver(t *testing.T) {
	t.Parallel()
	client, server := net.Pipe()
	t.Cleanup(func() { _ = server.Close() })
	g := &Gateways{links: make(chan *Link), done: make(chan struct{})}
	close(g.done)
	if g.handOver(NewAccepted(client, time.Second)) {
		t.Fatal("handOver reported a hand-over on closed gateways")
	}
	if _, err := client.Write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("write on the hung-up link = %v, want io.ErrClosedPipe", err)
	}
	open := &Gateways{links: make(chan *Link, 1), done: make(chan struct{})}
	link := NewAccepted(server, time.Second)
	if !open.handOver(link) || <-open.Links() != link {
		t.Fatal("open gateways did not hand the link over")
	}
}

func TestTicksArriveAtTheirPeriodUntilStopped(t *testing.T) {
	t.Parallel()
	ticks, stop := Ticks(5 * time.Millisecond)
	defer stop()
	select {
	case <-ticks:
	case <-t.Context().Done():
		t.Fatal("no tick arrived")
	}
}

func TestGatewaysRefuseADirectoryOthersCanReach(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	widen(t, filepath.Dir(path), ".", 0o755)
	if _, err := ListenGateways(t.Context(), path, time.Second); !errors.Is(err, ErrSocketNotPrivate) {
		t.Fatalf("ListenGateways in a group-readable directory = %v, want ErrSocketNotPrivate", err)
	}
}

func TestASecondGatewayWaitsForTheFirstToBeTakenAndBothAreServedInOrder(t *testing.T) {
	t.Parallel()
	path, gateways := listenGateways(t)
	first, second := dialHost(t, path), dialHost(t, path)
	if _, err := first.Write([]byte("{\"type\":\"NSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Write([]byte("{\"type\":\"HSYNC\"}\n")); err != nil {
		t.Fatal(err)
	}
	one, two := nextLink(t, gateways), nextLink(t, gateways)
	if got := (<-one.Incoming()).Type; got != "NSYNC" {
		t.Errorf("first link read %q, want the first connection's NSYNC", got)
	}
	if got := (<-two.Incoming()).Type; got != "HSYNC" {
		t.Errorf("second link read %q, want the second connection's HSYNC", got)
	}
}

func TestGatewaysClosedAfterTheirContextEndedCloseCleanlyAndTwice(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	gateways, err := ListenGateways(ctx, testsupport.SocketPath(t), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	for range 2 {
		if err := gateways.Close(); err != nil {
			t.Fatalf("Close = %v, want a clean close of a listener the context already closed", err)
		}
	}
}

func TestALinkOfAClosedConnectionRefusesToSetItsDeadlines(t *testing.T) {
	t.Parallel()
	client, server := net.Pipe()
	link := NewAccepted(client, time.Second)
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if err := link.Close(); err != nil {
		t.Fatal(err)
	}
	if err := link.Expect(time.Second); !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("Expect on a closed link = %v, want io.ErrClosedPipe", err)
	}
	if err := link.Settle(); !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("Settle on a closed link = %v, want io.ErrClosedPipe", err)
	}
}

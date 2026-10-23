package uds

import (
	"errors"
	"net"
	"testing"
)

func TestAWriteFailureAfterTheNodeHungUpIsACleanEnd(t *testing.T) {
	t.Parallel()
	host, node := net.Pipe()
	link := NewAccepted(host, HandshakeWait)
	defer func() { _ = link.Close() }()
	_ = node.Close()
	if _, err := endOrFail(link, errors.New("broken pipe")); err != nil {
		t.Fatalf("endOrFail = %v, want a clean end", err)
	}
}

func TestAWriteFailureWhileTheNodeStaysConnectedIsReported(t *testing.T) {
	t.Parallel()
	host, node := net.Pipe()
	defer func() { _ = node.Close() }()
	link := NewAccepted(host, HandshakeWait)
	defer func() { _ = link.Close() }()
	writeErr := errors.New("write stalled")
	if _, err := endOrFail(link, writeErr); !errors.Is(err, writeErr) {
		t.Fatalf("endOrFail = %v, want the write error", err)
	}
}

func TestAMessageDrainedWhileEndingCountsAsHeard(t *testing.T) {
	t.Parallel()
	host, node := net.Pipe()
	link := NewAccepted(host, HandshakeWait)
	defer func() { _ = link.Close() }()
	go func() {
		_, _ = node.Write([]byte("{\"type\":\"NBIRTH\"}\n"))
		_ = node.Close()
	}()
	if heard, err := endOrFail(link, errors.New("broken pipe")); !heard || err != nil {
		t.Fatalf("endOrFail = %v, %v, want heard and a clean end", heard, err)
	}
}

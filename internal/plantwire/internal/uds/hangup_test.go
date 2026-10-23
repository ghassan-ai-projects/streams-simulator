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
	if err := endOrFail(link, errors.New("broken pipe")); err != nil {
		t.Fatalf("endOrFail = %v, want a clean end", err)
	}
}

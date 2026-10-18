package uds

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"
)

func TestASimulatedNodeThatSendsMoreThanTheMailboxHoldsIsToldItOverflowedNotThatTheLinkClosed(t *testing.T) {
	t.Parallel()
	link := NewSimLink(time.Second)
	err := link.Send(make([]map[string]any, mailboxLimit+1))
	if !errors.Is(err, ErrSimOverflow) || errors.Is(err, ErrSimClosed) {
		t.Fatalf("Send over the limit = %v, want ErrSimOverflow", err)
	}
	link.Close()
	if err := link.Send([]map[string]any{{"type": "NDATA"}}); !errors.Is(err, ErrSimClosed) || errors.Is(err, ErrSimOverflow) {
		t.Fatalf("Send after Close = %v, want ErrSimClosed", err)
	}
}

func TestClosingTheGatewaysRemovesTheSocketTheyMadeAndNothingElseAtThePath(t *testing.T) {
	t.Parallel()
	path, gateways := listenGateways(t)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not a socket"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := gateways.Close(); err != nil {
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(path); err != nil || string(raw) != "not a socket" {
		t.Fatalf("file at the path = %q, %v, want the file that took the socket's place left alone", raw, err)
	}
}

func TestOnlyACancellationOrAClosedConnectionIsAnInterrupt(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		err  error
		want bool
	}{
		"a cancellation":         {context.Canceled, true},
		"a closed connection":    {net.ErrClosed, true},
		"a wrapped cancellation": {errors.Join(errors.New("barrier"), context.Canceled), true},
		"a deadline":             {context.DeadlineExceeded, false},
		"another error":          {errors.New("host lost"), false},
		"nothing":                {nil, false},
	} {
		if got := Interrupted(tc.err); got != tc.want {
			t.Errorf("%s: Interrupted = %v, want %v", name, got, tc.want)
		}
	}
}

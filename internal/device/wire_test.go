package device

import (
	"bytes"
	"testing"
)

// collect runs the fault gate over a fixed sequence of single-byte "frames" and
// returns the emitted stream, so each transport fault can be asserted exactly.
func collect(t *testing.T, faults WireFaults, frames ...string) string {
	t.Helper()
	var buf bytes.Buffer
	gate := newWireGate(&buf, faults)
	for _, f := range frames {
		if err := gate.send([]byte(f)); err != nil {
			t.Fatalf("send %q: %v", f, err)
		}
	}
	if err := gate.flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	return buf.String()
}

func TestWireFaultsDrop(t *testing.T) {
	if got := collect(t, WireFaults{Drop: map[int]bool{2: true}}, "a", "b", "c"); got != "ac" {
		t.Fatalf("drop of frame 2 should yield %q, got %q", "ac", got)
	}
}

func TestWireFaultsDuplicate(t *testing.T) {
	if got := collect(t, WireFaults{Duplicate: map[int]bool{2: true}}, "a", "b", "c"); got != "abbc" {
		t.Fatalf("duplicate of frame 2 should yield %q, got %q", "abbc", got)
	}
}

func TestWireFaultsSwap(t *testing.T) {
	// Frame 2 is delayed by one position: emitted after frame 3.
	if got := collect(t, WireFaults{Swap: map[int]bool{2: true}}, "a", "b", "c"); got != "acb" {
		t.Fatalf("swap of frame 2 should yield %q, got %q", "acb", got)
	}
}

func TestWireFaultsSwapAtEndFlushes(t *testing.T) {
	// A swap on the last frame has nothing to swap with; flush releases it.
	if got := collect(t, WireFaults{Swap: map[int]bool{3: true}}, "a", "b", "c"); got != "abc" {
		t.Fatalf("swap of the final frame should still emit it, got %q", got)
	}
}

func TestWireFaultsIdentityWhenEmpty(t *testing.T) {
	if got := collect(t, WireFaults{}, "a", "b", "c"); got != "abc" {
		t.Fatalf("no faults should be identity, got %q", got)
	}
}

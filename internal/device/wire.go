package device

import "io"

// WireFaults is a deterministic, scripted transport-fault plan applied to the
// device's OUTBOUND frames (receipt/result/state), indexed by their 1-based
// emission order across a connection. It models the transport hazards a real
// serial/gateway link imposes — a lost ack, a duplicated delivery, a reordered
// pair — without any wall-clock nondeterminism, so a test can assert an exact
// outcome. (Device-level physical faults — ack_lost/stuck/reboot — live in
// Faults; this layer is strictly about frames on the wire.)
type WireFaults struct {
	// Drop omits the frame at each listed 1-based index entirely.
	Drop map[int]bool
	// Duplicate emits the frame at each listed index twice.
	Duplicate map[int]bool
	// Swap delays the frame at each listed index by one position, so it is
	// emitted immediately AFTER the next emitted frame (a reordered pair).
	// Adjacent Swap indices are not supported and should be avoided.
	Swap map[int]bool
}

func (f WireFaults) empty() bool {
	return len(f.Drop) == 0 && len(f.Duplicate) == 0 && len(f.Swap) == 0
}

// wireGate applies a WireFaults plan to a stream of whole frames.
type wireGate struct {
	faults    WireFaults
	w         io.Writer
	i         int
	held      []byte
	heldDup   bool
	heldValid bool
}

func newWireGate(w io.Writer, faults WireFaults) *wireGate {
	return &wireGate{faults: faults, w: w}
}

// send offers one complete frame (including its trailing newline) to the wire,
// applying drop/duplicate/swap by emission index.
func (g *wireGate) send(frame []byte) error {
	g.i++
	idx := g.i
	if g.faults.Drop[idx] {
		return nil // dropped; any held frame stays held until a real emit
	}
	if g.faults.Swap[idx] && !g.heldValid {
		g.held = append([]byte(nil), frame...)
		g.heldDup = g.faults.Duplicate[idx]
		g.heldValid = true
		return nil
	}
	if err := g.writeOne(frame, g.faults.Duplicate[idx]); err != nil {
		return err
	}
	if g.heldValid {
		g.heldValid = false
		return g.writeOne(g.held, g.heldDup)
	}
	return nil
}

// flush emits any frame still held by a pending swap (e.g. at connection close).
func (g *wireGate) flush() error {
	if !g.heldValid {
		return nil
	}
	g.heldValid = false
	return g.writeOne(g.held, g.heldDup)
}

func (g *wireGate) writeOne(frame []byte, duplicate bool) error {
	if _, err := g.w.Write(frame); err != nil {
		return err
	}
	if duplicate {
		if _, err := g.w.Write(frame); err != nil {
			return err
		}
	}
	return nil
}

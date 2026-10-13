package uds

import (
	"fmt"
	"io"
)

type WireFaults struct {
	Drop map[int]bool

	Duplicate map[int]bool

	Swap map[int]bool
}

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

func (g *wireGate) send(frame []byte) error {
	g.i++
	index := g.i
	if g.faults.Drop[index] {
		return nil
	}
	if g.faults.Swap[index] && !g.heldValid {
		g.holdFrame(frame, index)
		return nil
	}
	if err := g.writeOne(frame, g.faults.Duplicate[index]); err != nil {
		return err
	}
	return g.flush()
}

func (g *wireGate) flush() error {
	if !g.heldValid {
		return nil
	}
	g.heldValid = false
	return g.writeOne(g.held, g.heldDup)
}

func (g *wireGate) writeOne(frame []byte, duplicate bool) error {
	if _, err := g.w.Write(frame); err != nil {
		return fmt.Errorf("write wire frame %d: %w", g.i, err)
	}
	if duplicate {
		if _, err := g.w.Write(frame); err != nil {
			return fmt.Errorf("write duplicate wire frame %d: %w", g.i, err)
		}
	}
	return nil
}

func (g *wireGate) holdFrame(frame []byte, index int) {
	g.held = append([]byte(nil), frame...)
	g.heldDup = g.faults.Duplicate[index]
	g.heldValid = true
}

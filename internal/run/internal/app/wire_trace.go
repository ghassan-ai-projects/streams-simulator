package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
)

type WireTrace struct {
	journal *journal.Journal
}

func OpenWireTrace(dir string) (*WireTrace, error) {
	opened, err := journal.NewInDir(dir, false)
	if err != nil {
		return nil, fmt.Errorf("run: wire trace: %w", err)
	}
	return &WireTrace{journal: opened}, nil
}

func (t *WireTrace) Wire(direction string, record map[string]any) {
	t.journal.Wire(0, direction, record)
}

func (t *WireTrace) Flush() error {
	if err := t.journal.Flush(); err != nil {
		return fmt.Errorf("run: wire trace: %w", err)
	}
	return nil
}

func (t *WireTrace) Close() error {
	if _, err := t.journal.Close(0, false, ""); err != nil {
		return fmt.Errorf("run: wire trace: %w", err)
	}
	return nil
}

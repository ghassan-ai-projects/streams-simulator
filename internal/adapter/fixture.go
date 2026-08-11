package adapter

// The adapter conformance fixture: a fixed 12-event native trace, committed
// with the simulator. Every adapter's golden file is rendered from exactly
// this fixture, so `adapter verify` is a byte-comparison against a stable
// input. Verified by hand before first use: each value, timestamp and digest
// below was computed by a human, not by the simulator.

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

//go:embed testdata/fixture.jsonl
var fixtureJSONL string

// FixtureEvents returns the parsed 12-event fixture.
func FixtureEvents() ([]model.SimEvent, error) {
	var out []model.SimEvent
	for i, line := range strings.Split(strings.TrimSpace(fixtureJSONL), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev model.SimEvent
		if err := model.DecodeBytes([]byte(line), &ev); err != nil {
			return nil, fmt.Errorf("fixture line %d: %w", i, err)
		}
		out = append(out, ev)
	}
	return out, nil
}

func loadFixture(path string) ([]model.SimEvent, error) {
	if path == "" {
		return FixtureEvents()
	}
	var out []model.SimEvent
	if err := decodeJSONL(path, &out); err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	return out, nil
}

func decodeJSONL(path string, dst *[]model.SimEvent) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("adapter: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev model.SimEvent
		if err := model.DecodeBytes([]byte(line), &ev); err != nil {
			return fmt.Errorf("adapter: %w", err)
		}
		*dst = append(*dst, ev)
	}
	return nil
}

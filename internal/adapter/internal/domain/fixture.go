package domain

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

//go:embed testdata/fixture.jsonl
var fixtureJSONL string

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

func DecodeFixtureRecords(raw []byte) ([]model.SimEvent, error) {
	var out []model.SimEvent
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev model.SimEvent
		if err := model.DecodeBytes([]byte(line), &ev); err != nil {
			return nil, fmt.Errorf("adapter: %w", err)
		}
		out = append(out, ev)
	}
	return out, nil
}

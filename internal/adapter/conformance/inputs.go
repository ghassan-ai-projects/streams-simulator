package conformance

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func verificationInputs(adapterPath, fixturePath string) (*model.Adapter, []model.SimEvent, error) {
	a, err := adapter.Load(adapterPath)
	if err != nil {
		return nil, nil, fmt.Errorf("adapter: %w", err)
	}
	fixture, err := loadFixture(fixturePath)
	if err != nil {
		return nil, nil, fmt.Errorf("adapter: %w", err)
	}
	if err := validateStrictObservedOrder(fixture); err != nil {
		return nil, nil, fmt.Errorf("adapter: %w", err)
	}
	return a, fixture, nil
}

func renderVerification(a *model.Adapter, fixture []model.SimEvent) ([]byte, error) {
	engine, err := adapter.NewEngine(a, verificationMetadata(fixture))
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	out, err := engine.RenderRun(fixture, mustParse(fixture[len(fixture)-1].ObservedTime))
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	return out, nil
}

func verificationMetadata(fixture []model.SimEvent) map[string]any {
	return map[string]any{
		"run_id": "verify", "sim_version": "0.1.0",
		"domain_id": "fixture", "domain_version": "0.0.0",
		"world_start_time": fixture[0].EventTime, "world_end_time": fixture[len(fixture)-1].EventTime,
		"seed": float64(0),
	}
}

func verifyRendered(a *model.Adapter, out []byte, base string) (*Result, error) {
	res := &Result{Adapter: a.ID}
	if complete, err := verifyOutputSchema(a, out, base, res); err != nil {
		return nil, err
	} else if !complete {
		return res, nil
	}
	if err := verifyGolden(a, out, base, res); err != nil {
		return nil, err
	}
	return res, nil
}

func loadFixture(path string) ([]model.SimEvent, error) {
	if path == "" {
		events, err := adapter.FixtureEvents()
		if err != nil {
			return nil, fmt.Errorf("adapter: embedded fixture: %w", err)
		}
		return events, nil
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
	return appendFixtureRecords(raw, dst)
}

func appendFixtureRecords(raw []byte, dst *[]model.SimEvent) error {
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

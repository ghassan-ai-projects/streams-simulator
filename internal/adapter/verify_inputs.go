package adapter

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func verificationInputs(adapterPath, fixturePath string) (*model.Adapter, []model.SimEvent, error) {
	a, err := Load(adapterPath)
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
	engine, err := NewEngine(a, verificationMetadata(fixture))
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

func verifyRendered(a *model.Adapter, out []byte, base string) (*VerifyResult, error) {
	res := &VerifyResult{Adapter: a.ID}
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

package app

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	labelSampleNS  int64 = 120 * 1e9
	labelHorizonNS int64 = 72 * 3600 * 1e9
)

func (p *connectPlan) writeLabel(out string) error {
	if len(p.session.Disturbances) != 1 {
		return nil
	}
	record, err := p.groundTruth(p.session.Disturbances[0])
	if err != nil {
		return fmt.Errorf("ground truth of the planned disturbance: %w", err)
	}
	raw, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("ground truth of the planned disturbance: %w", err)
	}
	return files.WriteRunFile(out, model.FileLabel, append(raw, '\n'))
}

func (p *connectPlan) groundTruth(d model.Disturbance) (*model.GroundTruthRecord, error) {
	seed := p.session.WorldSeed()
	solver, err := truth.NewSolver(p.spec, seed, labelSampleNS, labelHorizonNS)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	record, err := truth.BuildRecord(p.spec, solver, p.injection(d, seed))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return record, nil
}

func (p *connectPlan) injection(d model.Disturbance, seed uint64) truth.Injection {
	start := p.startNS()
	onset := domain.ScheduleDisturbances([]model.Disturbance{d}, start)[0].OnsetNS
	return truth.Injection{ScenarioID: p.spec.Spec.ID + "/0001", Seed: seed, EntityID: d.Instance, FaultID: d.Fault,
		OnsetNS: onset, StartNS: start, EntityIDs: world.PopulationIDs(p.spec, 0)}
}

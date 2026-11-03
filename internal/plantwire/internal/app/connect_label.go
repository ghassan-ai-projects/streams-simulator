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

const labelSampleNS int64 = 120 * 1e9

func (p *connectPlan) writeLabel(out string) error {
	records, err := p.groundTruths()
	if err != nil {
		return fmt.Errorf("ground truth of the planned disturbances: %w", err)
	}
	if len(records) == 1 {
		return writeLabelFile(out, model.FileLabel, records[0])
	}
	return writeLabelFile(out, model.FileLabels, records)
}

func writeLabelFile(out, name string, document any) error {
	raw, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("ground truth of the planned disturbances: %w", err)
	}
	return files.WriteRunFile(out, name, append(raw, '\n'))
}

func (p *connectPlan) groundTruths() ([]*model.GroundTruthRecord, error) {
	seed := p.session.WorldSeed()
	solver, err := truth.NewSolver(p.spec, seed, labelSampleNS, 0)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return p.recordsOf(solver, seed)
}

func (p *connectPlan) recordsOf(solver *truth.Solver, seed uint64) ([]*model.GroundTruthRecord, error) {
	records := []*model.GroundTruthRecord{}
	for i, d := range p.session.Disturbances {
		record, err := p.groundTruth(solver, d, p.injection(d, seed, i+1))
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (p *connectPlan) groundTruth(solver *truth.Solver, d model.Disturbance, injection truth.Injection) (*model.GroundTruthRecord, error) {
	record, err := truth.BuildRecord(p.spec, solver, injection)
	if err != nil {
		return nil, fmt.Errorf("%s on %s: %w", d.Fault, d.Instance, err)
	}
	return record, nil
}

func (p *connectPlan) injection(d model.Disturbance, seed uint64, ordinal int) truth.Injection {
	start := p.startNS()
	onset := domain.ScheduleDisturbances([]model.Disturbance{d}, start)[0].OnsetNS
	return truth.Injection{ScenarioID: fmt.Sprintf("%s/%04d", p.spec.Spec.ID, ordinal), Seed: seed, EntityID: d.Instance, FaultID: d.Fault,
		OnsetNS: onset, StartNS: start, EntityIDs: world.PopulationIDs(p.spec, 0)}
}

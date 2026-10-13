package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func Check(spec *domain.Compiled, options Options) (*Report, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	g, err := newGrid(model.DefaultStartTimeNS, options.Hours)
	if err != nil {
		return nil, err
	}
	w, err := world.New(spec, options.Seed, "lint", g.startNS, world.Options{Noiseless: true, EmitDisabled: true})
	if err != nil {
		return nil, fmt.Errorf("lint: %w", err)
	}
	return newReport(spec, options, g, judge(samplePlants(w, spec, g), g)), nil
}

func judge(plants []*plant, g *grid) []Finding {
	findings := []Finding{}
	for _, p := range plants {
		for _, s := range p.all {
			findings = append(findings, seriesFindings(s, g)...)
		}
		findings = append(findings, conservationFindings(p, g)...)
	}
	return findings
}

func seriesFindings(s *series, g *grid) []Finding {
	if finding, ok := nonFinite(s, g); ok {
		return []Finding{finding}
	}
	var findings []Finding
	for _, rule := range []func(*series, *grid) (Finding, bool){outOfRange, pinnedAtBound, notSettled} {
		if finding, ok := rule(s, g); ok {
			findings = append(findings, finding)
		}
	}
	return findings
}

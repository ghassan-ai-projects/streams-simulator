package domain

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const residualSigmas = 3

type balance struct {
	inputs  []*series
	outputs []*series
	limit   float64
}

func conservationFindings(p *plant, g *grid) []Finding {
	var findings []Finding
	for i := range p.kind.Faults {
		fault := &p.kind.Faults[i]
		if fault.Observability.Detector.Form != model.DetectorConservation {
			continue
		}
		if finding, ok := unbalanced(p, fault, g); ok {
			findings = append(findings, finding)
		}
	}
	return findings
}

func unbalanced(p *plant, fault *model.Fault, g *grid) (Finding, bool) {
	b, ok := p.balanceOf(&fault.Observability.Detector)
	if !ok {
		return Finding{}, false
	}
	breaches := b.breaches(len(g.timesNS))
	if len(breaches) == 0 {
		return Finding{}, false
	}
	worst := slices.MaxFunc(breaches, func(x, y int) int { return cmp.Compare(math.Abs(b.residual(x)), math.Abs(b.residual(y))) })
	detail := fmt.Sprintf("inputs minus outputs reaches %s at %s s, beyond %d x the combined noise sigma (limit %s)",
		number(b.residual(worst)), number(g.atS(worst)), residualSigmas, number(b.limit))
	return Finding{Severity: severityWarning, Code: codeConservation, Entity: p.id, Fault: fault.ID, AtS: g.atS(breaches[0]), Detail: detail}, true
}

func (p *plant) balanceOf(det *model.Detector) (balance, bool) {
	b := balance{inputs: p.seriesOf(det.Inputs), outputs: p.seriesOf(det.Outputs)}
	if len(b.inputs) != len(det.Inputs) || len(b.outputs) != len(det.Outputs) {
		return balance{}, false
	}
	b.limit = residualSigmas * combinedSigma(slices.Concat(b.inputs, b.outputs))
	return b, true
}

func (b balance) residual(i int) float64 {
	return sumAt(b.inputs, i) - sumAt(b.outputs, i)
}

func (b balance) breaches(samples int) []int {
	var found []int
	for i := range samples {
		if residual := b.residual(i); finite(residual) && math.Abs(residual) > b.limit {
			found = append(found, i)
		}
	}
	return found
}

func (p *plant) seriesOf(names []string) []*series {
	var found []*series
	for _, name := range names {
		if s := p.channels[name]; s != nil {
			found = append(found, s)
		}
	}
	return found
}

func combinedSigma(all []*series) float64 {
	var variance float64
	for _, s := range all {
		variance += s.channel.Noise.Sigma * s.channel.Noise.Sigma
	}
	return math.Sqrt(variance)
}

func sumAt(all []*series, i int) float64 {
	var sum float64
	for _, s := range all {
		sum += s.values[i]
	}
	return sum
}

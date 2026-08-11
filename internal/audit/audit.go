// Package audit implements the trivial-baseline audit: every candidate
// scenario runs against a panel of one-line detectors fitted with hindsight
// on the scenario itself — deliberately unfair to the scenario. If a
// detector tuned on the answer still cannot separate the fault from its
// control at >= 0.9 balanced accuracy, the scenario is non_trivial and may
// enter the graded suite. Trivial scenarios stay as mechanism regression
// fixtures but never count as evidence that reasoning helped.
package audit

import (
	"fmt"
	"math"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// BalancedAccuracyCutoff is the hindsight-fitted bar for triviality.
const BalancedAccuracyCutoff = 0.9

// DetectorNames is the panel in a stable order.
var DetectorNames = []string{
	"fixed_threshold", "zscore", "first_difference", "moving_median_residual", "channel_silence",
}

// Verdict is the audit outcome for one scenario.
type Verdict struct {
	Trivial   bool               `json:"trivial"`
	Scores    map[string]float64 `json:"scores"`
	Best      string             `json:"best"`
	BestScore float64            `json:"best_score"`
	Channels  []string           `json:"channels"`
	Samples   int                `json:"samples"`
}

// Panel audits one injection: it runs the world twice (clean control and
// faulted), samples the delivered channel readings on a grid, and fits each
// trivial detector with hindsight on the labeled series.
type Panel struct {
	spec     *domain.Compiled
	seed     uint64
	sampleNS int64
}

// NewPanel builds the audit panel.
func NewPanel(spec *domain.Compiled, seed uint64, sampleNS int64) *Panel {
	if sampleNS <= 0 {
		sampleNS = 60 * 1e9
	}
	return &Panel{spec: spec, seed: seed, sampleNS: sampleNS}
}

// Audit runs the panel on one injection. The control is the domain's
// declared negative-class scenario (transient_none-style) when one exists,
// else a clean world: the design's bar is separating the fault from its
// negative controls, not from an empty trace. setup applies pre-fault
// effector calls to both worlds (scenario context such as an aerator
// running at night).
func (p *Panel) Audit(entityID, faultID string, onsetNS, startNS int64, entityIDs []string, durationNS int64, setup []truth.SetupCall) (*Verdict, error) {
	controlFault := ""
	for i := range p.spec.Spec.Faults {
		if p.spec.Spec.Faults[i].IsNegativeClass {
			controlFault = p.spec.Spec.Faults[i].ID
			break
		}
	}
	clean, cleanEmissions, err := p.build(entityID, startNS, entityIDs, map[string]int64{controlFault: onsetNS}, setup, durationNS)
	if err != nil {
		return nil, fmt.Errorf("Audit: %w", err)
	}
	faulted, faultEmissions, err := p.build(entityID, startNS, entityIDs, map[string]int64{faultID: onsetNS}, setup, durationNS)
	if err != nil {
		return nil, fmt.Errorf("Audit: %w", err)
	}
	end := startNS + durationNS
	n := int((end-startNS)/p.sampleNS) + 1
	channels := p.spec.ChannelNames()
	labels := make([]float64, n)
	faultSeries := make(map[string][]float64, len(channels))
	controlSeries := make(map[string][]float64, len(channels))
	for _, ch := range channels {
		faultSeries[ch] = make([]float64, n)
		controlSeries[ch] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		t := startNS + int64(i)*p.sampleNS
		labels[i] = 0
		if t >= onsetNS {
			labels[i] = 1
		}
		for _, ch := range channels {
			faultSeries[ch][i] = faulted.Reading(entityID, ch, t)
			controlSeries[ch][i] = clean.Reading(entityID, ch, t)
		}
	}
	v := &Verdict{Scores: map[string]float64{}, Channels: channels, Samples: n}
	var bestScore float64
	for _, name := range DetectorNames {
		var score float64
		if name == "channel_silence" {
			score = fitSilence(labels, channels, faultEmissions, cleanEmissions, startNS, end, p.sampleNS)
		} else {
			score = p.fit(name, labels, faultSeries, controlSeries)
		}
		v.Scores[name] = score
		if score > bestScore {
			bestScore = score
			v.Best = name
		}
	}
	v.BestScore = bestScore
	v.Trivial = bestScore >= BalancedAccuracyCutoff
	return v, nil
}

// emissionLog records the native events of one world, keyed by channel.
type emissionLog map[string][]int64

func (p *Panel) build(entityID string, startNS int64, entityIDs []string, faults map[string]int64, setup []truth.SetupCall, durationNS int64) (*world.World, emissionLog, error) {
	w, err := world.New(p.spec, p.seed, "w-audit", startNS, world.Options{
		InitialEntities: entityIDs,
		ForceEffectorOK: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build: %w", err)
	}
	for _, call := range setup {
		if _, err := w.InvokeEffector(call.Effector, call.EntityID, call.CommandID, call.Args, call.AtNS); err != nil {
			return nil, nil, fmt.Errorf("build: %w", err)
		}
	}
	log := emissionLog{}
	w.SetEmitter(func(ev model.SimEvent) {
		t, _ := model.ParseTime(ev.EventTime)
		log[ev.Channel] = append(log[ev.Channel], t)
	})
	for fid, onset := range faults {
		if fid == "" {
			continue
		}
		if _, err := w.InjectFault(entityID, fid, onset, nil); err != nil {
			return nil, nil, fmt.Errorf("build: %w", err)
		}
	}
	// Emit everything up to the horizon so the log is complete. The audit
	// only needs the scenario window, so horizon is bounded by 24h.
	horizon := startNS + 24*3600*1e9
	if durationNS > 0 && startNS+durationNS < horizon {
		horizon = startNS + durationNS
	}
	if _, _, err := w.Advance(horizon); err != nil {
		return nil, nil, fmt.Errorf("build: %w", err)
	}
	return w, log, nil
}

// fitSilence flags samples that follow a gap longer than minGap in any
// channel's emission record.
func fitSilence(labels []float64, channels []string, fault, control emissionLog, startNS, endNS, sampleNS int64) float64 {
	best := 0.0
	n := len(labels)
	for _, ch := range channels {
		for gap := 3; gap <= 20; gap += 2 {
			preds := silencePredict(n, startNS, sampleNS, fault[ch], gap)
			cpreds := silencePredict(n, startNS, sampleNS, control[ch], gap)
			allLabels := append(append([]float64{}, labels...), make([]float64, len(cpreds))...)
			allPreds := append(preds, cpreds...)
			if ba := balancedAccuracy(allLabels, allPreds); ba > best {
				best = ba
			}
		}
	}
	return best
}

func silencePredict(n int, startNS, sampleNS int64, emissions []int64, minGap int) []float64 {
	out := make([]float64, n)
	lastEmit := int64(-1)
	ei := 0
	for i := 0; i < n; i++ {
		t := startNS + int64(i)*sampleNS
		for ei < len(emissions) && emissions[ei] <= t {
			lastEmit = emissions[ei]
			ei++
		}
		if lastEmit < 0 {
			continue
		}
		if t-lastEmit > int64(minGap)*sampleNS {
			out[i] = 1
		}
	}
	return out
}

// fit runs one detector over the faulted series (predictions) and scores
// balanced accuracy against the labels and the control series.
func (p *Panel) fit(name string, labels []float64, fault, control map[string][]float64) float64 {
	switch name {
	case "fixed_threshold":
		return p.fitThreshold(labels, fault, control)
	case "zscore":
		return p.fitZScore(labels, fault, control)
	case "first_difference":
		return p.fitFirstDifference(labels, fault, control)
	case "moving_median_residual":
		return p.fitMedianResidual(labels, fault, control)
	}
	return 0
}

func balancedAccuracy(labels, preds []float64) float64 {
	var tp, tn, fp, fn float64
	for i := range labels {
		switch {
		case labels[i] == 1 && preds[i] == 1:
			tp++
		case labels[i] == 0 && preds[i] == 0:
			tn++
		case labels[i] == 0 && preds[i] == 1:
			fp++
		case labels[i] == 1 && preds[i] == 0:
			fn++
		}
	}
	if tp+fn == 0 || tn+fp == 0 {
		return 0
	}
	return ((tp / (tp + fn)) + (tn / (tn + fp))) / 2
}

// fitThreshold sweeps single-channel level thresholds with hindsight.
func (p *Panel) fitThreshold(labels []float64, fault, control map[string][]float64) float64 {
	best := 0.0
	for _, ch := range p.spec.ChannelNames() {
		fs := fault[ch]
		cs := control[ch]
		all := append(append([]float64{}, fs...), cs...)
		if len(all) < 2 {
			continue
		}
		// Sweep a coarse grid of thresholds (20 quantiles).
		sorted := append([]float64{}, all...)
		sort.Float64s(sorted)
		grid := make([]float64, 0, 20)
		for q := 1; q < 20; q++ {
			grid = append(grid, sorted[len(sorted)*q/20])
		}
		for _, th := range grid {
			for _, direction := range []int{1, -1} {
				preds := make([]float64, len(fs))
				for i := range fs {
					if direction == 1 && fs[i] > th {
						preds[i] = 1
					}
					if direction == -1 && fs[i] < th {
						preds[i] = 1
					}
				}
				// The control must not fire spuriously: pool control labels 0.
				for i := range cs {
					if direction == 1 && cs[i] > th {
						preds = append(preds, 1)
					} else if direction == -1 && cs[i] < th {
						preds = append(preds, 1)
					} else {
						preds = append(preds, 0)
					}
				}
				allLabels := append(append([]float64{}, labels...), make([]float64, len(cs))...)
				if ba := balancedAccuracy(allLabels, preds); ba > best {
					best = ba
				}
			}
		}
	}
	return best
}

// fitZScore flags |x - rolling_mean| > k * rolling_std, best k and window.
func (p *Panel) fitZScore(labels []float64, fault, control map[string][]float64) float64 {
	best := 0.0
	for _, ch := range p.spec.ChannelNames() {
		fs := fault[ch]
		for _, w := range []int{5, 10, 30, 60} {
			for k := 0.5; k <= 4.01; k += 0.5 {
				preds := zscorePredict(fs, w, k)
				cpreds := zscorePredict(control[ch], w, k)
				allLabels := append(append([]float64{}, labels...), make([]float64, len(cpreds))...)
				allPreds := append(preds, cpreds...)
				if ba := balancedAccuracy(allLabels, allPreds); ba > best {
					best = ba
				}
			}
		}
	}
	return best
}

func zscorePredict(x []float64, window int, k float64) []float64 {
	out := make([]float64, len(x))
	for i := range x {
		lo := i - window + 1
		if lo < 0 {
			lo = 0
		}
		var sum, sumsq float64
		n := i - lo + 1
		for j := lo; j <= i; j++ {
			sum += x[j]
			sumsq += x[j] * x[j]
		}
		mean := sum / float64(n)
		variance := sumsq/float64(n) - mean*mean
		if variance < 1e-12 {
			continue
		}
		if math.Abs(x[i]-mean)/math.Sqrt(variance) > k {
			out[i] = 1
		}
	}
	return out
}

// fitFirstDifference flags the largest single-step changes, best threshold.
func (p *Panel) fitFirstDifference(labels []float64, fault, control map[string][]float64) float64 {
	best := 0.0
	for _, ch := range p.spec.ChannelNames() {
		fs := fault[ch]
		cs := control[ch]
		diffs := []float64{}
		for i := 1; i < len(fs); i++ {
			diffs = append(diffs, math.Abs(fs[i]-fs[i-1]))
		}
		for i := 1; i < len(cs); i++ {
			diffs = append(diffs, math.Abs(cs[i]-cs[i-1]))
		}
		if len(diffs) == 0 {
			continue
		}
		sorted := append([]float64{}, diffs...)
		sort.Float64s(sorted)
		for q := 1; q < 20; q++ {
			th := sorted[len(sorted)*q/20]
			preds := make([]float64, len(fs))
			for i := 1; i < len(fs); i++ {
				if math.Abs(fs[i]-fs[i-1]) > th {
					preds[i] = 1
				}
			}
			cpreds := make([]float64, len(cs))
			for i := 1; i < len(cs); i++ {
				if math.Abs(cs[i]-cs[i-1]) > th {
					cpreds[i] = 1
				}
			}
			allLabels := append(append([]float64{}, labels...), make([]float64, len(cs))...)
			allPreds := append(preds, cpreds...)
			if ba := balancedAccuracy(allLabels, allPreds); ba > best {
				best = ba
			}
		}
	}
	return best
}

// fitMedianResidual flags the residual from a trailing median, best window.
func (p *Panel) fitMedianResidual(labels []float64, fault, control map[string][]float64) float64 {
	best := 0.0
	for _, ch := range p.spec.ChannelNames() {
		fs := fault[ch]
		for _, w := range []int{5, 11, 21, 41} {
			for _, th := range []float64{0.1, 0.25, 0.5, 1.0, 2.0} {
				preds := medianResidualPredict(fs, w, th)
				cpreds := medianResidualPredict(control[ch], w, th)
				allLabels := append(append([]float64{}, labels...), make([]float64, len(cpreds))...)
				allPreds := append(preds, cpreds...)
				if ba := balancedAccuracy(allLabels, allPreds); ba > best {
					best = ba
				}
			}
		}
	}
	return best
}

func medianResidualPredict(x []float64, window int, th float64) []float64 {
	out := make([]float64, len(x))
	for i := range x {
		lo := i - window + 1
		if lo < 0 {
			lo = 0
		}
		win := make([]float64, i-lo+1)
		copy(win, x[lo:i+1])
		sort.Float64s(win)
		med := win[len(win)/2]
		if math.Abs(x[i]-med) > th {
			out[i] = 1
		}
	}
	return out
}

package audit

import (
	"math"
	"sort"
)

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

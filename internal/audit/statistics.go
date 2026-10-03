package audit

import (
	"math"
)

// fitThreshold sweeps single-channel level thresholds with hindsight.
func (p *Panel) fitThreshold(labels []float64, fault, control map[string][]float64) float64 {
	best := 0.0
	for _, channel := range p.spec.ChannelNames() {
		best = fitChannelThreshold(best, labels, fault[channel], control[channel])
	}
	return best
}

// fitZScore flags |x - rolling_mean| > k * rolling_std, best k and window.
func (p *Panel) fitZScore(labels []float64, fault, control map[string][]float64) float64 {
	best := 0.0
	for _, channel := range p.spec.ChannelNames() {
		for _, window := range []int{5, 10, 30, 60} {
			for _, threshold := range zscoreThresholds() {
				predictions := zscorePredict(fault[channel], window, threshold)
				controls := zscorePredict(control[channel], window, threshold)
				best = betterAccuracy(best, labels, predictions, controls)
			}
		}
	}
	return best
}

func zscorePredict(x []float64, window int, k float64) []float64 {
	out := make([]float64, len(x))
	for i := range x {
		mean, variance := trailingMoments(x, i, window)
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
	for _, channel := range p.spec.ChannelNames() {
		best = fitChannelDifference(best, labels, fault[channel], control[channel])
	}
	return best
}

// fitMedianResidual flags the residual from a trailing median, best window.
func (p *Panel) fitMedianResidual(labels []float64, fault, control map[string][]float64) float64 {
	best := 0.0
	for _, channel := range p.spec.ChannelNames() {
		for _, window := range []int{5, 11, 21, 41} {
			for _, threshold := range []float64{0.1, 0.25, 0.5, 1.0, 2.0} {
				predictions := medianResidualPredict(fault[channel], window, threshold)
				controls := medianResidualPredict(control[channel], window, threshold)
				best = betterAccuracy(best, labels, predictions, controls)
			}
		}
	}
	return best
}

func medianResidualPredict(x []float64, window int, th float64) []float64 {
	out := make([]float64, len(x))
	for i := range x {
		median := trailingMedian(x, i, window)
		if math.Abs(x[i]-median) > th {
			out[i] = 1
		}
	}
	return out
}

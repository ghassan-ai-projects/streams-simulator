package audit

import (
	"math"
	"sort"
)

func fitChannelThreshold(best float64, labels, fault, control []float64) float64 {
	all := append(append([]float64{}, fault...), control...)
	if len(all) < 2 {
		return best
	}
	for _, threshold := range quantileThresholds(all) {
		for _, direction := range []int{1, -1} {
			predictions := thresholdPredict(fault, threshold, direction)
			controls := thresholdPredict(control, threshold, direction)
			best = betterAccuracy(best, labels, predictions, controls)
		}
	}
	return best
}

func quantileThresholds(values []float64) []float64 {
	sorted := append([]float64{}, values...)
	sort.Float64s(sorted)
	grid := make([]float64, 0, 20)
	for q := 1; q < 20; q++ {
		grid = append(grid, sorted[len(sorted)*q/20])
	}
	return grid
}

func thresholdPredict(values []float64, threshold float64, direction int) []float64 {
	out := make([]float64, len(values))
	for i, value := range values {
		if direction == 1 && value > threshold {
			out[i] = 1
		}
		if direction == -1 && value < threshold {
			out[i] = 1
		}
	}
	return out
}

func zscoreThresholds() []float64 {
	var thresholds []float64
	for k := 0.5; k <= 4.01; k += 0.5 {
		thresholds = append(thresholds, k)
	}
	return thresholds
}

func fitChannelDifference(best float64, labels, fault, control []float64) float64 {
	differences := append(firstDifferences(fault), firstDifferences(control)...)
	if len(differences) == 0 {
		return best
	}
	for _, threshold := range quantileThresholds(differences) {
		predictions := differencePredict(fault, threshold)
		controls := differencePredict(control, threshold)
		best = betterAccuracy(best, labels, predictions, controls)
	}
	return best
}

func firstDifferences(values []float64) []float64 {
	differences := []float64{}
	for i := 1; i < len(values); i++ {
		differences = append(differences, math.Abs(values[i]-values[i-1]))
	}
	return differences
}

func differencePredict(values []float64, threshold float64) []float64 {
	out := make([]float64, len(values))
	for i := 1; i < len(values); i++ {
		if math.Abs(values[i]-values[i-1]) > threshold {
			out[i] = 1
		}
	}
	return out
}

func trailingMoments(values []float64, index, window int) (float64, float64) {
	lo := max(index-window+1, 0)
	var sum, sumsq float64
	n := index - lo + 1
	for j := lo; j <= index; j++ {
		sum += values[j]
		sumsq += values[j] * values[j]
	}
	mean := sum / float64(n)
	return mean, sumsq/float64(n) - mean*mean
}

func trailingMedian(values []float64, index, window int) float64 {
	lo := max(index-window+1, 0)
	trailing := make([]float64, index-lo+1)
	copy(trailing, values[lo:index+1])
	sort.Float64s(trailing)
	return trailing[len(trailing)/2]
}

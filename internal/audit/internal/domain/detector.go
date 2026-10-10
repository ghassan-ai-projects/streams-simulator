package domain

// fitSilence flags samples that follow a gap longer than minGap in any
// channel's emission record.
func fitSilence(labels []float64, channels []string, fault, control emissionLog, startNS, endNS, sampleNS int64) float64 {
	best := 0.0
	for _, channel := range channels {
		for gap := 3; gap <= 20; gap += 2 {
			predictions := silencePredict(len(labels), startNS, sampleNS, fault[channel], gap)
			controls := silencePredict(len(labels), startNS, sampleNS, control[channel], gap)
			best = betterAccuracy(best, labels, predictions, controls)
		}
	}
	return best
}

func silencePredict(n int, startNS, sampleNS int64, emissions []int64, minGap int) []float64 {
	out := make([]float64, n)
	lastEmit := int64(-1)
	index := 0
	for i := range out {
		at := startNS + int64(i)*sampleNS
		index, lastEmit = latestEmission(emissions, index, lastEmit, at)
		if lastEmit >= 0 && at-lastEmit > int64(minGap)*sampleNS {
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
	var counts classificationCounts
	for i := range labels {
		counts.record(labels[i], preds[i])
	}
	if counts.tp+counts.fn == 0 || counts.tn+counts.fp == 0 {
		return 0
	}
	return ((counts.tp / (counts.tp + counts.fn)) + (counts.tn / (counts.tn + counts.fp))) / 2
}

func latestEmission(emissions []int64, index int, last, at int64) (int, int64) {
	for index < len(emissions) && emissions[index] <= at {
		last = emissions[index]
		index++
	}
	return index, last
}

type classificationCounts struct{ tp, tn, fp, fn float64 }

func (c *classificationCounts) record(label, prediction float64) {
	switch {
	case label == 1 && prediction == 1:
		c.tp++
	case label == 0 && prediction == 0:
		c.tn++
	case label == 0 && prediction == 1:
		c.fp++
	case label == 1 && prediction == 0:
		c.fn++
	}
}

func betterAccuracy(best float64, labels, predictions, controls []float64) float64 {
	allLabels := append(append([]float64{}, labels...), make([]float64, len(controls))...)
	allPredictions := append(predictions, controls...)
	if score := balancedAccuracy(allLabels, allPredictions); score > best {
		return score
	}
	return best
}

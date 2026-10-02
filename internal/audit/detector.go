package audit

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

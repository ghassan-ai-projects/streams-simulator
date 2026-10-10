package domain

type auditEvidence struct {
	labels         []float64
	channels       []string
	fault, control map[string][]float64
}

func (p *Panel) controlFault() string {
	for i := range p.spec.Spec.Faults {
		if p.spec.Spec.Faults[i].IsNegativeClass {
			return p.spec.Spec.Faults[i].ID
		}
	}
	return ""
}

func (p *Panel) sampleEvidence(clean, faulted map[string][]float64, onset, start, duration int64) auditEvidence {
	end := start + duration
	n := int((end-start)/p.sampleNS) + 1
	channels := p.spec.ChannelNames()
	evidence := auditEvidence{labels: make([]float64, n), channels: channels,
		fault: make(map[string][]float64, len(channels)), control: make(map[string][]float64, len(channels))}
	for _, channel := range channels {
		evidence.fault[channel] = make([]float64, n)
		evidence.control[channel] = make([]float64, n)
	}
	p.fillEvidence(&evidence, clean, faulted, onset, start)
	return evidence
}

func (p *Panel) fillEvidence(e *auditEvidence, clean, faulted map[string][]float64, onset, start int64) {
	for i := range e.labels {
		t := start + int64(i)*p.sampleNS
		e.labels[i] = 0
		if t >= onset {
			e.labels[i] = 1
		}
		for _, channel := range e.channels {
			e.fault[channel][i] = gridValue(faulted[channel], i)
			e.control[channel][i] = gridValue(clean[channel], i)
		}
	}
}

func (p *Panel) gradeEvidence(e auditEvidence, clean, fault emissionLog, start, duration int64) *Verdict {
	v := &Verdict{Scores: map[string]float64{}, Channels: e.channels, Samples: len(e.labels)}
	for _, name := range DetectorNames {
		score := p.detectorScore(name, e, clean, fault, start, start+duration)
		v.Scores[name] = score
		if score > v.BestScore {
			v.BestScore = score
			v.Best = name
		}
	}
	v.Trivial = v.BestScore >= BalancedAccuracyCutoff
	return v
}

func (p *Panel) detectorScore(name string, e auditEvidence, clean, fault emissionLog, start, end int64) float64 {
	if name == "channel_silence" {
		return fitSilence(e.labels, e.channels, fault, clean, start, end, p.sampleNS)
	}
	return p.fit(name, e.labels, e.fault, e.control)
}

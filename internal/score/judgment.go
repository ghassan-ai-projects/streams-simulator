package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func judgment(r *run.Run, gt *model.GroundTruthRecord) JudgmentMetrics {
	v := r.Verdict()
	m := JudgmentMetrics{DetectionCount: len(v.Detections)}
	if gt.IsNegativeClass && m.DetectionCount > 0 {
		m.FalsePositive = true
	}
	if gt.FirstObservableTimeNS <= 0 {
		return m
	}
	bestAt := int64(0)
	// The graded detection: first detection on the scenario entity, after
	// first_observable_time.
	for _, d := range v.Detections {
		t, err := model.ParseTime(d.DetectedAt)
		if err != nil {
			continue
		}
		if d.EntityID != gt.EntityID {
			continue
		}
		if t < gt.FirstObservableTimeNS {
			m.Suspicious = true
			continue
		}
		if !m.Detected || t < bestAt {
			m.Detected = true
			bestAt = t
			m.DetectionLatencyNS = t - gt.FirstObservableTimeNS
			m.LabelCorrect = d.Label == gt.Label
		}
	}
	return m
}

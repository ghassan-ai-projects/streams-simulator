package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func judgment(r *run.Run, gt *model.GroundTruthRecord) JudgmentMetrics {
	return judgmentFrom(r.Verdict(), gt)
}

func judgmentFrom(v *model.Verdict, gt *model.GroundTruthRecord) JudgmentMetrics {
	m := JudgmentMetrics{DetectionCount: len(v.Detections)}
	m.FalsePositive = gt.IsNegativeClass && m.DetectionCount > 0
	if gt.FirstObservableTimeNS <= 0 {
		return m
	}
	bestAt := int64(0)
	for _, detection := range v.Detections {
		gradeDetection(&m, &bestAt, detection, gt)
	}
	return m
}

func gradeDetection(m *JudgmentMetrics, bestAt *int64, detection model.Detection, gt *model.GroundTruthRecord) {
	at, err := model.ParseTime(detection.DetectedAt)
	if err != nil || detection.EntityID != gt.EntityID {
		return
	}
	if at < gt.FirstObservableTimeNS {
		m.Suspicious = true
		return
	}
	if !m.Detected || at < *bestAt {
		m.Detected, *bestAt = true, at
		m.DetectionLatencyNS = at - gt.FirstObservableTimeNS
		m.LabelCorrect = detection.Label == gt.Label
	}
}

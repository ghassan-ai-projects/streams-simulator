package domain

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"sort"
)

func (r *Runner) admitDelivery(ev model.SimEvent) bool {
	// Deduplicate exact re-feeds; distinct observed times count separately.
	key := fmt.Sprintf("%d@%s", ev.Seq, ev.ObservedTime)
	if r.seen[key] {
		return false
	}
	r.seen[key] = true
	r.recordsSeen++
	return true
}

func (r *Runner) seriesFor(entity, channel string) *series {
	key := entity + "\x00" + channel
	s := r.stats[key]
	if s == nil {
		s = &series{}
		if r.stats == nil {
			r.stats = map[string]*series{}
		}
		r.stats[key] = s
	}
	return s
}

func recordArrivalGap(s *series, at int64) {
	if s.seen && at > s.lastEmitNS {
		s.gaps = append(s.gaps, at-s.lastEmitNS)
		if len(s.gaps) > 60 {
			s.gaps = s.gaps[1:]
		}
		s.silenceReported = false // a new sample ends any quiet episode
	}
}

func markPresent(s *series, at int64) {
	s.lastEmitNS = at
	s.seen = true
}

func (r *Runner) appendReading(s *series, value float64, at int64) {
	s.window = append(s.window, value)
	if len(s.window) > r.cfg.Window {
		s.window = s.window[1:]
	}
	markPresent(s, at)
}

func (r *Runner) gradeReading(ev model.SimEvent, s *series, at int64, suspicious bool) {
	if !suspicious {
		s.suspicious = 0
		return
	}
	s.suspicious++
	if s.suspicious >= r.cfg.MinConsecutive {
		r.detectConsecutive(ev, s, at)
	}
}

func (r *Runner) detectConsecutive(ev model.SimEvent, s *series, at int64) {
	s.lastSeq = ev.Seq
	detection := r.detect(ev.EntityID, ev.Channel, at, s)
	r.actuateDetection(ev.EntityID, at)
	// Rebuild the baseline from post-anomaly readings.
	s.suspicious = 0
	s.window = nil
	r.detections = append(r.detections, detection)
}

func (r *Runner) actuateDetection(entity string, at int64) {
	if r.cfg.OnDetectionEffector != "" && !r.actuated[entity] {
		command := r.issue(entity, at)
		r.actions = append(r.actions, command)
		r.actuated[entity] = true
	}
}

func (r *Runner) sortedSeriesKeys() []string {
	keys := make([]string, 0, len(r.stats))
	// determinism-safe: sort keys before producing detections.
	for key := range r.stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (r *Runner) detectSeriesSilence(key string, s *series, now int64) {
	if !s.seen || len(s.gaps) == 0 || s.silenceReported {
		return
	}
	parts := splitKey(key)
	period := median(s.gaps)
	if now-s.lastEmitNS > int64(r.cfg.AbsenceFactor*float64(period)) {
		r.detections = append(r.detections, silenceDetection(parts[0], now, s.lastSeq))
		s.silenceReported = true
	}
}

func silenceDetection(entity string, at int64, seq int64) model.Detection {
	return model.Detection{EntityID: entity, DetectedAt: model.FormatTime(at),
		Confidence: 0.9, Narrative: "channel silence",
		EvidenceRefs: []string{fmt.Sprintf("seq:%d", seq)}}
}

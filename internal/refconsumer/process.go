package refconsumer

import (
	"fmt"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Process consumes one batch of native-format JSONL sim events (delivery
// order) and returns the consumer's cumulative verdict. endNS is the world
// instant the batch covers, used for the quiescence watermark and absence
// detection. Sequences already processed by an earlier call are skipped, so
// a harness may re-feed the full trace without double-counting.
func (r *Runner) Process(trace []byte, endNS int64) (*model.Verdict, error) {
	now := max(r.processTrace(trace), endNS)
	r.detectSilence(now)
	return r.reportVerdict(now)
}

func (r *Runner) processTrace(trace []byte) int64 {
	var now int64
	for _, line := range splitLines(trace) {
		var ev model.SimEvent
		if err := model.DecodeBytes([]byte(line), &ev); err != nil {
			// A malformed record is itself evidence: report it and move on.
			continue
		}
		now = max(now, r.consumeEvent(ev))
	}
	return now
}

func (r *Runner) consumeEvent(ev model.SimEvent) int64 {
	// Deduplicate only exact re-feeds of an already-processed delivery;
	// reordered and duplicated deliveries (same seq, different observed
	// time) are each counted, as a live consumer would count them.
	deliveryKey := fmt.Sprintf("%d@%s", ev.Seq, ev.ObservedTime)
	if r.seen[deliveryKey] {
		return 0
	}
	r.seen[deliveryKey] = true
	r.recordsSeen++
	t, _ := model.ParseTime(ev.ObservedTime)
	key := ev.EntityID + "\x00" + ev.Channel
	s := r.stats[key]
	if s == nil {
		s = &series{}
		if r.stats == nil {
			r.stats = map[string]*series{}
		}
		r.stats[key] = s
	}
	if s.seen && t > s.lastEmitNS {
		s.gaps = append(s.gaps, t-s.lastEmitNS)
		if len(s.gaps) > 60 {
			s.gaps = s.gaps[1:]
		}
		s.silenceReported = false // a new sample ends any quiet episode
	}
	r.evaluateReading(ev, s, t)
	return t
}

func (r *Runner) evaluateReading(ev model.SimEvent, s *series, t int64) {
	if v, ok := asFloat(ev.Value); ok {
		// Compare against the baseline BEFORE this reading: the window
		// must not be diluted by the anomaly it is meant to detect.
		suspicious := r.suspicious(s, v)
		s.window = append(s.window, v)
		if len(s.window) > r.cfg.Window {
			s.window = s.window[1:]
		}
		s.lastEmitNS = t
		s.seen = true
		if suspicious {
			s.suspicious++
			if s.suspicious >= r.cfg.MinConsecutive {
				s.lastSeq = ev.Seq
				det := r.detect(ev.EntityID, ev.Channel, t, s)
				if r.cfg.OnDetectionEffector != "" && !r.actuated[ev.EntityID] {
					cmd := r.issue(ev.EntityID, t)
					r.actions = append(r.actions, cmd)
					r.actuated[ev.EntityID] = true
				}
				// Rebuild the baseline from post-anomaly readings.
				s.suspicious = 0
				s.window = nil
				r.detections = append(r.detections, det)
			}
		} else {
			s.suspicious = 0
		}
	} else {
		// Heartbeats and strings refresh the presence marker.
		s.lastEmitNS = t
		s.seen = true
	}
}

func (r *Runner) detectSilence(now int64) {
	// determinism-safe: keys collected below and sorted, so silence
	// detections append in a stable order across processes.
	keys := make([]string, 0, len(r.stats))
	// determinism-safe
	for key := range r.stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		s := r.stats[key]
		if !s.seen || len(s.gaps) == 0 || s.silenceReported {
			continue
		}
		parts := splitKey(key)
		period := median(s.gaps)
		if now-s.lastEmitNS > int64(r.cfg.AbsenceFactor*float64(period)) {
			r.detections = append(r.detections, model.Detection{
				EntityID: parts[0], DetectedAt: model.FormatTime(now),
				Confidence: 0.9, Narrative: "channel silence",
				EvidenceRefs: []string{fmt.Sprintf("seq:%d", s.lastSeq)},
			})
			s.silenceReported = true
		}
	}
}

func (r *Runner) reportVerdict(now int64) (*model.Verdict, error) {
	// The consumer asserts quiescence through the instant it has fully
	// processed: the later of the last observed record and the declared end
	// of the window. Without this assertion a harness cannot tell when a
	// closed-loop advance is reproducible.
	if r.quiescence != nil {
		if err := r.quiescence.ReportQuiesced(now); err != nil {
			return nil, fmt.Errorf("Process: report quiescence: %w", err)
		}
	}
	v := &model.Verdict{
		SchemaVersion: "0.1",
		RunID:         r.runID,
		Consumer: model.ConsumerInfo{
			Name: "streamsim-refconsumer", Version: "0.1.0",
			ConfigDigest: r.configDigest(),
		},
		Detections: r.detections,
		Actions:    r.actions,
		Counters: map[string]int64{
			"records_seen": r.recordsSeen,
			"detections":   int64(len(r.detections)),
			"actions":      int64(len(r.actions)),
		},
	}
	if err := r.report.SubmitVerdict(v); err != nil {
		return nil, fmt.Errorf("Process: %w", err)
	}
	return v, nil
}

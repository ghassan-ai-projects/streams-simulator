// Package refconsumer is the reference consumer: a deliberately simple
// moving-window detector that closes the loop against the simulator the way
// a real consumer would. It is config-driven (threshold, window, which
// effector to actuate) and carries no domain knowledge of its own — the
// same binary serves as a baseline for any domain. It never reads the
// ledger, never reads ground truth, and treats every actuator ack as
// provisional.
package refconsumer

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// EffectorInvoker is the narrow surface a consumer actuates through
// (satisfied by the simulator's run and the operator view).
type EffectorInvoker interface {
	InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error)
}

// VerdictSink accepts the final report.
type VerdictSink interface {
	SubmitVerdict(v *model.Verdict) error
}

// QuiescenceReporter asserts the consumer's quiescence watermark. It is
// satisfied by sinks that report through the MCP operator surface; run-based
// sinks whose ReportQuiesced returns nothing are bridged automatically.
type QuiescenceReporter interface {
	ReportQuiesced(throughNS int64) error
}

// quiescenceBridge adapts a sink with a void ReportQuiesced.
type quiescenceBridge struct{ q func(throughNS int64) }

func (b quiescenceBridge) ReportQuiesced(throughNS int64) error {
	b.q(throughNS)
	return nil
}

// Nameplate is the static world description the consumer is given (subset
// of the simulator's nameplate).
type Nameplate struct {
	WorldID   string
	Entities  []EntityInfo
	Channels  []ChannelInfo
	Effectors []EffectorInfo
}

// EntityInfo describes one entity.
type EntityInfo struct {
	ID   string
	Type string
}

// ChannelInfo describes one channel.
type ChannelInfo struct {
	Name       string
	Unit       string
	RangeMin   float64
	RangeMax   float64
	Resolution float64
}

// EffectorInfo describes one effector.
type EffectorInfo struct {
	Name   string
	Schema map[string]any
}

// Config tunes the detector. Everything here is consumer configuration,
// reported in the verdict for reproducibility.
type Config struct {
	Threshold           float64 // z-score at which a reading is suspicious
	Window              int     // trailing window for mean/std
	MinConsecutive      int     // consecutive suspicious readings before declaring
	OnDetectionEffector string  // effector to actuate on detection ("" = observe only)
	AbsenceFactor       float64 // silence = this many channel periods before flagging
}

// DefaultConfig is a reasonable baseline.
func DefaultConfig() Config {
	return Config{Threshold: 4, Window: 30, MinConsecutive: 3, AbsenceFactor: 3}
}

// Runner executes the consumer over a native-format trace. State persists
// across Process calls, so a harness can feed one batch per advance: the
// consumer reacts between advances exactly as a live consumer would, and
// repeated delivery of an already-processed sequence number is skipped.
type Runner struct {
	cfg         Config
	np          *Nameplate
	invoker     EffectorInvoker
	report      VerdictSink
	quiescence  QuiescenceReporter
	runID       string
	commandSeq  int
	actuated    map[string]bool
	stats       map[string]*series
	seen        map[string]bool // (seq, observed_time) deliveries already processed
	detections  []model.Detection
	actions     []model.Action
	recordsSeen int64
}

// New builds a consumer runner. If the report sink also reports quiescence,
// Process asserts the consumer's watermark through it after each run.
func New(cfg Config, np *Nameplate, invoker EffectorInvoker, report VerdictSink, runID string) *Runner {
	r := &Runner{cfg: cfg, np: np, invoker: invoker, report: report, runID: runID, actuated: map[string]bool{}, seen: map[string]bool{}}
	if q, ok := report.(QuiescenceReporter); ok {
		r.quiescence = q
	} else if q, ok := report.(interface{ ReportQuiesced(int64) }); ok {
		r.quiescence = quiescenceBridge{q: q.ReportQuiesced}
	}
	return r
}

// series is one entity/channel's running statistics.
type series struct {
	window          []float64
	gaps            []int64 // observed inter-arrival gaps, for silence detection
	lastEmitNS      int64
	lastSeq         int64
	suspicious      int
	seen            bool
	silenceReported bool // one silence detection per quiet episode
}

// Process consumes one batch of native-format JSONL sim events (delivery
// order) and returns the consumer's cumulative verdict. endNS is the world
// instant the batch covers, used for the quiescence watermark and absence
// detection. Sequences already processed by an earlier call are skipped, so
// a harness may re-feed the full trace without double-counting.
func (r *Runner) Process(trace []byte, endNS int64) (*model.Verdict, error) {
	now := int64(0)

	for _, line := range splitLines(trace) {
		var ev model.SimEvent
		if err := model.DecodeBytes([]byte(line), &ev); err != nil {
			// A malformed record is itself evidence: report it and move on.
			continue
		}
		// Deduplicate only exact re-feeds of an already-processed delivery;
		// reordered and duplicated deliveries (same seq, different observed
		// time) are each counted, as a live consumer would count them.
		deliveryKey := fmt.Sprintf("%d@%s", ev.Seq, ev.ObservedTime)
		if r.seen[deliveryKey] {
			continue
		}
		r.seen[deliveryKey] = true
		r.recordsSeen++
		t, _ := model.ParseTime(ev.ObservedTime)
		if t > now {
			now = t
		}
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
	// Absence: a channel that was present and went quiet for several of its
	// own observed periods (the cadence is measured, not assumed).
	if endNS > now {
		now = endNS
	}
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

func (r *Runner) suspicious(s *series, v float64) bool {
	if len(s.window) < 5 {
		return false
	}
	mean, std := meanStd(s.window)
	if std < 1e-12 {
		return false
	}
	return math.Abs(v-mean) > r.cfg.Threshold*std
}

func (r *Runner) detect(entity, channel string, t int64, s *series) model.Detection {
	// Evidence refs: the scoring side validates citations against delivered
	// records; the consumer cites the channel it detected on.
	return model.Detection{
		EntityID:     entity,
		DetectedAt:   model.FormatTime(t),
		Confidence:   0.8,
		Narrative:    fmt.Sprintf("%s deviated beyond %.1f sigma of its %d-reading window", channel, r.cfg.Threshold, len(s.window)),
		EvidenceRefs: []string{fmt.Sprintf("seq:%d", s.lastSeq)},
	}
}

func (r *Runner) issue(entity string, t int64) model.Action {
	r.commandSeq++
	cmdID := fmt.Sprintf("rc-%05d", r.commandSeq)
	if r.invoker != nil {
		if _, err := r.invoker.InvokeEffector(r.cfg.OnDetectionEffector, entity, cmdID, r.argsFor(entity), t); err != nil {
			return model.Action{
				CommandID: cmdID, Effector: r.cfg.OnDetectionEffector, EntityID: entity,
				IssuedAt: model.FormatTime(t), OutcomeBelieved: model.BelievedFailed,
			}
		}
	}
	return model.Action{
		CommandID: cmdID, Effector: r.cfg.OnDetectionEffector, EntityID: entity,
		IssuedAt: model.FormatTime(t), OutcomeBelieved: model.BelievedUnknown,
	}
}

// argsFor constructs a minimal argument set for the detection effector:
// required string properties named *_id are bound to the entity id. This is
// consumer configuration made from the nameplate, not simulator knowledge.
func (r *Runner) argsFor(entity string) map[string]any {
	out := map[string]any{}
	var schema map[string]any
	for _, e := range r.np.Effectors {
		if e.Name == r.cfg.OnDetectionEffector {
			if e.Schema != nil {
				schema = e.Schema
			}
		}
	}
	if schema == nil {
		return out
	}
	if req, ok := schema["required"].([]any); ok {
		for _, rq := range req {
			name, _ := rq.(string)
			if name == "" {
				continue
			}
			if props, ok := schema["properties"].(map[string]any); ok {
				if p, ok := props[name].(map[string]any); ok {
					if p["type"] == "string" && len(name) >= 3 && name[len(name)-3:] == "_id" {
						out[name] = entity
					}
				}
			}
		}
	}
	return out
}

func (r *Runner) configDigest() string {
	d, _ := canonical.Digest(map[string]any{
		"threshold": r.cfg.Threshold, "window": r.cfg.Window,
		"min_consecutive": r.cfg.MinConsecutive, "effector": r.cfg.OnDetectionEffector,
		"absence_factor": r.cfg.AbsenceFactor,
	})
	return d
}

func meanStd(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	var sq float64
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(sq / float64(len(xs)))
}

func splitLines(b []byte) []string {
	var out []string
	start := 0
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			out = append(out, string(b[start:i]))
			start = i + 1
		}
	}
	if start < len(b) {
		out = append(out, string(b[start:]))
	}
	return out
}

func asFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	}
	return 0, false
}

func median(xs []int64) int64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]int64{}, xs...)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	return sorted[len(sorted)/2]
}

func splitKey(key string) []string {
	for i := 0; i < len(key); i++ {
		if key[i] == 0 {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key}
}

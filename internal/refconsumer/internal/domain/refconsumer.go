// Package domain holds the reference consumer's rules: a deliberately simple
// moving-window detector, its statistics and series handling, the verdict it
// builds and the Runner that closes the loop through narrow ports. It carries
// no domain knowledge, performs no I/O and never reads the ledger or ground
// truth.
package domain

import (
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

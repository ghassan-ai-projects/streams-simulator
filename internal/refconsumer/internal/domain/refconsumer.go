// Package domain holds the reference consumer's rules: a deliberately simple
// moving-window detector, its statistics and series handling, the verdict it
// builds and the Runner that closes the loop through narrow ports. It carries
// no domain knowledge, performs no I/O and never reads the ledger or ground
// truth.
package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type EffectorInvoker interface {
	InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*model.EffectorAcknowledgement, error)
}

type VerdictSink interface {
	SubmitVerdict(v *model.Verdict) error
}

type QuiescenceReporter interface {
	ReportQuiesced(throughNS int64) error
}

type quiescenceBridge struct{ q func(throughNS int64) }

func (b quiescenceBridge) ReportQuiesced(throughNS int64) error {
	b.q(throughNS)
	return nil
}

type Nameplate struct {
	WorldID   string
	Entities  []EntityInfo
	Channels  []ChannelInfo
	Effectors []EffectorInfo
}

type EntityInfo struct {
	ID   string
	Type string
}

type ChannelInfo struct {
	Name       string
	Unit       string
	RangeMin   float64
	RangeMax   float64
	Resolution float64
}

type EffectorInfo struct {
	Name   string
	Schema map[string]any
}

type Config struct {
	Threshold           float64
	Window              int
	MinConsecutive      int
	OnDetectionEffector string
	AbsenceFactor       float64
}

func DefaultConfig() Config {
	return Config{Threshold: 4, Window: 30, MinConsecutive: 3, AbsenceFactor: 3}
}

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
	seen        map[string]bool
	detections  []model.Detection
	actions     []model.Action
	recordsSeen int64
}

func New(cfg Config, np *Nameplate, invoker EffectorInvoker, report VerdictSink, runID string) *Runner {
	r := &Runner{cfg: cfg, np: np, invoker: invoker, report: report, runID: runID, actuated: map[string]bool{}, seen: map[string]bool{}}
	if q, ok := report.(QuiescenceReporter); ok {
		r.quiescence = q
	} else if q, ok := report.(interface{ ReportQuiesced(int64) }); ok {
		r.quiescence = quiescenceBridge{q: q.ReportQuiesced}
	}
	return r
}

type series struct {
	window          []float64
	gaps            []int64
	lastEmitNS      int64
	lastSeq         int64
	suspicious      int
	seen            bool
	silenceReported bool
}

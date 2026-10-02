package run

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
	"os"
	"path/filepath"
	"strconv"
)

// New creates a run: world, perturbation layer, adapter engine and sink.
func New(ctx context.Context, cfg Config) (*Run, error) {
	cfg = defaultConfig(cfg)
	w, err := world.New(cfg.Domain, cfg.Seed, cfg.WorldID, cfg.StartTimeNS, world.Options{
		InitialEntities:  cfg.EntityIDs,
		Noiseless:        cfg.Noiseless,
		ForceFailureMode: cfg.ForceFailureMode,
	})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	r := &Run{
		ID:               cfg.RunID,
		Config:           cfg,
		World:            w,
		Perturb:          perturb.New(w.ID, cfg.Seed, cfg.Domain),
		ctx:              ctx,
		worldStartTimeNS: cfg.StartTimeNS,
		worldEndTimeNS:   cfg.StartTimeNS,
		envTargets:       map[string]string{},
		quiesceNotify:    make(chan struct{}),
	}
	if err := r.openLedger(); err != nil {
		return nil, err
	}
	if err := r.openAdapter(); err != nil {
		return nil, err
	}
	if err := r.openSink(); err != nil {
		return nil, err
	}
	if err := r.beginTrace(); err != nil {
		return nil, err
	}
	// Reproducible unless the wall clock drives delivery.
	r.reproducible = cfg.TimeMode != model.TimeWall

	w.SetEmitter(r.onEmit)
	return r, nil
}

func defaultConfig(cfg Config) Config {
	if cfg.SinkName == "" {
		cfg.SinkName = model.SinkFile
	}
	if cfg.TimeMode == "" {
		cfg.TimeMode = model.TimeStepped
	}
	if cfg.StartTimeNS == 0 && !cfg.StartTimeSet {
		cfg.StartTimeNS = model.DefaultStartTimeNS
	}
	if cfg.RunID == "" {
		cfg.RunID = "r-" + strconv.FormatUint(canonicalHash(cfg.Domain.Spec.ID, cfg.Seed), 36)
	}
	if cfg.QuiescenceClock == nil {
		cfg.QuiescenceClock = realQuiescenceClock{}
	}
	return cfg
}

func (r *Run) openLedger() error {
	if r.Config.LedgerPath != "" {
		if err := os.MkdirAll(filepath.Dir(r.Config.LedgerPath), 0o700); err != nil {
			return fmt.Errorf("run: ledger dir: %w", err)
		}
		lf, err := os.OpenFile(r.Config.LedgerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return fmt.Errorf("run: open ledger %s: %w", r.Config.LedgerPath, err)
		}
		r.ledgerWriter = bufio.NewWriter(lf)
		r.ledgerFile = lf
	}
	return nil
}

func (r *Run) openAdapter() error {
	meta := map[string]any{
		"run_id": r.Config.RunID, "sim_version": model.SimVersion,
		"domain_id": r.Config.Domain.Spec.ID, "domain_version": r.Config.Domain.Spec.Version,
		"world_start_time": model.FormatTime(r.Config.StartTimeNS), "seed": r.Config.Seed,
	}
	eng, err := adapter.NewEngine(r.Config.Adapter, meta)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	r.Engine = eng
	return nil
}

func (r *Run) openSink() error {

	switch r.Config.SinkName {
	case model.SinkInproc:
		r.Sink = &sink.Inproc{}
	case model.SinkFile:
		f, err := sink.NewFile(r.Config.SinkTarget)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		r.Sink = f
	case model.SinkHTTPPush:
		if r.Config.SinkTarget == "" {
			return fmt.Errorf("run: http-push sink requires a URL")
		}
		r.Sink = sink.NewHTTPPush(r.ctx, r.Config.SinkTarget)
	default:
		return fmt.Errorf("run: unsupported sink %q", r.Config.SinkName)
	}
	return nil
}

func (r *Run) beginTrace() error {

	if lines, err := r.Engine.Begin(); err != nil {
		return fmt.Errorf("run: adapter begin: %w", err)
	} else if err := writeSinkLines(r.Sink, lines); err != nil {
		return fmt.Errorf("run: adapter preamble: %w", err)
	}
	return nil
}

// canonicalHash derives a short stable id from the domain and seed.
func canonicalHash(domainID string, seed uint64) uint64 {
	b, _ := canonical.MarshalString(map[string]any{"d": domainID, "s": seed})
	return fnv([]byte(b))
}

func fnv(b []byte) uint64 {
	h := uint64(14695981039346656037)
	for _, c := range b {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}

// Trace returns the delivered trace bytes (available after End).
func (r *Run) Trace() []byte { return append([]byte(nil), r.trace...) }

// SetEvidenceRecorder installs a hook invoked for every delivered event
// (post-perturbation, pre-render). The prefix-indistinguishability harness
// uses it to compare what consumers would actually see.
func (r *Run) SetEvidenceRecorder(fn func(model.SimEvent)) {
	r.evidenceRec = fn
}

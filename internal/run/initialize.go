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
	if cfg.LedgerPath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.LedgerPath), 0o700); err != nil {
			return nil, fmt.Errorf("run: ledger dir: %w", err)
		}
		lf, err := os.OpenFile(cfg.LedgerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return nil, fmt.Errorf("run: open ledger %s: %w", cfg.LedgerPath, err)
		}
		r.ledgerWriter = bufio.NewWriter(lf)
		r.ledgerFile = lf
	}
	meta := map[string]any{
		"run_id": cfg.RunID, "sim_version": model.SimVersion,
		"domain_id": cfg.Domain.Spec.ID, "domain_version": cfg.Domain.Spec.Version,
		"world_start_time": model.FormatTime(cfg.StartTimeNS), "seed": cfg.Seed,
	}
	eng, err := adapter.NewEngine(cfg.Adapter, meta)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	r.Engine = eng

	switch cfg.SinkName {
	case model.SinkInproc:
		r.Sink = &sink.Inproc{}
	case model.SinkFile:
		f, err := sink.NewFile(cfg.SinkTarget)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		r.Sink = f
	case model.SinkHTTPPush:
		if cfg.SinkTarget == "" {
			return nil, fmt.Errorf("run: http-push sink requires a URL")
		}
		r.Sink = sink.NewHTTPPush(ctx, cfg.SinkTarget)
	default:
		return nil, fmt.Errorf("run: unsupported sink %q", cfg.SinkName)
	}
	if lines, err := eng.Begin(); err != nil {
		return nil, fmt.Errorf("run: adapter begin: %w", err)
	} else if err := writeSinkLines(r.Sink, lines); err != nil {
		return nil, fmt.Errorf("run: adapter preamble: %w", err)
	}
	// Reproducible unless the wall clock drives delivery.
	r.reproducible = cfg.TimeMode != model.TimeWall

	w.SetEmitter(r.onEmit)
	return r, nil
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

package run

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// New creates a run: world, perturbation layer, adapter engine and sink.
func New(ctx context.Context, cfg Config) (*Run, error) {
	r, err := newRun(ctx, defaultConfig(cfg))
	if err != nil {
		return nil, err
	}
	if err := r.openPipeline(); err != nil {
		return nil, err
	}
	r.attachWorld()
	return r, nil
}

func newRun(ctx context.Context, cfg Config) (*Run, error) {
	w, err := newRunWorld(cfg)
	if err != nil {
		return nil, err
	}
	return newRunState(ctx, cfg, w)
}

func newRunWorld(cfg Config) (*world.World, error) {
	w, err := world.New(cfg.Domain, cfg.Seed, cfg.WorldID, cfg.StartTimeNS, world.Options{
		InitialEntities:  cfg.EntityIDs,
		Noiseless:        cfg.Noiseless,
		ForceFailureMode: cfg.ForceFailureMode,
	})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return w, nil
}

func newRunState(ctx context.Context, cfg Config, w *world.World) (*Run, error) {
	layer, err := perturb.New(w.ID, cfg.Seed, cfg.Domain)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return runWith(ctx, cfg, w, layer), nil
}

func runWith(ctx context.Context, cfg Config, w *world.World, layer *perturb.Layer) *Run {
	return &Run{
		ID:               cfg.RunID,
		Config:           cfg,
		World:            w,
		Perturb:          layer,
		ctx:              ctx,
		worldStartTimeNS: cfg.StartTimeNS,
		worldEndTimeNS:   cfg.StartTimeNS,
		envTargets:       map[string]string{},
		quiesceNotify:    make(chan struct{}),
	}
}

func (r *Run) openPipeline() error {
	if err := r.openLedger(); err != nil {
		return err
	}
	if err := r.openAdapter(); err != nil {
		return err
	}
	if err := r.openSink(); err != nil {
		return err
	}
	return r.beginTrace()
}

func (r *Run) attachWorld() {
	// Reproducible unless the wall clock drives delivery.
	r.reproducible = r.Config.TimeMode != model.TimeWall
	r.World.SetEmitter(r.onEmit)
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
	return defaultRunIdentity(cfg)
}

func defaultRunIdentity(cfg Config) Config {
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
		return r.openFileSink()
	case model.SinkHTTPPush:
		return r.openHTTPPushSink()
	default:
		return fmt.Errorf("run: unsupported sink %q", r.Config.SinkName)
	}
	return nil
}

func (r *Run) openFileSink() error {
	file, err := sink.NewFile(r.Config.SinkTarget)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	r.Sink = file
	return nil
}

func (r *Run) openHTTPPushSink() error {
	if r.Config.SinkTarget == "" {
		return fmt.Errorf("run: http-push sink requires a URL")
	}
	r.Sink = sink.NewHTTPPush(r.ctx, r.Config.SinkTarget)
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

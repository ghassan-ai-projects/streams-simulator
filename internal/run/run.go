// Package run wires the pipeline world -> perturbation -> adapter -> sink ->
// ledger into one deterministic entity, records every mutation in a command
// log, and can collapse the whole run into a run artifact that replay
// reproduces byte-for-byte with no server running.
//
// A run is a pure function of (sim_version, domain_digest, adapter_digest,
// seed, command_log, sink).
package run

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Config pins everything that enters the determinism tuple.
type Config struct {
	Domain           *domain.Compiled
	Adapter          *model.Adapter
	Seed             uint64
	SinkName         string
	SinkTarget       string // file path or http-push URL
	TimeMode         string
	StartTimeNS      int64
	StartTimeSet     bool // true when StartTimeNS was supplied (0 is a legal start)
	EntityIDs        []string
	ScenarioProfile  string
	ClockMultiplier  float64
	Label            string
	RunID            string
	WorldID          string
	Noiseless        bool
	ForceFailureMode string // force every effector into a failure mode (tests)
	QuiescenceClock  QuiescenceClock
	LedgerPath       string // append-only durable ledger path ("" = in-memory only until End)
}

// DefaultQuiescenceTimeout bounds an await_consumer wait before the run is
// marked incomplete. The duration is fixed; the clock is injectable so
// deterministic tests never sleep.
const DefaultQuiescenceTimeout = 30 * time.Second

// QuiescenceClock supplies the deadline for await_consumer waits. The real
// implementation is a wall-clock timer; tests inject a fake they can fire.
type QuiescenceClock interface {
	NewTimer(time.Duration) QuiescenceTimer
}

// QuiescenceTimer is one deadline from a QuiescenceClock.
type QuiescenceTimer interface {
	C() <-chan time.Time
	Stop() bool
}

// realQuiescenceClock is the default quiescence clock.
type realQuiescenceClock struct{}

type realQuiescenceTimer struct{ t *time.Timer }

func (rt realQuiescenceTimer) C() <-chan time.Time { return rt.t.C }
func (rt realQuiescenceTimer) Stop() bool          { return rt.t.Stop() }

func (realQuiescenceClock) NewTimer(d time.Duration) QuiescenceTimer {
	return realQuiescenceTimer{t: time.NewTimer(d)}
}

// ErrConsumerNotQuiesced marks an await_consumer timeout. The world has
// already advanced; the run is incomplete, never silently successful.
var ErrConsumerNotQuiesced = errors.New("consumer_not_quiesced")

// Run is one deterministic execution.
type Run struct {
	ID      string
	Config  Config
	World   *world.World
	Perturb *perturb.Layer
	Engine  *adapter.Engine
	Sink    sink.Sink
	ctx     context.Context

	commandLog        []model.Command
	perturbHistory    []string
	ledger            []model.LedgerRecord
	history           []stateSnapshot
	verdict           *model.Verdict
	quiescedThroughNS int64
	commandMu         sync.Mutex // serializes world-mutating commands across goroutines
	quiesceMu         sync.Mutex
	quiesceNotify     chan struct{}
	evidenceRec       func(model.SimEvent) // delivered-event hook (test harness)
	quiesceParked     func()               // fired when a quiescence wait blocks (test harness)
	ledgerWriter      *bufio.Writer
	ledgerFile        *os.File

	finished     bool
	incomplete   bool
	runErr       error
	trace        []byte
	traceDigest  string
	reproducible bool
	unblinded    bool
	unblindedAt  string

	worldStartTimeNS int64
	worldEndTimeNS   int64

	envTargets map[string]string
	allowEnv   bool
}

type stateSnapshot struct {
	Seq    int64              `json:"seq"`
	TimeNS int64              `json:"time_ns"`
	Entity string             `json:"entity_id"`
	States map[string]float64 `json:"states"`
}

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

// onEmit is the world's emitter: perturb -> adapter -> sink -> ledger.
func (r *Run) onEmit(ev model.SimEvent) {
	atNS, _ := model.ParseTime(ev.EventTime)
	// World-state history: what was actually happening when the record was
	// emitted (director-only, for post-hoc analysis).
	states := map[string]float64{}
	for _, name := range r.Config.Domain.StateNames() {
		states[name] = r.World.StateValue(ev.EntityID, name, atNS)
	}
	r.history = append(r.history, stateSnapshot{Seq: ev.Seq, TimeNS: atNS, Entity: ev.EntityID, States: states})
	recs := r.Perturb.Process(ev, atNS)
	for _, d := range recs {
		if d.Malformed {
			// One bad record must not poison a file: render a broken line.
			if err := r.writeMalformed(ev); err != nil {
				r.appendLedger(model.LedgerRecord{
					DeliveryID: d.DeliveryID, Seq: ev.Seq, WorldID: ev.WorldID, EntityID: ev.EntityID,
					Channel: ev.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
					Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
				})
				r.fail(err)
				return
			}
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: ev.Seq, WorldID: ev.WorldID, EntityID: ev.EntityID,
				Channel: ev.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: true, DeliveryReason: model.DeliveryMangled,
				WrittenAtNS: r.World.Clock(),
			})
			continue
		}
		if !d.Delivered {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: ev.Seq, WorldID: ev.WorldID, EntityID: ev.EntityID,
				Channel: ev.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: d.Reason,
				WrittenAtNS: r.World.Clock(),
			})
			continue
		}
		if r.evidenceRec != nil {
			r.evidenceRec(d.Event)
		}
		line, err := r.Engine.RenderStreamRecord(&d.Event)
		if err != nil {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
				Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
			})
			r.fail(err)
			return
		}
		if line == "" {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
				Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: model.DeliveryOmitted, WrittenAtNS: r.World.Clock(),
			})
			continue
		}
		if err := r.Sink.Write([]byte(line)); err != nil {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
				Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
			})
			r.fail(err)
			return
		}
		otNS, _ := model.ParseTime(d.Event.ObservedTime)
		r.appendLedger(model.LedgerRecord{
			DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
			Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: otNS,
			Delivered: true, DeliveryReason: d.Reason, WrittenAtNS: r.World.Clock(),
		})
	}
}

func (r *Run) writeMalformed(ev model.SimEvent) error {
	// A broken line in the adapter's encoding: unterminated JSON.
	return r.Sink.Write([]byte(`{"seq":` + strconv.FormatInt(ev.Seq, 10) + `,"broken":`))
}

// fail aborts the run with an error; the run is marked incomplete.
func (r *Run) fail(err error) {
	if err == nil || r.runErr != nil {
		return
	}
	r.runErr = fmt.Errorf("run %s aborted: %w", r.ID, err)
	r.incomplete = true
	r.reproducible = false
}

// RenderRecord exposes the adapter's per-record rendering (used by the MCP
// trace export path and tests).
func (r *Run) RenderRecord(ev *model.SimEvent) (string, error) {
	line, err := r.Engine.RenderRecord(ev)
	if err != nil {
		return "", fmt.Errorf("run: render: %w", err)
	}
	return line, nil
}

// Advance moves the world to toNS, delivering everything along the way.
// awaitConsumer blocks until quiescence is reported through toNS; the wait
// is bounded by DefaultQuiescenceTimeout and canceled by ctx. On a
// quiescence timeout the world has already moved, so the advance is logged
// (replay reproduces the same state), the run is marked incomplete, and the
// emitted count is still returned alongside the error.
func (r *Run) Advance(ctx context.Context, toNS int64, awaitConsumer bool) (int, error) {
	r.commandMu.Lock()
	if r.runErr != nil {
		r.commandMu.Unlock()
		return 0, r.runErr
	}
	emitted, effects, err := r.World.Advance(toNS)
	if err != nil {
		r.commandMu.Unlock()
		return 0, fmt.Errorf("run: advance: %w", err)
	}
	r.worldEndTimeNS = toNS
	// Flush windowing perturbations (reorder, flaps, backfill) at the
	// boundary.
	for _, d := range r.Perturb.Flush(toNS) {
		r.deliver(d)
	}
	if r.runErr != nil {
		r.commandMu.Unlock()
		return emitted, r.runErr
	}
	// Log the advance before awaiting quiescence: the world has already
	// moved, and replay must reproduce exactly this state even when the
	// consumer never reports quiescence.
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(),
		Op: model.OpClockAdvance, Args: map[string]any{"to_ns": toNS, "await_consumer": awaitConsumer},
	})
	// Command boundary: the ledger rows and trace bytes for this advance are
	// now on file descriptors, so a crash here loses nothing acknowledged.
	if err := r.flushDurable(); err != nil {
		r.fail(err)
		r.commandMu.Unlock()
		return emitted, err
	}
	if !awaitConsumer {
		r.commandMu.Unlock()
		_ = effects
		return emitted, nil
	}
	// The quiescence wait is a consumer-sync barrier, not a world mutation:
	// release the command mutex so the consumer's effector call can land
	// while the advance waits. Without this the closed loop deadlocks.
	r.commandMu.Unlock()
	if err := r.awaitQuiescence(ctx, toNS); err != nil {
		// Only a timeout is a simulator failure: a canceled wait is a
		// caller-side abandonment and leaves the run open-loop. The failure
		// state is written under the command mutex so a concurrent End or
		// Score never reads it half-written.
		if errors.Is(err, ErrConsumerNotQuiesced) {
			r.commandMu.Lock()
			r.fail(err)
			r.commandMu.Unlock()
		}
		return emitted, err
	}
	_ = effects
	return emitted, nil
}

// appendLedger records one delivery row in memory and, when durable
// persistence is configured, appends it to the ledger file immediately. The
// file is flushed at command boundaries, so a crash between boundaries loses
// nothing that was acknowledged at a boundary.
func (r *Run) appendLedger(rec model.LedgerRecord) {
	r.ledger = append(r.ledger, rec)
	if r.ledgerWriter != nil {
		if raw, err := json.Marshal(rec); err == nil {
			_, _ = r.ledgerWriter.Write(raw)
			_ = r.ledgerWriter.WriteByte('\n')
		}
	}
}

// flushDurable pushes the ledger writer and the file sink (when present) to
// their file descriptors. Called at every command boundary; End adds fsync.
func (r *Run) flushDurable() error {
	if r.ledgerWriter != nil {
		if err := r.ledgerWriter.Flush(); err != nil {
			return fmt.Errorf("run: flush ledger: %w", err)
		}
	}
	if f, ok := r.Sink.(interface{ Flush() error }); ok {
		if err := f.Flush(); err != nil {
			return fmt.Errorf("run: flush sink: %w", err)
		}
	}
	return nil
}

func (r *Run) deliver(d perturb.Delivered) {
	if d.Malformed {
		if err := r.writeMalformed(d.Event); err != nil {
			atNS, _ := model.ParseTime(d.Event.EventTime)
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
				Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
			})
			r.fail(err)
			return
		}
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.appendLedger(model.LedgerRecord{
			DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
			Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
			Delivered: true, DeliveryReason: model.DeliveryMangled, WrittenAtNS: r.World.Clock(),
		})
		return
	}
	if !d.Delivered {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.appendLedger(model.LedgerRecord{
			DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
			Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
			Delivered: false, DeliveryReason: d.Reason, WrittenAtNS: r.World.Clock(),
		})
		return
	}
	line, err := r.Engine.RenderStreamRecord(&d.Event)
	if err != nil {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.appendLedger(model.LedgerRecord{
			DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
			Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
			Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
		})
		r.fail(err)
		return
	}
	if line == "" {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.appendLedger(model.LedgerRecord{
			DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
			Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
			Delivered: false, DeliveryReason: model.DeliveryOmitted, WrittenAtNS: r.World.Clock(),
		})
		return
	}
	if r.evidenceRec != nil {
		r.evidenceRec(d.Event)
	}
	if err := r.Sink.Write([]byte(line)); err != nil {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.appendLedger(model.LedgerRecord{
			DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
			Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
			Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
		})
		r.fail(err)
		return
	}
	atNS, _ := model.ParseTime(d.Event.EventTime)
	otNS, _ := model.ParseTime(d.Event.ObservedTime)
	r.appendLedger(model.LedgerRecord{
		DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
		Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: otNS,
		Delivered: true, DeliveryReason: d.Reason, WrittenAtNS: r.World.Clock(),
	})
}

// InjectFault records and applies a world fault.
func (r *Run) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	fid, err := r.World.InjectFault(entityID, faultID, onsetNS, params)
	if err != nil {
		return "", fmt.Errorf("run: inject fault: %w", err)
	}
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(), Op: model.OpFaultInject,
		Args: map[string]any{
			"entity_id": entityID, "fault": faultID, "onset_ns": onsetNS, "params": params,
		},
	})
	return fid, nil
}

// ClearFault records and clears a fault.
func (r *Run) ClearFault(faultID string, atNS int64) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	if err := r.World.ClearFault(faultID, atNS); err != nil {
		return fmt.Errorf("ClearFault: %w", err)
	}
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(), Op: model.OpFaultClear,
		Args: map[string]any{"fault_id": faultID, "at_ns": atNS},
	})
	return nil
}

// ApplyPerturb records and applies a delivery perturbation.
func (r *Run) ApplyPerturb(name string, params map[string]any, fromNS, untilNS int64) (string, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	id, err := r.Perturb.Apply(name, params, fromNS, untilNS)
	if err != nil {
		return "", fmt.Errorf("run: perturb: %w", err)
	}
	r.perturbHistory = append(r.perturbHistory, name)
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(), Op: model.OpPerturbApply,
		Args: map[string]any{
			"perturbation": name, "params": params, "from_ns": fromNS, "until_ns": untilNS,
		},
	})
	return id, nil
}

// ClearPerturb records and deactivates a perturbation.
func (r *Run) ClearPerturb(id string) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	if err := r.Perturb.Clear(id); err != nil {
		return fmt.Errorf("ClearPerturb: %w", err)
	}
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(), Op: model.OpPerturbClear,
		Args: map[string]any{"perturb_id": id},
	})
	return nil
}

// InvokeEffector records and performs an effector call.
func (r *Run) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if r.finished {
		return nil, fmt.Errorf("InvokeEffector: run is finished")
	}
	res, err := r.World.InvokeEffector(effector, entityID, commandID, args, atNS)
	if err != nil {
		// Interlock and effector refusals are recorded as commands too, so a
		// replay reproduces them.
		r.commandLog = append(r.commandLog, model.Command{
			Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(), Op: model.OpEffectorInvoke,
			Args: map[string]any{
				"effector": effector, "entity_id": entityID, "command_id": commandID,
				"args": args, "at_ns": atNS,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("InvokeEffector: %w", err)
		}
		return nil, nil
	}
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: r.World.Clock(), Op: model.OpEffectorInvoke,
		Args: map[string]any{
			"effector": effector, "entity_id": entityID, "command_id": commandID,
			"args": args, "at_ns": atNS,
		},
	})
	return res, nil
}

// AddEntity records and performs an entity birth.
func (r *Run) AddEntity(id string, atNS int64) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	if err := r.World.AddEntity(id, atNS, nil); err != nil {
		return fmt.Errorf("AddEntity: %w", err)
	}
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: atNS, Op: model.OpEntityAdd,
		Args: map[string]any{"entity_id": id, "at_ns": atNS},
	})
	return nil
}

// RetireEntity records and performs an entity retirement.
func (r *Run) RetireEntity(entityID, reason string, atNS int64) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	r.World.Retire(entityID, reason, atNS)
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: atNS, Op: model.OpEntityRetire,
		Args: map[string]any{"entity_id": entityID, "reason": reason, "at_ns": atNS},
	})
	return nil
}

// ConfigureEnvTarget records an env.inject target (pause/kill/partition).
func (r *Run) ConfigureEnvTarget(target string, allow bool) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.envTargets[target] = target
	r.allowEnv = r.allowEnv || allow
}

// EnvInject records an environment fault against a configured target. No
// environment-fault parameters are declared yet, so any params are rejected
// rather than recorded and ignored.
func (r *Run) EnvInject(target, fault string, params map[string]any, atNS int64) (string, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	if !r.allowEnv {
		return "", fmt.Errorf("run: env.inject not enabled for this world (no configured target)")
	}
	if len(params) > 0 {
		return "", fmt.Errorf("run: env fault %q accepts no parameters (got %d)", fault, len(params))
	}
	r.commandLog = append(r.commandLog, model.Command{
		Seq: int64(len(r.commandLog)), AtNS: atNS, Op: model.OpEnvInject,
		Args: map[string]any{"target": target, "fault": fault, "params": params, "at_ns": atNS},
	})
	return "env-" + strconv.Itoa(len(r.commandLog)), nil
}

// ReportQuiesced records the consumer's quiescence assertion. The watermark
// is monotonic, so a report can only move it forward: a stale report for an
// earlier instant can never satisfy a later await, and a report received
// between advances is honored on the next wait (the fast path).
func (r *Run) ReportQuiesced(throughNS int64) {
	r.quiesceMu.Lock()
	defer r.quiesceMu.Unlock()
	if throughNS > r.quiescedThroughNS {
		r.quiescedThroughNS = throughNS
		close(r.quiesceNotify)
		r.quiesceNotify = make(chan struct{})
	}
}

// SetQuiesceParkedHook installs a callback fired each time a quiescence wait
// is about to block (test harness; nil by default).
func (r *Run) SetQuiesceParkedHook(h func()) {
	r.quiesceParked = h
}

func (r *Run) awaitQuiescence(ctx context.Context, toNS int64) error {
	timer := r.Config.QuiescenceClock.NewTimer(DefaultQuiescenceTimeout)
	defer timer.Stop()
	for {
		r.quiesceMu.Lock()
		if r.quiescedThroughNS >= toNS {
			r.quiesceMu.Unlock()
			return nil
		}
		ch := r.quiesceNotify
		through := r.quiescedThroughNS
		r.quiesceMu.Unlock()
		if r.quiesceParked != nil {
			r.quiesceParked()
		}
		select {
		case <-ch:
			continue
		case <-ctx.Done():
			return fmt.Errorf("run: quiescence wait canceled: %w", ctx.Err())
		case <-timer.C():
			return fmt.Errorf("run: %w: quiesced through %d, asked for %d", ErrConsumerNotQuiesced, through, toNS)
		}
	}
}

// Domain exposes the compiled domain spec.
func (r *Run) Domain() *domain.Compiled { return r.Config.Domain }

// Digest is the world digest: over (sim_version, domain digest, seed,
// world config). Two creates with the same arguments must agree.
func (r *Run) Digest() string { return worldDigest(r) }

// SetFailureMode overrides the effector failure-mode distribution for
// subsequent invocations (test-only knob).
func (r *Run) SetFailureMode(mode string) {
	r.World.SetFailureMode(mode)
}

// UnblindedStamp reports whether the run was permanently stamped.
func (r *Run) UnblindedStamp() bool { return r.unblinded }

// AppliedPerturbations returns the perturbations applied during the run, in
// application order (the scorer's perturbation-fidelity input).
func (r *Run) AppliedPerturbations() []string {
	out := make([]string, len(r.perturbHistory))
	copy(out, r.perturbHistory)
	return out
}

// SubmitVerdict stores and validates a consumer verdict.
func (r *Run) SubmitVerdict(v *model.Verdict) error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if r.finished {
		// The operator endpoint outlives run.end; a verdict arriving then
		// must not be silently dropped from an already-written artifact.
		return fmt.Errorf("SubmitVerdict: run is finished")
	}
	if v == nil {
		return fmt.Errorf("SubmitVerdict: verdict is required")
	}
	if v.RunID != r.ID {
		return fmt.Errorf("run: verdict run_id %q does not match run %q", v.RunID, r.ID)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
	}
	if err := model.ValidateVerdict(raw); err != nil {
		return fmt.Errorf("SubmitVerdict: %w", err)
	}
	r.verdict = cloneVerdict(v)
	return nil
}

// Verdict returns the submitted verdict, or nil.
func (r *Run) Verdict() *model.Verdict { return cloneVerdict(r.verdict) }

// Ledger returns the delivery ledger in delivery order.
func (r *Run) Ledger() []model.LedgerRecord {
	return append([]model.LedgerRecord(nil), r.ledger...)
}

// TraceDigest is the sha256 of the delivered trace bytes.
func (r *Run) TraceDigest() string { return r.traceDigest }

// Unblind permanently stamps the run and excludes it from scorecards.
func (r *Run) Unblind() {
	r.unblinded = true
	r.unblindedAt = time.Now().UTC().Format(time.RFC3339Nano)
}

// Unblinded reports the stamp state.
func (r *Run) Unblinded() (bool, string) { return r.unblinded, r.unblindedAt }

// Reproducible reports whether the run is hash-reproducible.
func (r *Run) Reproducible() bool { return r.reproducible }

// End finalizes the run: closes the sink, computes the trace digest, and
// writes the run artifact, ledger, world-state history and (if any) verdict
// into outDir.
func (r *Run) End(outDir string) (*model.RunArtifact, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	if r.finished {
		return nil, fmt.Errorf("run: already finished")
	}
	if r.runErr == nil {
		if lines, err := r.Engine.End(r.worldEndTimeNS); err != nil {
			r.fail(fmt.Errorf("End: adapter postamble: %w", err))
		} else if err := writeSinkLines(r.Sink, lines); err != nil {
			r.fail(fmt.Errorf("End: adapter postamble: %w", err))
		}
	}
	trace, err := r.Sink.Close()
	if err != nil {
		return nil, fmt.Errorf("End: %w", err)
	}
	r.trace = trace
	r.traceDigest = canonical.DigestBytes(trace)
	r.finished = true

	if outDir != "" {
		if err := os.MkdirAll(outDir, 0o700); err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
		tracePath := filepath.Join(outDir, "trace.jsonl")
		if r.Config.SinkName == model.SinkFile && r.Config.SinkTarget != "" {
			tracePath = r.Config.SinkTarget
		}
		if err := os.WriteFile(tracePath, trace, 0o600); err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
		if r.ledgerFile != nil {
			// The durable ledger already holds every row; fsync it and
			// leave the file as the canonical ledger.
			if err := r.ledgerWriter.Flush(); err != nil {
				return nil, fmt.Errorf("End: flush ledger: %w", err)
			}
			if err := r.ledgerFile.Sync(); err != nil {
				return nil, fmt.Errorf("End: sync ledger: %w", err)
			}
		} else if err := writeJSONL(filepath.Join(outDir, "ledger.jsonl"), r.ledger); err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
		if err := writeJSONL(filepath.Join(outDir, "world_state_history.jsonl"), r.history); err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
		if r.verdict != nil {
			raw, _ := json.MarshalIndent(r.verdict, "", "  ")
			if err := os.WriteFile(filepath.Join(outDir, "verdict.json"), raw, 0o600); err != nil {
				return nil, fmt.Errorf("End: %w", err)
			}
		}
	}
	art := r.artifact()
	if outDir != "" {
		raw, err := json.MarshalIndent(art, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
		if err := os.WriteFile(filepath.Join(outDir, "run.json"), raw, 0o600); err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
	}
	if r.runErr != nil {
		return art, r.runErr
	}
	return art, nil
}

func writeSinkLines(dst sink.Sink, lines []string) error {
	for _, line := range lines {
		if err := dst.Write([]byte(line)); err != nil {
			return err
		}
	}
	return nil
}

// artifact assembles the run artifact from the run state.
func (r *Run) artifact() *model.RunArtifact {
	cmdLog := make([]model.Command, len(r.commandLog))
	copy(cmdLog, r.commandLog)
	sort.SliceStable(cmdLog, func(i, j int) bool { return cmdLog[i].Seq < cmdLog[j].Seq })
	counts := model.Counts{
		Emitted:         r.World.EmittedCount(),
		Perturbed:       int64(countLedger(r.ledger, model.DeliveryDuplicated, model.DeliveryDroppedByPerturb, model.DeliveryMangled, model.DeliveryDelayed, model.DeliveryRewritten, model.DeliveryReordered, model.DeliveryOmitted)),
		DroppedByDesign: int64(countLedger(r.ledger, model.DeliveryDroppedByPerturb)),
		EffectorCalls:   int64(len(r.World.EffectorCalls())),
		FaultsInjected:  int64(r.World.ActiveFaultsCount()),
	}
	return &model.RunArtifact{
		SchemaVersion: "0.1",
		SimVersion:    model.SimVersion,
		RunID:         r.ID,
		Label:         r.Config.Label,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Domain: model.ArtifactRef{
			ID: r.Config.Domain.Spec.ID, Version: r.Config.Domain.Spec.Version, Digest: r.Config.Domain.Digest,
		},
		DomainSpec: json.RawMessage(append([]byte(nil), r.Config.Domain.Raw...)),
		Seed:       r.Config.Seed,
		Sink:       r.Config.SinkName,
		Adapter: model.ArtifactRef{
			ID: r.Config.Adapter.ID, Version: r.Config.Adapter.Version, Digest: adapterDigest(r.Config.Adapter),
		},
		AdapterSpec: json.RawMessage(append([]byte(nil), r.Config.Adapter.Raw...)),
		TimeMode:    r.Config.TimeMode,
		WorldConfig: model.WorldConfig{
			StartTimeNS:     r.worldStartTimeNS,
			EntityIDs:       r.World.InitialEntityIDs(),
			ScenarioProfile: r.Config.ScenarioProfile,
			ClockMultiplier: r.Config.ClockMultiplier,
		},
		WorldDigest:         worldDigest(r),
		CommandLog:          cmdLog,
		ExpectedTraceDigest: r.traceDigest,
		Reproducible:        r.reproducible,
		Incomplete:          r.incomplete,
		Error:               errorString(r.runErr),
		Unblinded:           r.unblinded,
		UnblindedAt:         r.unblindedAt,
		Platform:            model.CurrentPlatform(),
		Counts:              counts,
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func countLedger(ledger []model.LedgerRecord, reasons ...string) int {
	n := 0
	for _, l := range ledger {
		for _, r := range reasons {
			if l.DeliveryReason == r {
				n++
				break
			}
		}
	}
	return n
}

func writeJSONL(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	defer func() { _ = f.Close() }()
	switch x := v.(type) {
	case []model.LedgerRecord:
		for _, rec := range x {
			raw, _ := json.Marshal(rec)
			if _, err := f.Write(append(raw, '\n')); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	case []stateSnapshot:
		for _, rec := range x {
			raw, _ := json.Marshal(rec)
			if _, err := f.Write(append(raw, '\n')); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	return nil
}

func adapterDigest(a *model.Adapter) string {
	raw, _ := json.Marshal(a)
	return canonical.DigestBytes(raw)
}

func worldDigest(r *Run) string {
	entityIDs := make([]any, 0, len(r.World.InitialEntityIDs()))
	for _, id := range r.World.InitialEntityIDs() {
		entityIDs = append(entityIDs, id)
	}
	raw := map[string]any{
		"sim_version":        model.SimVersion,
		"domain":             r.Config.Domain.Digest,
		"seed":               r.Config.Seed,
		"start_time_ns":      r.worldStartTimeNS,
		"entity_ids":         entityIDs,
		"scenario_profile":   r.Config.ScenarioProfile,
		"clock_multiplier":   r.Config.ClockMultiplier,
		"time_mode":          r.Config.TimeMode,
		"noiseless":          r.Config.Noiseless,
		"force_failure_mode": r.Config.ForceFailureMode,
	}
	d, err := canonical.Digest(raw)
	if err != nil {
		// Every value above is one of canonical's explicitly supported types.
		// Keep the failure visible if that invariant changes rather than
		// silently emitting an empty identity.
		return "invalid-world-digest:" + err.Error()
	}
	return d
}

// History returns the world-state history (director-only).
func (r *Run) History() []stateSnapshot {
	out := make([]stateSnapshot, len(r.history))
	for i, rec := range r.history {
		out[i] = rec
		out[i].States = maps.Clone(rec.States)
	}
	return out
}

// RecordHistory snapshots hidden state for post-hoc analysis.
func (r *Run) RecordHistory(seq int64, atNS int64, entity string, states map[string]float64) {
	r.history = append(r.history, stateSnapshot{Seq: seq, TimeNS: atNS, Entity: entity, States: maps.Clone(states)})
}

func cloneVerdict(v *model.Verdict) *model.Verdict {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	out := &model.Verdict{}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil
	}
	return out
}

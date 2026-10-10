// Package app orchestrates one deterministic run: it wires the pipeline
// world -> perturbation -> adapter -> sink -> ledger into one entity, records
// every mutation in a command log, collapses the run into a run artifact and
// replays an artifact byte-for-byte with no server running.
//
// A run is a pure function of (sim_version, domain_digest, adapter_digest,
// seed, command_log, sink).
package app

import (
	"context"
	"errors"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/quiesce"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Config pins everything that enters the determinism tuple.
type Config struct {
	Domain          *domain.Compiled
	Adapter         *model.Adapter
	Seed            uint64
	SinkName        string
	SinkTarget      string // file path or http-push URL
	TimeMode        string
	StartTimeNS     int64
	StartTimeSet    bool // true when StartTimeNS was supplied (0 is a legal start)
	EntityIDs       []string
	ScenarioProfile string
	// ClockMultiplier is part of the world identity and the artifact; no
	// stepped-time code reads it (reserved for scaled wall time).
	ClockMultiplier  float64
	Label            string
	RunID            string
	WorldID          string
	Noiseless        bool
	ForceFailureMode world.FailureMode // force every effector into a failure mode (tests)
	// quiescenceClock supplies the await_consumer deadline; only this layer's
	// tests replace the real clock, so it is not part of the public Config.
	quiescenceClock QuiescenceClock
	LedgerPath      string // append-only durable ledger path ("" = in-memory only until End)
}

// DefaultQuiescenceTimeout bounds an await_consumer wait before the run is
// marked incomplete.
const DefaultQuiescenceTimeout = quiesce.DefaultTimeout

// QuiescenceClock supplies the deadline for await_consumer waits. The real
// implementation is a wall-clock timer; tests inject a fake they can fire.
type QuiescenceClock = quiesce.Clock

// QuiescenceTimer is one deadline from a QuiescenceClock.
type QuiescenceTimer = quiesce.Timer

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
	history           []model.StateSnapshot
	verdict           *model.Verdict
	quiescedThroughNS int64
	commandMu         sync.Mutex // serializes world-mutating commands across goroutines
	quiesceMu         sync.Mutex
	quiesceNotify     chan struct{}
	evidenceRec       func(model.SimEvent) // delivered-event hook (test harness)
	quiesceParked     func()               // fired when a quiescence wait blocks (test harness)
	durableLedger     *durable.Ledger

	finished     bool
	incomplete   bool
	runErr       error
	trace        []byte
	traceDigest  string
	reproducible bool
	unblinded    bool
	unblindedAt  string

	worldStartTimeNS    int64
	worldEndTimeNS      int64
	lastObservedNS      int64
	hasObservedTime     bool
	lastTraceArrivalNS  int64
	hasTraceArrivalTime bool

	envTargets map[string]string
	allowEnv   bool
}

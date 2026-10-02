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
	"errors"
	"os"
	"sync"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
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

func (rt realQuiescenceTimer) Stop() bool { return rt.t.Stop() }

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

	worldStartTimeNS    int64
	worldEndTimeNS      int64
	lastObservedNS      int64
	hasObservedTime     bool
	lastTraceArrivalNS  int64
	hasTraceArrivalTime bool

	envTargets map[string]string
	allowEnv   bool
}

type stateSnapshot struct {
	Seq    int64              `json:"seq"`
	TimeNS int64              `json:"time_ns"`
	Entity string             `json:"entity_id"`
	States map[string]float64 `json:"states"`
}

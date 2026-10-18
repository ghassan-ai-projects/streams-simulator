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
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/quiesce"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type Config struct {
	Domain          *domain.Compiled
	Adapter         *model.Adapter
	Seed            uint64
	SinkName        string
	SinkTarget      string
	TimeMode        string
	StartTimeNS     int64
	StartTimeSet    bool
	EntityIDs       []string
	ScenarioProfile string

	ClockMultiplier  float64
	Label            string
	RunID            string
	WorldID          string
	Noiseless        bool
	ForceFailureMode world.FailureMode

	quiescenceClock QuiescenceClock
	LedgerPath      string

	JournalDir string

	Explore bool

	retainJournal bool

	silent bool
}

const DefaultQuiescenceTimeout = quiesce.DefaultTimeout

type QuiescenceClock = quiesce.Clock

type QuiescenceTimer = quiesce.Timer

var ErrDigestMismatch = errors.New("digest mismatch")

const MaxRunSpanNS = rules.MaxRunSpanNS

var ErrRunSpanExceeded = errors.New("the run's clock would move further than 30 days past its start")

var ErrUnknownProfile = errors.New("unknown profile")

var ErrConsumerNotQuiesced = errors.New("consumer_not_quiesced")

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
	sealedFaults      []model.FaultInterval
	ledger            []model.LedgerRecord
	history           []model.StateSnapshot
	verdict           *model.Verdict
	quiescedThroughNS int64
	commandMu         sync.Mutex
	quiesceMu         sync.Mutex
	quiesceNotify     chan struct{}
	evidenceRec       func(model.SimEvent)
	deliveryObserver  func(model.SimEvent)
	quiesceParked     func()
	durableLedger     *durable.Ledger
	journal           *journal.Journal
	journalDigests    journal.Digests
	replayedDigest    map[string]string

	finished     bool
	incomplete   bool
	unfinished   string
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

	fieldTraceDigest string
}

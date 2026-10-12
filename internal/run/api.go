package run

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Config pins everything that enters the determinism tuple: the domain, the
// adapter, the seed, the sink and the world start.
type Config = layer.Config

// ReplayResult is the outcome of replaying an artifact: whether the trace
// digest reproduced, the first divergence, and any version mismatch.
type ReplayResult = layer.ReplayResult

// ReplayEvidence is a replay outcome together with the effector calls the
// replayed world made, for offline scoring of the closed loop.
type ReplayEvidence = layer.ReplayEvidence

// ErrDigestMismatch marks a replay whose domain, adapter or world digest is not
// the one the artifact recorded.
var ErrDigestMismatch = layer.ErrDigestMismatch

// MaxRunSpanNS bounds how far a run's clock may move past its start: 30 days
// of simulated time. Every surface refuses an advance beyond it.
const MaxRunSpanNS = layer.MaxRunSpanNS

// ErrConsumerNotQuiesced marks an await_consumer timeout. The world has
// already advanced; the run is incomplete, never silently successful.
var ErrConsumerNotQuiesced = layer.ErrConsumerNotQuiesced

// WorldStatus is a consistent read of the world's clock and queues.
type WorldStatus = layer.WorldStatus

// Directories say where the domains and adapters a run names are looked up.
type Directories = layer.Directories

// ScenarioExport is a run expressed as a scenario: the scenario, the number of
// command-log entries it omits, the number of refused effector calls it drops,
// and its file content.
type ScenarioExport = layer.ScenarioExport

// ScenarioRunOptions say where a scenario's domain and adapter are looked up,
// which sink delivers the run and where its evidence is published (empty
// publishes nothing).
type ScenarioRunOptions = layer.ScenarioRunOptions

// RunSummary is what a finished scenario run left behind.
type RunSummary = layer.RunSummary

// ScenarioCheckError is the violation the static check or dry run found in a
// scenario, before any sink opened; Err is the violation itself.
type ScenarioCheckError = layer.ScenarioCheckError

// MaxStartNS is the latest a run's world may start; a start outside
// [0, MaxStartNS] is refused.
const MaxStartNS = world.MaxStartNS

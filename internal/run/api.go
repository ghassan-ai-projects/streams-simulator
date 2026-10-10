package run

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/app"

// Config pins everything that enters the determinism tuple: the domain, the
// adapter, the seed, the sink and the world start.
type Config = layer.Config

// ReplayResult is the outcome of replaying an artifact: whether the trace
// digest reproduced, the first divergence, and any version mismatch.
type ReplayResult = layer.ReplayResult

// ReplayEvidence is a replay outcome together with the effector calls the
// replayed world made, for offline scoring of the closed loop.
type ReplayEvidence = layer.ReplayEvidence

// ErrConsumerNotQuiesced marks an await_consumer timeout. The world has
// already advanced; the run is incomplete, never silently successful.
var ErrConsumerNotQuiesced = layer.ErrConsumerNotQuiesced

// WorldStatus is a consistent read of the world's clock and queues.
type WorldStatus = layer.WorldStatus

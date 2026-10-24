package run

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type Config = layer.Config

type ReplayResult = layer.ReplayResult

type ReplayEvidence = layer.ReplayEvidence

var ErrDigestMismatch = layer.ErrDigestMismatch

const MaxRunSpanNS = layer.MaxRunSpanNS

var ErrConsumerNotQuiesced = layer.ErrConsumerNotQuiesced

type WorldStatus = layer.WorldStatus

type Directories = layer.Directories

type ScenarioExport = layer.ScenarioExport

type ScenarioRunOptions = layer.ScenarioRunOptions

type RunSummary = layer.RunSummary

type ScenarioCheckError = layer.ScenarioCheckError

type ScenarioHooks = layer.ScenarioHooks

type ScenarioOutcome = layer.ScenarioOutcome

var (
	ErrFieldTrace          = layer.ErrFieldTrace
	ErrFieldTraceMissing   = layer.ErrFieldTraceMissing
	ErrFieldTraceDigest    = layer.ErrFieldTraceDigest
	ErrFieldTraceUncovered = layer.ErrFieldTraceUncovered
	ErrFieldTraceFull      = layer.ErrFieldTraceFull

	ErrInterruptLogMissing   = layer.ErrInterruptLogMissing
	ErrInterruptLogDigest    = layer.ErrInterruptLogDigest
	ErrInterruptLogUncovered = layer.ErrInterruptLogUncovered
)

const MaxStartNS = world.MaxStartNS

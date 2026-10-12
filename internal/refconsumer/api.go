package refconsumer

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
)

// Config tunes the reference detector: z-score threshold, trailing window,
// consecutive readings before declaring, the effector to actuate on detection
// and the silence factor.
type Config = layer.Config

// Nameplate is the static world description the consumer is given (a subset
// of the simulator's nameplate).
type Nameplate = layer.Nameplate

// EntityInfo describes one entity of the nameplate.
type EntityInfo = layer.EntityInfo

// ChannelInfo describes one channel of the nameplate.
type ChannelInfo = layer.ChannelInfo

// EffectorInfo describes one effector of the nameplate.
type EffectorInfo = layer.EffectorInfo

// EffectorInvoker is the narrow surface a consumer actuates through
// (satisfied by the simulator's run and the operator view).
type EffectorInvoker = layer.EffectorInvoker

// VerdictSink accepts the final report.
type VerdictSink = layer.VerdictSink

// QuiescenceReporter asserts the consumer's quiescence watermark.
type QuiescenceReporter = layer.QuiescenceReporter

// ConsumeOptions say which trace to consume, where the verdict goes, how the
// detector is tuned, and, when MCPEndpoint is set, the operator endpoint,
// capability token and run the consumer actuates and reports against.
type ConsumeOptions = app.Options

// ConsumeSummary is what a consumed trace left behind: the verdict file and
// how many detections it reports.
type ConsumeSummary = app.Summary

package refconsumer

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"

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

package refconsumer

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
)

type Config = layer.Config

type Nameplate = layer.Nameplate

type EntityInfo = layer.EntityInfo

type ChannelInfo = layer.ChannelInfo

type EffectorInfo = layer.EffectorInfo

type EffectorInvoker = layer.EffectorInvoker

type VerdictSink = layer.VerdictSink

type QuiescenceReporter = layer.QuiescenceReporter

type ConsumeOptions = app.Options

type ConsumeSummary = app.Summary

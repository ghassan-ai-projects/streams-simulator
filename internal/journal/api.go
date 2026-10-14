package journal

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
)

type Record = layer.Record

type Opened = layer.Opened

type Call = layer.Call

type Digests = layer.Digests

const (
	StreamObserver = layer.StreamObserver
	StreamDirector = layer.StreamDirector
)

const (
	ModeEvaluate = layer.ModeEvaluate
	ModeExplore  = layer.ModeExplore
)

const (
	ReasonExplore  = layer.ReasonExplore
	ReasonRevealed = layer.ReasonRevealed
)

const (
	ObserverFile = layer.ObserverFile
	DirectorFile = layer.DirectorFile
)

const (
	FramingPreamble  = layer.FramingPreamble
	FramingPostamble = layer.FramingPostamble
)

const (
	TypeRunOpened       = layer.TypeRunOpened
	TypeObservation     = layer.TypeObservation
	TypeClockAdvanced   = layer.TypeClockAdvanced
	TypeCommand         = layer.TypeCommand
	TypeEffectorOutcome = layer.TypeEffectorOutcome
	TypeKPITruth        = layer.TypeKPITruth
	TypeFaultTruth      = layer.TypeFaultTruth
)

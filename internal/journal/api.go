package journal

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
)

// Record is one decoded journal line: the envelope and the fields of its type.
type Record = layer.Record

// Opened is the identity of a run, the first record of the observer stream.
type Opened = layer.Opened

// Call names one effector call.
type Call = layer.Call

// Digests are the SHA-256 digests of a journal's two streams.
type Digests = layer.Digests

// The two streams.
const (
	StreamObserver = layer.StreamObserver
	StreamDirector = layer.StreamDirector
)

// Run modes recorded on run.opened.
const (
	ModeEvaluate = layer.ModeEvaluate
	ModeExplore  = layer.ModeExplore
)

// Reasons recorded on run.unblinded.
const (
	ReasonExplore  = layer.ReasonExplore
	ReasonRevealed = layer.ReasonRevealed
)

// File names under a run directory: the observer stream beside the artifact,
// the director stream under sealed/.
const (
	ObserverFile = layer.ObserverFile
	DirectorFile = layer.DirectorFile
)

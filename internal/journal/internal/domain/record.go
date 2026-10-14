// Package domain holds the run journal's rules: the record vocabulary of the
// observer and director streams, their deterministic encoding, the digest of a
// stream and the reader that recovers a journal cut at any byte.
package domain

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	StreamObserver = "observer"
	StreamDirector = "director"
)

const (
	ObserverFile = model.FileObserverJournal
	DirectorFile = model.FileDirectorJournal
)

const (
	ModeEvaluate = model.ModeEvaluate
	ModeExplore  = model.ModeExplore
)

const (
	ReasonExplore  = "explore"
	ReasonRevealed = "revealed"
)

const (
	FramingPreamble  = "preamble"
	FramingPostamble = "postamble"
)

const (
	TypeRunOpened        = "run.opened"
	TypeObservation      = "observation"
	TypeClockAdvanced    = "clock.advanced"
	TypeEffectorAck      = "effector.acknowledged"
	TypeConsumerReported = "consumer.reported"
	TypeCommand          = "command"
	TypeDelivery         = "delivery"
	TypeState            = "state"
	TypeEffectorOutcome  = "effector.outcome"
	TypeRunUnblinded     = "run.unblinded"
	TypeRunClosed        = "run.closed"
	TypePacing           = "pacing"
	TypeWire             = "wire"
	TypeKPITruth         = "kpi.truth"
	TypeFaultTruth       = "fault.truth"
)

const (
	keyStream = "stream"
	keySeq    = "seq"
	keyType   = "type"
	keyAt     = "at_ns"
)

func IsSession(recordType string) bool {
	return recordType == TypeConsumerReported || recordType == TypePacing || recordType == TypeWire
}

type Record struct {
	Stream string
	Seq    int64
	Type   string
	AtNS   int64
	Fields map[string]any
}

func encodeRecord(r Record) ([]byte, error) {
	line := make(map[string]any, len(r.Fields)+4)
	maps.Copy(line, r.Fields)
	line[keyStream], line[keySeq], line[keyType], line[keyAt] = r.Stream, r.Seq, r.Type, r.AtNS
	raw, err := json.Marshal(line)
	if err != nil {
		return nil, fmt.Errorf("journal: encode %s record: %w", r.Type, err)
	}
	return raw, nil
}

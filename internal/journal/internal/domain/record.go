// Package domain holds the run journal's rules: the record vocabulary of the
// observer and director streams, their deterministic encoding, the digest of a
// stream and the reader that recovers a journal cut at any byte.
package domain

import (
	"encoding/json"
	"fmt"
	"maps"
)

// The two streams of a journal. The observer stream carries what a consumer
// could know; the director stream carries what makes the run true.
const (
	StreamObserver = "observer"
	StreamDirector = "director"
)

// File names under a run directory: the observer stream beside the artifact,
// the director stream under sealed/.
const (
	ObserverFile = "journal.jsonl"
	DirectorFile = "sealed/journal.jsonl"
)

// Run modes stamped on run.opened. An explore run is unblinded from creation.
const (
	ModeEvaluate = "evaluate"
	ModeExplore  = "explore"
)

// Reasons a run was unblinded.
const (
	ReasonExplore  = "explore"
	ReasonRevealed = "revealed"
)

// Record types (docs/design/RUN_JOURNAL.md section 2).
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
	typePacing           = "pacing"
)

// The envelope keys every line carries.
const (
	keyStream = "stream"
	keySeq    = "seq"
	keyType   = "type"
	keyAt     = "at_ns"
)

// IsSession reports whether a record type is a session record: an out-of-band
// report or live control that is not in the command log, so replay cannot
// reproduce it. Session records are in the file but outside the stream digest.
func IsSession(recordType string) bool {
	return recordType == TypeConsumerReported || recordType == typePacing
}

// Record is one journal line: the envelope every line carries and the fields
// of its type.
type Record struct {
	Stream string
	Seq    int64
	Type   string
	AtNS   int64
	Fields map[string]any
}

// encodeRecord returns the record as one JSON object without a newline. Keys are in
// sorted order, so the bytes are a pure function of the record.
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

package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Kind int

const (
	DigestOnly Kind = iota
	ObserverJournal
	DirectorJournal
	LedgerRows
)

type Tally struct {
	Rows              int64
	Delivered         int64
	Dropped           int64
	DeliveredByReason map[string]int64
	DroppedByReason   map[string]int64
}

type Scan struct {
	Torn         bool
	RawDigest    string
	ReplayDigest string
	Records      []journal.Record
	Delivered    int64
	Ledger       Tally
	Err          error
}

func (d Document) scan(kind Kind) *Scan {
	if d.Scan != nil {
		return d.Scan
	}
	scanner := NewScanner(kind)
	for _, line := range committedLines(d.Raw) {
		scanner.Line(line)
	}
	return scanner.Finish(canonical.DigestBytes(d.Raw), len(d.Raw) > 0 && d.Raw[len(d.Raw)-1] != '\n')
}

type Scanner struct {
	kind   Kind
	scan   Scan
	lines  int
	replay hash.Hash
}

func NewScanner(kind Kind) *Scanner {
	s := &Scanner{kind: kind}
	if kind == ObserverJournal || kind == DirectorJournal {
		s.replay = sha256.New()
	}
	s.scan.Ledger = Tally{DeliveredByReason: map[string]int64{}, DroppedByReason: map[string]int64{}}
	return s
}

func (s *Scanner) Line(line []byte) {
	s.lines++
	switch s.kind {
	case ObserverJournal, DirectorJournal:
		s.journalLine(line)
	case LedgerRows:
		s.ledgerLine(line)
	case DigestOnly:
	}
}

func (s *Scanner) Finish(rawDigest string, torn bool) *Scan {
	s.scan.RawDigest, s.scan.Torn = rawDigest, torn
	if s.replay != nil {
		s.scan.ReplayDigest = "sha256:" + hex.EncodeToString(s.replay.Sum(nil))
	}
	return &s.scan
}

func (s *Scanner) journalLine(line []byte) {
	if journal.Replayable(line) {
		s.replay.Write(line)
		s.replay.Write([]byte{'\n'})
	}
	if s.scan.Err != nil {
		return
	}
	record, err := journal.DecodeLine(line, s.lines)
	if err != nil {
		s.scan.Err = err
		return
	}
	s.keep(record)
}

func (s *Scanner) keep(record journal.Record) {
	if _, framing := record.Fields["framing"]; record.Type == journal.TypeObservation && !framing {
		s.scan.Delivered++
	}
	if s.lines == 1 || retained(s.kind, record.Type) {
		s.scan.Records = append(s.scan.Records, record)
	}
}

func retained(kind Kind, recordType string) bool {
	if kind == ObserverJournal {
		return recordType != journal.TypeObservation
	}
	switch recordType {
	case journal.TypeRunOpened, journal.TypeCommand, journal.TypeEffectorOutcome, journal.TypeKPITruth, journal.TypeFaultTruth:
		return true
	}
	return false
}

func (s *Scanner) ledgerLine(line []byte) {
	if len(line) == 0 || s.scan.Err != nil {
		return
	}
	var row model.LedgerRecord
	if err := json.Unmarshal(line, &row); err != nil {
		s.scan.Err = fmt.Errorf("report: ledger.jsonl line %d: %w", s.lines, err)
		return
	}
	s.scan.Ledger.add(row)
}

func (t *Tally) add(row model.LedgerRecord) {
	t.Rows++
	byReason, total := t.DroppedByReason, &t.Dropped
	if row.Delivered {
		byReason, total = t.DeliveredByReason, &t.Delivered
	}
	byReason[row.DeliveryReason]++
	*total++
}

func committedLines(raw []byte) [][]byte {
	lines := bytes.Split(raw, []byte{'\n'})
	return lines[:len(lines)-1]
}

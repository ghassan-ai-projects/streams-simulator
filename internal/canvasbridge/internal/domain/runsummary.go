package domain

import (
	"bytes"
	"encoding/json"
)

type RunSummary struct {
	Plant      string
	Explore    bool
	Closed     bool
	Incomplete bool
	DurationNs int64
}

type runOpened struct {
	Type    string `json:"type"`
	Mode    string `json:"mode"`
	StartNs int64  `json:"start_time_ns"`
	Domain  struct {
		ID string `json:"id"`
	} `json:"domain"`
}

type runClosed struct {
	Type       string `json:"type"`
	AtNs       int64  `json:"at_ns"`
	Incomplete bool   `json:"incomplete"`
}

func SummarizeRun(firstLine, tail []byte) RunSummary {
	var opened runOpened
	if json.Unmarshal(firstLine, &opened) != nil || opened.Type != "run.opened" {
		return RunSummary{}
	}
	summary := RunSummary{Plant: opened.Domain.ID, Explore: opened.Mode == "explore"}
	var closed runClosed
	if json.Unmarshal(lastLine(tail), &closed) == nil && closed.Type == "run.closed" {
		summary.Closed, summary.Incomplete = true, closed.Incomplete
		summary.DurationNs = closed.AtNs - opened.StartNs
	}
	return summary
}

// RunArtifact is the part of a run folder's run.json the listing reads.
type RunArtifact struct {
	Seed       *uint64 `json:"seed"`
	Error      string  `json:"error"`
	CommandLog []struct {
		Op string `json:"op"`
	} `json:"command_log"`
}

// ParseRunArtifact reads run.json; an unreadable file yields no seed and no faults.
func ParseRunArtifact(runJSON []byte) RunArtifact {
	var artifact RunArtifact
	if json.Unmarshal(runJSON, &artifact) != nil {
		return RunArtifact{}
	}
	return artifact
}

func (a RunArtifact) Faults() int {
	count := 0
	for _, command := range a.CommandLog {
		if command.Op == "fault.inject" {
			count++
		}
	}
	return count
}

func (s RunSummary) State(running bool) RunState {
	return RunState{Ended: s.Closed && !running, Explore: s.Explore}
}

var sessionRecordTypes = map[string]bool{"consumer.reported": true, "wire": true, "pacing": true, "origin": true,
	"binding": true, "live": true, "session": true}

type sessionNote struct {
	Type   string          `json:"type"`
	Stream string          `json:"stream"`
	Mode   string          `json:"mode"`
	Plan   json.RawMessage `json:"plan"`
}

// RecordedSession reads the session record, which the binary writes as the first observer session record of a run.
func RecordedSession(head []byte) (json.RawMessage, bool) {
	for line := range bytes.SplitSeq(head, []byte("\n")) {
		var note sessionNote
		if json.Unmarshal(line, &note) != nil || !sessionRecordTypes[note.Type] {
			continue
		}
		return note.Plan, note.connected()
	}
	return nil, false
}

func (n sessionNote) connected() bool {
	return n.Type == "session" && n.Stream == "observer" && n.Mode == "connected" && bytes.HasPrefix(n.Plan, []byte("{"))
}

// Complete is true for a run that ended without being abandoned or aborted.
func (s RunSummary) Complete(running bool) bool {
	return s.State(running).Ended && !s.Incomplete
}

// NewRunEntry lists a run; the reason is shown only for a run that is not complete.
func NewRunEntry(worldID string, summary RunSummary, running bool, artifact RunArtifact) RunEntry {
	complete := summary.Complete(running)
	entry := RunEntry{WorldID: worldID, Ended: summary.State(running).Ended, Complete: complete, Running: running, Seed: artifact.Seed,
		Explore: summary.Explore, Plant: summary.Plant, DurationNs: summary.DurationNs, Faults: artifact.Faults()}
	if !complete {
		entry.IncompleteReason = artifact.Error
	}
	return entry
}

func lastLine(tail []byte) []byte {
	trimmed := bytes.TrimRight(tail, "\n")
	return bytes.TrimSpace(trimmed[bytes.LastIndexByte(trimmed, '\n')+1:])
}

package domain

import (
	"bytes"
	"encoding/json"
)

type RunSummary struct {
	Plant      string
	Explore    bool
	Closed     bool
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
	Type string `json:"type"`
	AtNs int64  `json:"at_ns"`
}

func SummarizeRun(firstLine, tail []byte) RunSummary {
	var opened runOpened
	if json.Unmarshal(firstLine, &opened) != nil || opened.Type != "run.opened" {
		return RunSummary{}
	}
	summary := RunSummary{Plant: opened.Domain.ID, Explore: opened.Mode == "explore"}
	var closed runClosed
	if json.Unmarshal(lastLine(tail), &closed) == nil && closed.Type == "run.closed" {
		summary.Closed = true
		summary.DurationNs = closed.AtNs - opened.StartNs
	}
	return summary
}

type runCommands struct {
	CommandLog []struct {
		Op string `json:"op"`
	} `json:"command_log"`
}

func FaultsInjected(runJSON []byte) int {
	var run runCommands
	if json.Unmarshal(runJSON, &run) != nil {
		return 0
	}
	count := 0
	for _, command := range run.CommandLog {
		if command.Op == "fault.inject" {
			count++
		}
	}
	return count
}

func (s RunSummary) State(running bool) RunState {
	return RunState{Ended: s.Closed && !running, Explore: s.Explore}
}

func lastLine(tail []byte) []byte {
	trimmed := bytes.TrimRight(tail, "\n")
	return bytes.TrimSpace(trimmed[bytes.LastIndexByte(trimmed, '\n')+1:])
}

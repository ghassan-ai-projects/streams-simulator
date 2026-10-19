package domain

import (
	"encoding/json"
	"errors"
)

const (
	EventInventory = "inventory"
	EventControl   = "control"
	EventCommand   = "command"
	EventClock     = "clock"
	EventFlush     = "flush"
	EventReply     = "reply"
	EventVerdict   = "verdict"
	EventResolve   = "resolve"
	EventStop      = "stop"
	EventEffect    = "effect"
	EventPurge     = "purge"
)

const (
	OpHold        = "hold"
	OpRelease     = "release"
	OpApproveMode = "approve_mode"
	OpApprove     = "approve"
	OpDeny        = "deny"
	OpEStop       = "estop"
	OpState       = "state"
)

const (
	ResultApplied = "applied"
	ResultRefused = "refused"

	VerdictPass   = "pass"
	VerdictRefuse = "refuse"
	VerdictQueue  = "queue"

	ResolutionForward = "forward"
	ResolutionRefuse  = "refuse"
)

const (
	ReasonMalformed      = "malformed"
	ReasonUnknownBoard   = "unknown_board"
	ReasonUnknownDevice  = "unknown_device"
	ReasonUnknownCommand = "unknown_command"
	ReasonTooManyScopes  = "too_many_scopes"
	ReasonNodeBusy       = "node_busy"
	reasonExpired        = "expired"
	reasonNotReady       = "not_ready"
	reasonDuplicate      = "duplicate"
	reasonSessionEnding  = "session_ending"
)

var ErrControlMalformed = errors.New("plantwire: the control request breaks its contract")

type InterruptEntry map[string]any

func newEntry(event string, pairs ...any) InterruptEntry {
	entry := InterruptEntry{"event": event}
	for i := 0; i+1 < len(pairs); i += 2 {
		key, _ := pairs[i].(string)
		entry[key] = pairs[i+1]
	}
	return entry
}

func (e InterruptEntry) Event() string { return text(e, "event") }

func (e InterruptEntry) Op() string { return text(e, "op") }

func (e InterruptEntry) Text(key string) string { return text(e, key) }

func (e InterruptEntry) Int(key string) int64 { return exactInt(e[key]) }

func exactInt(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	}
	return 0
}

func anys(items []string) []any {
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = item
	}
	return out
}

func stringsOf(value any) []string {
	listed, _ := value.([]any)
	out := make([]string, 0, len(listed))
	for _, item := range listed {
		name, _ := item.(string)
		out = append(out, name)
	}
	return out
}

func (e InterruptEntry) at(clock string, nodeNS int64) InterruptEntry {
	e["clock"], e["node_ns"] = clock, nodeNS
	return e
}

func EffectEntry(commandID, device, outcome, clock string, nodeNS int64) InterruptEntry {
	return newEntry(EventEffect, "command_id", commandID, "device", device, "outcome", outcome).at(clock, nodeNS)
}

func PurgeEntry(commandID, device, reason, clock string, nodeNS int64) InterruptEntry {
	return newEntry(EventPurge, "command_id", commandID, "device", device, "reason", reason).at(clock, nodeNS)
}

func ChangesInterrupts(entries []InterruptEntry) bool {
	for _, entry := range entries {
		switch {
		case entry.Event() == EventInventory, entry.Event() == EventControl, entry.Event() == EventResolve:
			return true
		case entry.Event() == EventVerdict && entry.Text("verdict") == VerdictQueue:
			return true
		}
	}
	return false
}

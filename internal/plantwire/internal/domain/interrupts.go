package domain

import (
	"cmp"
	"maps"
	"slices"
)

const (
	ScopeAll    = "all"
	ScopeBoard  = "board"
	ScopeDevice = "device"

	ClockWorld = "world"
	ClockWall  = "wall"

	ReasonLocalControl = "local_control"

	maxInterruptScopes  = 64
	maxPendingApprovals = 64
	maxWaitingRepeats   = 8
	maxInterruptDevices = 4096
	maxLoggedArgsBytes  = 2048
)

type Scope struct {
	Kind string
	ID   string
}

func (s Scope) Fields() map[string]any { return map[string]any{"kind": s.Kind, "id": s.ID} }

func scopeOf(fields any) Scope {
	object, _ := fields.(map[string]any)
	return Scope{Kind: text(object, "kind"), ID: text(object, "id")}
}

func compareScopes(a, b Scope) int {
	return cmp.Or(cmp.Compare(scopeRank(a.Kind), scopeRank(b.Kind)), cmp.Compare(a.ID, b.ID))
}

func scopeRank(kind string) int {
	return slices.Index([]string{ScopeAll, ScopeBoard, ScopeDevice}, kind)
}

func (s Scope) covers(d InterruptDevice) bool {
	switch s.Kind {
	case ScopeAll:
		return true
	case ScopeBoard:
		return d.Board != "" && d.Board == s.ID
	}
	return s.Kind == ScopeDevice && d.Device == s.ID
}

type InterruptDevice struct {
	Device, Board, Clock, SafeStop string
	Unreachable                    bool
}

func (d InterruptDevice) Fields() map[string]any {
	return map[string]any{"device": d.Device, "board": d.Board, "clock": d.Clock, "safe_stop": d.SafeStop, "unreachable": d.Unreachable}
}

func inventoryOf(entry InterruptEntry) []InterruptDevice {
	listed, _ := entry["devices"].([]any)
	devices := make([]InterruptDevice, 0, min(len(listed), maxInterruptDevices))
	for _, item := range listed[:min(len(listed), maxInterruptDevices)] {
		fields, _ := item.(map[string]any)
		devices = append(devices, InterruptDevice{Device: text(fields, "device"), Board: text(fields, "board"),
			Clock: text(fields, "clock"), SafeStop: text(fields, "safe_stop"), Unreachable: fields["unreachable"] == true})
	}
	slices.SortFunc(devices, func(a, b InterruptDevice) int { return cmp.Compare(a.Device, b.Device) })
	return devices
}

type pendingApproval struct {
	id, key, content, device, name, clock string
	deadlineNS, queuedNS                  int64
	waiting                               []string
	args                                  map[string]any
	truncated                             bool
}

type Interrupts struct {
	devices  []InterruptDevice
	holds    map[Scope]bool
	approves map[Scope]bool
	pending  []*pendingApproval
}

func NewInterrupts() *Interrupts {
	return &Interrupts{holds: map[Scope]bool{}, approves: map[Scope]bool{}}
}

func (m *Interrupts) deviceNamed(name string) InterruptDevice {
	index := slices.IndexFunc(m.devices, func(d InterruptDevice) bool { return d.Device == name })
	if index < 0 {
		return InterruptDevice{Device: name}
	}
	return m.devices[index]
}

func (m *Interrupts) known(scope Scope) bool {
	switch scope.Kind {
	case ScopeAll:
		return true
	case ScopeBoard:
		return slices.ContainsFunc(m.devices, func(d InterruptDevice) bool { return d.Board == scope.ID })
	}
	return slices.ContainsFunc(m.devices, func(d InterruptDevice) bool { return d.Device == scope.ID })
}

func sortedScopes(set map[Scope]bool) []Scope {
	return slices.SortedFunc(maps.Keys(set), compareScopes)
}

func (m *Interrupts) holdCovers(d InterruptDevice) bool {
	return slices.ContainsFunc(sortedScopes(m.holds), func(scope Scope) bool { return scope.covers(d) })
}

func (m *Interrupts) approveCovers(d InterruptDevice) bool {
	return slices.ContainsFunc(sortedScopes(m.approves), func(scope Scope) bool { return scope.covers(d) })
}

func (m *Interrupts) latchedCovers(d InterruptDevice) bool {
	return slices.ContainsFunc(sortedScopes(m.holds), func(scope Scope) bool { return m.holds[scope] && scope.covers(d) })
}

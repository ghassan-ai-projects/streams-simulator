package domain

import "slices"

func (m *Interrupts) control(in InterruptEntry) []InterruptEntry {
	switch in.Op() {
	case OpHold, OpRelease, OpApproveMode, OpEStop:
		return m.scoped(in)
	case OpApprove, OpDeny:
		return m.decide(in)
	case OpState:
		return []InterruptEntry{m.reply(in, ResultApplied, "")}
	}
	return []InterruptEntry{m.reply(in, ResultRefused, ReasonMalformed)}
}

func (m *Interrupts) scoped(in InterruptEntry) []InterruptEntry {
	scope := scopeOf(in["scope"])
	if reason := m.scopeFault(in, scope); reason != "" {
		return []InterruptEntry{m.reply(in, ResultRefused, reason)}
	}
	return m.apply(in, scope)
}

func (m *Interrupts) apply(in InterruptEntry, scope Scope) []InterruptEntry {
	switch in.Op() {
	case OpHold:
		return m.hold(in, scope, false)
	case OpEStop:
		return m.hold(in, scope, true)
	case OpRelease:
		return m.clear(in, scope)
	}
	return m.approveMode(in, scope)
}

func (m *Interrupts) scopeFault(in InterruptEntry, scope Scope) string {
	op := in.Op()
	switch {
	case !validScope(scope):
		return ReasonMalformed
	case op == OpRelease || (op == OpApproveMode && in["enabled"] != true):
		return ""
	case !m.known(scope):
		return unknownReason(scope)
	case m.addsScope(op, scope) && len(m.holds)+len(m.approves) >= maxInterruptScopes:
		return ReasonTooManyScopes
	}
	return ""
}

func validScope(scope Scope) bool {
	switch scope.Kind {
	case ScopeAll:
		return scope.ID == ""
	case ScopeBoard, ScopeDevice:
		return scope.ID != ""
	}
	return false
}

func unknownReason(scope Scope) string {
	if scope.Kind == ScopeBoard {
		return ReasonUnknownBoard
	}
	return ReasonUnknownDevice
}

func (m *Interrupts) addsScope(op string, scope Scope) bool {
	if op == OpApproveMode {
		_, present := m.approves[scope]
		return !present
	}
	_, present := m.holds[scope]
	return !present
}

func (m *Interrupts) hold(in InterruptEntry, scope Scope, latch bool) []InterruptEntry {
	m.holds[scope] = m.holds[scope] || latch
	out := m.refuseCovered(in, scope)
	if !latch {
		return append(out, m.reply(in, ResultApplied, ""))
	}
	stops, stopped, skipped := m.stops(in, scope)
	return append(append(stops, out...), m.reply(in, ResultApplied, "", "stopped", anys(stopped), "skipped", anys(skipped)))
}

func (m *Interrupts) stops(in InterruptEntry, scope Scope) (entries []InterruptEntry, stopped, skipped []string) {
	for _, d := range m.devices {
		if !scope.covers(d) {
			continue
		}
		if d.SafeStop == "" {
			skipped = append(skipped, d.Device)
			continue
		}
		stopped = append(stopped, d.Device)
		entries = append(entries, newEntry(EventStop, "device", d.Device, "command", d.SafeStop).at(in.Text("clock"), in.Int("node_ns")))
	}
	return entries, stopped, skipped
}

func (m *Interrupts) clear(in InterruptEntry, scope Scope) []InterruptEntry {
	delete(m.holds, scope)
	return []InterruptEntry{m.reply(in, ResultApplied, "")}
}

func (m *Interrupts) approveMode(in InterruptEntry, scope Scope) []InterruptEntry {
	if in["enabled"] == true {
		m.approves[scope] = true
	} else {
		delete(m.approves, scope)
	}
	return []InterruptEntry{m.reply(in, ResultApplied, "")}
}

func (m *Interrupts) decide(in InterruptEntry) []InterruptEntry {
	index := m.pendingIndex(in.Text("command_id"))
	if index < 0 {
		return []InterruptEntry{m.reply(in, ResultRefused, ReasonUnknownCommand)}
	}
	p := m.pending[index]
	m.pending = slices.Delete(m.pending, index, index+1)
	return []InterruptEntry{m.resolution(in, p), m.reply(in, ResultApplied, "")}
}

func (m *Interrupts) resolution(in InterruptEntry, p *pendingApproval) InterruptEntry {
	if in.Op() == OpApprove && !m.holdCovers(m.deviceNamed(p.device)) {
		return resolveOf(in, p, ResolutionForward, "")
	}
	return resolveOf(in, p, ResolutionRefuse, ReasonLocalControl)
}

func (m *Interrupts) reply(in InterruptEntry, result, reason string, extra ...any) InterruptEntry {
	pairs := append([]any{"op", in.Op(), "result", result, "state", m.View()}, extra...)
	entry := newEntry(EventReply, pairs...).at(in.Text("clock"), in.Int("node_ns"))
	if reason != "" {
		entry["reason"] = reason
	}
	if id := in.Text("command_id"); id != "" {
		entry["command_id"] = id
	}
	return entry
}

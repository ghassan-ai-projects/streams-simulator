package domain

import "slices"

func (m *Interrupts) Step(in InterruptEntry) []InterruptEntry {
	switch in.Event() {
	case EventInventory:
		m.devices = inventoryOf(in)
	case EventControl:
		return m.control(in)
	case EventCommand:
		return m.command(in)
	case EventClock:
		return m.expire(in)
	case EventFlush:
		return m.flush(in)
	}
	return nil
}

func (m *Interrupts) command(in InterruptEntry) []InterruptEntry {
	device := m.deviceNamed(in.Text("device"))
	switch {
	case isDeclaredSafeStop(device, in):
		return verdictOf(in, VerdictPass, "")
	case m.idWaiting(in.Text("command_id")):
		return verdictOf(in, VerdictRefuse, reasonDuplicate)
	case m.holdCovers(device):
		return verdictOf(in, VerdictRefuse, ReasonLocalControl)
	case m.approveCovers(device):
		return m.queue(in)
	}
	return verdictOf(in, VerdictPass, "")
}

func isDeclaredSafeStop(d InterruptDevice, in InterruptEntry) bool {
	args, _ := in["args"].(map[string]any)
	return d.SafeStop != "" && in.Text("name") == d.SafeStop && len(args) == 0 && in["args_truncated"] != true
}

func (m *Interrupts) idWaiting(id string) bool {
	return slices.ContainsFunc(m.pending, func(p *pendingApproval) bool { return p.id == id || slices.Contains(p.waiting, id) })
}

func verdictOf(in InterruptEntry, verdict, reason string) []InterruptEntry {
	entry := newEntry(EventVerdict, "command_id", in.Text("command_id"), "verdict", verdict).at(in.Text("clock"), in.Int("node_ns"))
	if reason != "" {
		entry["reason"] = reason
	}
	return []InterruptEntry{entry}
}

func (m *Interrupts) queue(in InterruptEntry) []InterruptEntry {
	if p := m.pendingKey(in.Text("key")); p != nil {
		return m.repeat(in, p)
	}
	switch {
	case in.Int("deadline_ns") < in.Int("node_ns"):
		return verdictOf(in, VerdictRefuse, reasonExpired)
	case len(m.pending) >= maxPendingApprovals:
		return verdictOf(in, VerdictRefuse, reasonNotReady)
	}
	m.pending = append(m.pending, pendingOf(in))
	return verdictOf(in, VerdictQueue, "")
}

func pendingOf(in InterruptEntry) *pendingApproval {
	args, _ := in["args"].(map[string]any)
	return &pendingApproval{id: in.Text("command_id"), key: in.Text("key"), content: in.Text("content"), device: in.Text("device"),
		name: in.Text("name"), clock: in.Text("clock"), deadlineNS: in.Int("deadline_ns"), queuedNS: in.Int("node_ns"),
		args: args, truncated: in["args_truncated"] == true}
}

func (m *Interrupts) repeat(in InterruptEntry, p *pendingApproval) []InterruptEntry {
	switch {
	case p.content != in.Text("content"):
		return verdictOf(in, VerdictRefuse, reasonDuplicate)
	case len(p.waiting) >= maxWaitingRepeats:
		return verdictOf(in, VerdictRefuse, reasonNotReady)
	}
	p.waiting = append(p.waiting, in.Text("command_id"))
	return verdictOf(in, VerdictQueue, "")
}

func (m *Interrupts) pendingKey(key string) *pendingApproval {
	index := slices.IndexFunc(m.pending, func(p *pendingApproval) bool { return key != "" && p.key == key })
	if index < 0 {
		return nil
	}
	return m.pending[index]
}

func (m *Interrupts) pendingIndex(id string) int {
	return slices.IndexFunc(m.pending, func(p *pendingApproval) bool { return p.id == id })
}

func (m *Interrupts) expire(in InterruptEntry) []InterruptEntry {
	clock, now := in.Text("clock"), in.Int("node_ns")
	return m.resolveWhere(in, func(p *pendingApproval) (bool, string) { return p.clock == clock && p.deadlineNS < now, reasonExpired })
}

func (m *Interrupts) flush(in InterruptEntry) []InterruptEntry {
	return m.resolveWhere(in, func(*pendingApproval) (bool, string) { return true, reasonSessionEnding })
}

func (m *Interrupts) refuseCovered(in InterruptEntry, scope Scope) []InterruptEntry {
	return m.resolveWhere(in, func(p *pendingApproval) (bool, string) {
		return scope.covers(m.deviceNamed(p.device)), ReasonLocalControl
	})
}

func (m *Interrupts) resolveWhere(in InterruptEntry, selected func(*pendingApproval) (bool, string)) []InterruptEntry {
	var out []InterruptEntry
	kept := m.pending[:0:0]
	for _, p := range m.pending {
		if refused, reason := selected(p); refused {
			out = append(out, resolveOf(in, p, ResolutionRefuse, reason))
			continue
		}
		kept = append(kept, p)
	}
	m.pending = kept
	return out
}

func resolveOf(in InterruptEntry, p *pendingApproval, resolution, reason string) InterruptEntry {
	entry := newEntry(EventResolve, "command_id", p.id, "resolution", resolution, "waiting", anys(p.waiting)).at(in.Text("clock"), in.Int("node_ns"))
	if reason != "" {
		entry["reason"] = reason
	}
	return entry
}

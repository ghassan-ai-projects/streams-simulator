package domain

func (m *Interrupts) View() map[string]any {
	return map[string]any{"holds": m.holdList(), "approve": m.approveList(), "pending": m.pendingList(), "devices": m.affected()}
}

func EmptyInterruptView() map[string]any { return NewInterrupts().View() }

func (m *Interrupts) holdList() []any {
	list := []any{}
	for _, scope := range sortedScopes(m.holds) {
		list = append(list, map[string]any{"kind": scope.Kind, "id": scope.ID, "latched": m.holds[scope]})
	}
	return list
}

func (m *Interrupts) approveList() []any {
	list := []any{}
	for _, scope := range sortedScopes(m.approves) {
		list = append(list, scope.Fields())
	}
	return list
}

func (m *Interrupts) pendingList() []any {
	list := []any{}
	for _, p := range m.pending {
		list = append(list, map[string]any{"command_id": p.id, "device": p.device, "command": p.name, "deadline_ns": p.deadlineNS,
			"queued_at_ns": p.queuedNS, "clock": p.clock})
	}
	return list
}

func (m *Interrupts) affected() []any {
	list := []any{}
	for _, d := range m.devices {
		mode := m.modeOf(d)
		if mode != "" {
			list = append(list, map[string]any{"device": d.Device, "board": d.Board, "mode": mode, "latched": m.latchedCovers(d)})
		}
	}
	return list
}

func (m *Interrupts) modeOf(d InterruptDevice) string {
	switch {
	case m.holdCovers(d):
		return "hold"
	case m.approveCovers(d):
		return "approve"
	}
	return ""
}

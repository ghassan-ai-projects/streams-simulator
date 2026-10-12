package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/world"

type WorldStatus struct {
	ClockNS        int64
	Emitted        int64
	ActiveFaults   int
	NextEventNS    int64
	PendingEffects int
}

func (r *Run) Status() WorldStatus {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return WorldStatus{
		ClockNS:        r.World.Clock(),
		Emitted:        r.World.EmittedCount(),
		ActiveFaults:   r.World.ActiveFaultsCount(),
		NextEventNS:    r.World.NextEventNS(),
		PendingEffects: r.World.PendingKicks(),
	}
}

func (r *Run) Faults() []world.FaultInfo {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.World.ListFaults()
}

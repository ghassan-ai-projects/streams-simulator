package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/world"

// WorldStatus is a consistent read of the world's clock and queues, taken
// between commands.
type WorldStatus struct {
	ClockNS        int64
	Emitted        int64
	ActiveFaults   int
	NextEventNS    int64
	PendingEffects int
}

// Status reads the world's clock and queues. The world has no lock of its own;
// the run's command lock is what keeps a read from overlapping an advance.
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

// Faults lists the active faults, between commands.
func (r *Run) Faults() []world.FaultInfo {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.World.ListFaults()
}

package app

import (
	"context"
	"sync"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type Pacer interface {
	WaitUntil(ctx context.Context, simElapsedNS int64) error
	Reanchor(simElapsedNS int64, rate float64)
}

type job struct {
	worldID  string
	run      *run.Run
	scenario model.Scenario
	pacer    Pacer
	ctx      context.Context
	cancel   context.CancelFunc
	waitCtx  context.Context
	stopWait context.CancelFunc
	done     chan struct{}

	mu            sync.Mutex
	status        rules.JobStatus
	advances      int
	stopRequested bool
	gate          gate
	immediate     immediateGate
	faults        *faultLedger
}

func newJob(parent context.Context, r *run.Run, status rules.JobStatus, scenario model.Scenario, pacer Pacer, faults []rules.JobFault) *job {
	ctx, cancel := context.WithCancel(parent)
	waitCtx, stopWait := context.WithCancel(ctx)
	return &job{worldID: status.WorldID, run: r, scenario: scenario, pacer: pacer, ctx: ctx, cancel: cancel,
		waitCtx: waitCtx, stopWait: stopWait, done: make(chan struct{}), status: status, gate: newGate(waitCtx), faults: newFaultLedger(faults)}
}

func (j *job) snapshot() rules.JobStatus {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.copyStatus()
}

func (j *job) copyStatus() rules.JobStatus {
	return rules.CloneJobStatus(j.status)
}

func (j *job) requestStop() rules.JobStatus {
	j.mu.Lock()
	if j.status.State == rules.JobRunning || j.status.State == rules.JobPaused {
		j.status.State, j.stopRequested = rules.JobStopping, true
		j.gate.release()
		j.stopWait()
	}
	j.mu.Unlock()
	return j.snapshot()
}

func (j *job) finish(state rules.JobState, failure string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status.State, j.status.Failure = state, failure
}

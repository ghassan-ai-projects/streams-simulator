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
}

func newJob(parent context.Context, r *run.Run, status rules.JobStatus, scenario model.Scenario, pacer Pacer) *job {
	ctx, cancel := context.WithCancel(parent)
	waitCtx, stopWait := context.WithCancel(ctx)
	return &job{worldID: status.WorldID, run: r, scenario: scenario, pacer: pacer, ctx: ctx, cancel: cancel,
		waitCtx: waitCtx, stopWait: stopWait, done: make(chan struct{}), status: status}
}

func (j *job) snapshot() rules.JobStatus {
	j.mu.Lock()
	defer j.mu.Unlock()
	status := j.status
	status.Calls = append([]rules.JobCallOutcome{}, j.status.Calls...)
	return status
}

func (j *job) requestStop() rules.JobStatus {
	j.mu.Lock()
	if j.status.State == rules.JobRunning {
		j.status.State, j.stopRequested = rules.JobStopping, true
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

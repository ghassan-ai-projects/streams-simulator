package app

import (
	"context"
	"errors"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

var errPaused = errors.New("job paused")

type gate struct {
	paused     bool
	stepping   bool
	wake       chan struct{}
	parked     chan struct{}
	paceCtx    context.Context
	paceCancel context.CancelCauseFunc
	parent     context.Context
}

func newGate(parent context.Context) gate {
	g := gate{wake: make(chan struct{}), parent: parent}
	g.renewPace()
	return g
}

func (g *gate) renewPace() {
	g.paceCtx, g.paceCancel = context.WithCancelCause(g.parent)
}

func (g *gate) release() {
	g.paused, g.stepping = false, false
	g.signal()
}

func (g *gate) signal() {
	close(g.wake)
	g.wake = make(chan struct{})
}

func (g *gate) parkSignal() chan struct{} {
	if g.parked == nil {
		g.parked = make(chan struct{})
	}
	return g.parked
}

func (j *job) pause() (rules.JobStatus, error) {
	parked, err := j.beginPause()
	if err != nil {
		return rules.JobStatus{}, err
	}
	j.awaitParked(parked)
	return j.snapshot(), nil
}

func (j *job) beginPause() (<-chan struct{}, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	switch j.status.State {
	case rules.JobPaused:
		return j.stepInFlight(), nil
	case rules.JobRunning:
		j.status.State, j.gate.paused = rules.JobPaused, true
		j.gate.paceCancel(errPaused)
		j.gate.renewPace()
		return j.gate.parkSignal(), nil
	}
	return nil, notRunning(j.status)
}

func (j *job) stepInFlight() <-chan struct{} {
	if j.gate.stepping {
		return j.gate.parkSignal()
	}
	return nil
}

func (j *job) resume() (rules.JobStatus, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if !j.takesCommands() {
		return rules.JobStatus{}, notRunning(j.status)
	}
	if j.status.State == rules.JobPaused {
		j.status.State = rules.JobRunning
		j.gate.release()
		j.reanchor()
	}
	return j.copyStatus(), nil
}

func (j *job) step() (rules.JobStatus, error) {
	parked, err := j.beginStep()
	if err != nil {
		return rules.JobStatus{}, err
	}
	j.awaitParked(parked)
	return j.snapshot(), nil
}

func (j *job) beginStep() (<-chan struct{}, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.status.State != rules.JobPaused {
		return nil, notPaused(j.status)
	}
	if j.gate.stepping {
		return nil, rules.ToolErrorf(rules.CodeJobNotPaused, "job %s is already taking a step", j.worldID)
	}
	parked := j.gate.parkSignal()
	j.gate.stepping = true
	j.gate.signal()
	return parked, nil
}

func (j *job) setRate(rate float64) (rules.JobStatus, error) {
	if err := rules.CheckJobRate(rate); err != nil {
		return rules.JobStatus{}, err
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if !j.takesCommands() {
		return rules.JobStatus{}, notRunning(j.status)
	}
	j.status.Rate = rate
	j.reanchor()
	return j.copyStatus(), nil
}

func (j *job) takesCommands() bool {
	return j.status.State == rules.JobRunning || j.status.State == rules.JobPaused
}

func (j *job) reanchor() {
	j.pacer.Reanchor(j.status.ReachedNS-j.status.StartNS, j.status.Rate)
}

func (j *job) awaitParked(parked <-chan struct{}) {
	if parked == nil {
		return
	}
	select {
	case <-parked:
	case <-j.done:
	case <-j.ctx.Done():
	}
}

func notRunning(status rules.JobStatus) error {
	return rules.ToolErrorf(rules.CodeJobNotRunning, "job %s is %s; only a running or paused job takes controls", status.WorldID, status.State)
}

func notPaused(status rules.JobStatus) error {
	if status.State == rules.JobRunning {
		return rules.ToolErrorf(rules.CodeJobNotPaused, "job %s is running; pause it before stepping", status.WorldID)
	}
	return notRunning(status)
}

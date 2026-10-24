package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const jobListTool = "sim.job.list"

func (b *Bridge) settleJobs(structured json.RawMessage) {
	for _, job := range domain.JobsOf(structured) {
		if job.Active() {
			b.running[job.WorldID] = true
			continue
		}
		delete(b.running, job.WorldID)
	}
}

func (b *Bridge) reconcileJobs(ctx context.Context) {
	if !b.runLive() || !b.directorMu.TryLock() {
		return
	}
	defer b.directorMu.Unlock()
	if b.director == nil {
		b.forgetRunning()
		return
	}
	b.askJobs(ctx)
}

func (b *Bridge) askJobs(ctx context.Context) {
	result, err := b.director.Call(ctx, jobListTool, nil)
	switch {
	case errors.Is(err, process.ErrGone):
		b.dropDirector()
	case err == nil && result.Failure == "":
		b.stateMu.Lock()
		defer b.stateMu.Unlock()
		b.settleJobs(result.Structured)
	}
}

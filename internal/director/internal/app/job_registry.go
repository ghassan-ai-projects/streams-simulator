package app

import (
	"sync"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/runner"
)

type jobRegistry struct {
	startMu  sync.Mutex
	mu       sync.Mutex
	byWorld  map[string]*job
	order    []string
	newPacer func(rate float64) Pacer
}

func newJobRegistry() *jobRegistry {
	return &jobRegistry{byWorld: map[string]*job{}, newPacer: func(rate float64) Pacer { return runner.NewPacer(rate) }}
}

func (r *jobRegistry) find(worldID string) *job {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.byWorld[worldID]
}

func (r *jobRegistry) add(j *job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byWorld[j.worldID] = j
	r.order = append(r.order, j.worldID)
}

func (r *jobRegistry) all() []*job {
	r.mu.Lock()
	defer r.mu.Unlock()
	jobs := make([]*job, 0, len(r.order))
	for _, worldID := range r.order {
		jobs = append(jobs, r.byWorld[worldID])
	}
	return jobs
}

func (r *jobRegistry) refuseWhenActive() error {
	for _, j := range r.all() {
		if status := j.snapshot(); rules.JobActive(status.State) {
			return rules.ToolErrorf(rules.CodeJobActive, "job %s is still running; stop it or wait for it to finish", status.WorldID)
		}
	}
	return nil
}

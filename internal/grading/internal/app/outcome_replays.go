package app

import (
	"context"
	"runtime"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const maxOutcomeWorkers = 4

type outcomeReplay func(context.Context, model.OutcomeMode) (model.OutcomeRun, error)

type outcomeReplays struct {
	ctx     context.Context
	replay  outcomeReplay
	modes   []model.OutcomeMode
	mu      sync.Mutex
	next    int
	failed  int
	cancels []context.CancelFunc
	runs    []model.OutcomeRun
	errs    []error
}

func outcomeWorkers(requested int) int {
	if requested > 0 {
		return requested
	}
	return min(maxOutcomeWorkers, runtime.GOMAXPROCS(0))
}

func replayOutcomes(ctx context.Context, modes []model.OutcomeMode, workers int, replay outcomeReplay) ([]model.OutcomeRun, error) {
	replays := &outcomeReplays{ctx: ctx, replay: replay, modes: modes, failed: len(modes),
		cancels: make([]context.CancelFunc, len(modes)), runs: make([]model.OutcomeRun, len(modes)), errs: make([]error, len(modes))}
	var group sync.WaitGroup
	for range min(outcomeWorkers(workers), len(modes)) {
		group.Go(replays.work)
	}
	group.Wait()
	return replays.result()
}

func (r *outcomeReplays) work() {
	for i, ctx, ok := r.claim(); ok; i, ctx, ok = r.claim() {
		run, err := r.replay(ctx, r.modes[i])
		r.finish(i, run, err)
	}
}

func (r *outcomeReplays) claim() (int, context.Context, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.next >= r.failed {
		return 0, nil, false
	}
	i := r.next
	r.next++
	ctx, cancel := context.WithCancel(r.ctx)
	r.cancels[i] = cancel
	return i, ctx, true
}

func (r *outcomeReplays) finish(i int, run model.OutcomeRun, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancels[i]()
	r.runs[i], r.errs[i] = run, err
	if err == nil || i > r.failed {
		return
	}
	r.failed = i
	for _, cancel := range r.cancels[i+1:] {
		if cancel != nil {
			cancel()
		}
	}
}

func (r *outcomeReplays) result() ([]model.OutcomeRun, error) {
	if r.failed < len(r.modes) {
		return nil, r.errs[r.failed]
	}
	return r.runs, nil
}

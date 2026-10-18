package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/world"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

func (r *Run) ReportQuiesced(throughNS int64) {
	throughNS = min(throughNS, r.worldClock())
	r.quiesceMu.Lock()
	defer r.quiesceMu.Unlock()
	if throughNS > r.quiescedThroughNS {
		r.quiescedThroughNS = throughNS
		r.journal.ConsumerReported(throughNS, false)
		close(r.quiesceNotify)
		r.quiesceNotify = make(chan struct{})
	}
}

func (r *Run) worldClock() int64 {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.World.Clock()
}

func (r *Run) SetQuiesceParkedHook(h func()) {
	r.quiesceParked = h
}

func (r *Run) awaitQuiescence(ctx context.Context, toNS int64) error {
	timer := r.Config.quiescenceClock.NewTimer(DefaultQuiescenceTimeout)
	defer timer.Stop()
	for {
		ch, through, done := r.quiescenceStatus(toNS)
		if done {
			return nil
		}
		r.notifyQuiescenceParked()
		if err := waitForQuiescence(ctx, timer, ch, through, toNS); err != nil {
			return err
		}
	}
}

func (r *Run) Domain() *domain.Compiled { return r.Config.Domain }

func (r *Run) Digest() string { return worldDigest(r) }

func (r *Run) SetFailureMode(mode world.FailureMode) {
	r.World.SetFailureMode(mode)
}

func (r *Run) UnblindedStamp() bool {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.unblinded
}

func (r *Run) AppliedPerturbations() []string {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.appliedPerturbations()
}

func (r *Run) appliedPerturbations() []string {
	out := make([]string, len(r.perturbHistory))
	copy(out, r.perturbHistory)
	return out
}

func (r *Run) quiescedThrough() int64 {
	r.quiesceMu.Lock()
	defer r.quiesceMu.Unlock()
	return r.quiescedThroughNS
}

func (r *Run) quiescenceStatus(toNS int64) (chan struct{}, int64, bool) {
	r.quiesceMu.Lock()
	defer r.quiesceMu.Unlock()
	if r.quiescedThroughNS >= toNS {
		return nil, r.quiescedThroughNS, true
	}
	return r.quiesceNotify, r.quiescedThroughNS, false
}

func (r *Run) notifyQuiescenceParked() {
	if r.quiesceParked != nil {
		r.quiesceParked()
	}
}

func waitForQuiescence(ctx context.Context, timer QuiescenceTimer, ch chan struct{}, through, toNS int64) error {
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("run: quiescence wait canceled: %w", ctx.Err())
	case <-timer.C():
		return fmt.Errorf("run: %w: quiesced through %d, asked for %d", ErrConsumerNotQuiesced, through, toNS)
	}
}

func (r *Run) JournalWire(direction string, record map[string]any) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.journal.Wire(r.journalNow(), direction, record)
}

func (r *Run) JournalOrigin(fields map[string]any) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.journal.Origin(r.journalNow(), fields)
}

func (r *Run) JournalBinding(fields map[string]any) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.journal.Binding(r.journalNow(), fields)
}

func (r *Run) JournalLive(fields map[string]any) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.journal.Live(r.journalNow(), fields)
}

func (r *Run) JournalSession(fields map[string]any) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.journal.Session(r.journalNow(), fields)
}

func (r *Run) FlushJournal() error {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	if err := r.journal.Flush(); err != nil {
		return fmt.Errorf("run: flush the journal: %w", err)
	}
	return nil
}

func (r *Run) JournalPacing(throughSeq uint64, waitedMS int64, missed bool) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.journal.Pacing(r.journalNow(), throughSeq, waitedMS, missed)
}

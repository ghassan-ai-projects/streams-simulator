package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/clock"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func (r *Run) openJournal() error {
	if r.Config.JournalDir == "" {
		return nil
	}
	opened, err := journal.NewInDir(r.Config.JournalDir, r.Config.retainJournal)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	r.journal = opened
	return nil
}

func (r *Run) journalOpened() {
	r.journal.Opened(r.openedRecord(r.mode()))
	if r.Config.Explore {
		r.unblinded, r.unblindedAt = true, clock.Stamp()
		r.journal.Unblinded(r.worldStartTimeNS, journal.ReasonExplore)
	}
}

func (r *Run) openedRecord(mode string) journal.Opened {
	cfg := r.Config
	return journal.Opened{RunID: r.ID, SimVersion: model.SimVersion,
		Domain:  model.ArtifactRef{ID: cfg.Domain.Spec.ID, Version: cfg.Domain.Spec.Version, Digest: cfg.Domain.Digest},
		Adapter: model.ArtifactRef{ID: cfg.Adapter.ID, Version: cfg.Adapter.Version, Digest: rules.AdapterDigest(cfg.Adapter)},
		Seed:    cfg.Seed, StartTimeNS: r.worldStartTimeNS, EntityIDs: r.World.InitialEntityIDs(), Mode: mode}
}

func (r *Run) mode() string {
	if r.Config.Explore {
		return journal.ModeExplore
	}
	return journal.ModeEvaluate
}

func (r *Run) journalNow() int64 { return r.World.Clock() }

func (r *Run) flushJournal() {
	if err := r.journal.Flush(); err != nil {
		r.fail(err)
	}
}

func (r *Run) journalEffectorCall(call journal.Call, result *world.InvokeResult, err error) {
	now := r.journalNow()
	if err != nil {
		r.journal.Outcome(now, call, nil, err.Error())
		return
	}
	r.journal.Outcome(now, call, result, "")
	r.journal.Acknowledged(now, call, world.Acknowledge(result))
}

func (r *Run) closeJournal() {
	r.journalFinalFaultTruth()
	if truth := r.World.KPITruth(); truth != nil {
		r.journal.KPITruth(r.journalNow(), truth)
	}
	if r.unblinded && !r.Config.Explore {
		r.journal.Unblinded(r.journalNow(), journal.ReasonRevealed)
	}
	digests, err := r.journal.Close(r.journalNow(), r.incomplete, r.failureText())
	r.journalDigests = digests
	if err != nil {
		r.fail(err)
	}
}

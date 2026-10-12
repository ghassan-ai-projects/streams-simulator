package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/clock"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// The run journal (docs/design/RUN_JOURNAL.md) is written as the run happens.
// Every record is a pure function of the run, so a replay writes the same
// bytes; a record that cannot be written fails the run at the next boundary.

// openJournal replaces the in-memory journal with the files under
// Config.JournalDir when the run has a directory.
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

// journalOpened records the run's identity and, for an explore run, stamps it
// unblinded from creation.
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

// mode is the run's mode as the artifact records it.
func (r *Run) mode() string {
	if r.Config.Explore {
		return journal.ModeExplore
	}
	return journal.ModeEvaluate
}

// journalNow is the world time journal records are stamped with.
func (r *Run) journalNow() int64 { return r.World.Clock() }

// flushJournal pushes both streams to their files; a failure fails the run.
func (r *Run) flushJournal() {
	if err := r.journal.Flush(); err != nil {
		r.fail(err)
	}
}

// journalEffectorCall records an effector call: the acknowledgement the
// operator received (observer) and the world's record of the call (director).
func (r *Run) journalEffectorCall(call journal.Call, result *world.InvokeResult, err error) {
	now := r.journalNow()
	if err != nil {
		r.journal.Outcome(now, call, nil, err.Error())
		return
	}
	r.journal.Outcome(now, call, result, "")
	r.journal.Acknowledged(now, call, world.Acknowledge(result))
}

// closeJournal ends both streams and records their digests for the artifact.
func (r *Run) closeJournal() {
	if r.unblinded && !r.Config.Explore {
		r.journal.Unblinded(r.journalNow(), journal.ReasonRevealed)
	}
	digests, err := r.journal.Close(r.journalNow(), r.incomplete, rules.ErrorString(r.runErr))
	r.journalDigests = digests
	if err != nil {
		r.fail(err)
	}
}

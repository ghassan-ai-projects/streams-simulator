package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type streams struct {
	observer []journal.Record
	director []journal.Record
}

func TruthAllowed(folder Folder) bool {
	if folder.Mismatch != "" {
		return false
	}
	if folder.Artifact != nil {
		return true
	}
	return openRun(folder.Observer.scan(ObserverJournal).Records).Mode == journal.ModeExplore
}

func Assemble(folder Folder, scoring *Scoring) (*Report, error) {
	if !TruthAllowed(folder) {
		folder.Director, folder.Ledger, folder.History = Document{}, Document{}, Document{}
	}
	in, err := parse(folder)
	if err != nil {
		return nil, err
	}
	report := newReport(folder, in, scoring)
	if folder.Mismatch != "" {
		return report, nil
	}
	return withTruth(report, folder.Artifact, in)
}

func newReport(folder Folder, in streams, scoring *Scoring) *Report {
	run := runOf(folder, in.observer)
	return &Report{
		SchemaVersion:    SchemaVersion,
		TruthUnavailable: truthUnavailable(folder, run),
		TornTail:         tornFiles(folder),
		Run:              run,
		Digests:          digestsOf(folder.Artifact),
		Verification:     verify(folder),
		Counts:           countsOf(folder),
		Consumer:         consumerOf(folder.Verdict),
		Scoring:          scoring,
	}
}

func truthUnavailable(folder Folder, run Run) string {
	switch {
	case folder.Mismatch != "":
		return folder.Mismatch
	case folder.Artifact == nil && run.Mode != journal.ModeExplore:
		return TruthSealed
	}
	return ""
}

func parse(folder Folder) (streams, error) {
	observer := folder.Observer.scan(ObserverJournal)
	if observer.Err != nil {
		return streams{}, fmt.Errorf("report: journal.jsonl: %w", observer.Err)
	}
	director := folder.Director.scan(DirectorJournal)
	if director.Err != nil {
		return streams{}, fmt.Errorf("report: sealed/journal.jsonl: %w", director.Err)
	}
	if err := folder.Ledger.scan(LedgerRows).Err; err != nil {
		return streams{}, err
	}
	return streams{observer: observer.Records, director: director.Records}, nil
}

func withTruth(report *Report, art *model.RunArtifact, in streams) (*Report, error) {
	if !report.Run.Ended && report.Run.Mode != journal.ModeExplore {
		return report, nil
	}
	truth, err := truthFrom(art, in.director)
	if err != nil {
		return nil, err
	}
	report.Truth = truth
	if art == nil {
		report.Run.Seed = directorSeed(in.director)
	}
	return report, nil
}

func truthFrom(art *model.RunArtifact, director []journal.Record) (*Truth, error) {
	commands, err := commandsOf(art, director)
	if err != nil {
		return nil, err
	}
	return truthOf(commands, director)
}

func consumerOf(verdict *model.Verdict) *Consumer {
	if verdict == nil {
		return nil
	}
	c := verdict.Consumer
	return &Consumer{Name: c.Name, Version: c.Version, Arm: c.Arm, ConfigDigest: c.ConfigDigest}
}

func digestsOf(art *model.RunArtifact) *Digests {
	if art == nil {
		return nil
	}
	return &Digests{World: art.WorldDigest, Trace: art.ExpectedTraceDigest,
		Journal: art.JournalDigest, SealedJournal: art.SealedJournalDigest, Ledger: art.LedgerDigest, History: art.HistoryDigest}
}

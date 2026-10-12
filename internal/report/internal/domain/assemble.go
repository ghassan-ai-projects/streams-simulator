package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type streams struct {
	observer []journal.Record
	director []journal.Record
	ledger   []model.LedgerRecord
}

func TruthAllowed(folder Folder) bool {
	if folder.Artifact != nil {
		return true
	}
	records, _ := journal.Read(folder.Observer.Raw)
	return openRun(records).Mode == journal.ModeExplore
}

func Assemble(folder Folder, scoring *Scoring) (*Report, error) {
	if !TruthAllowed(folder) {
		folder.Director, folder.Ledger = Document{}, Document{}
	}
	in, err := parse(folder)
	if err != nil {
		return nil, err
	}
	return withTruth(newReport(folder, in, scoring), folder.Artifact, in)
}

func newReport(folder Folder, in streams, scoring *Scoring) *Report {
	return &Report{
		SchemaVersion: SchemaVersion,
		Run:           runOf(folder.Artifact, in.observer),
		Digests:       digestsOf(folder.Artifact),
		Verification:  verify(folder),
		Counts:        countsOf(folder, in),
		Consumer:      consumerOf(folder.Verdict),
		Scoring:       scoring,
	}
}

func parse(folder Folder) (streams, error) {
	observer, err := readStream("journal.jsonl", folder.Observer)
	if err != nil {
		return streams{}, err
	}
	director, err := readStream("sealed/journal.jsonl", folder.Director)
	if err != nil {
		return streams{}, err
	}
	ledger, err := readLedger(folder.Ledger)
	return streams{observer: observer, director: director, ledger: ledger}, err
}

func readStream(name string, doc Document) ([]journal.Record, error) {
	records, err := journal.Read(doc.Raw)
	if err != nil {
		return nil, fmt.Errorf("report: %s: %w", name, err)
	}
	return records, nil
}

func withTruth(report *Report, art *model.RunArtifact, in streams) (*Report, error) {
	if !report.Run.Ended && report.Run.Mode != journal.ModeExplore {
		return report, nil
	}
	commands, err := commandsOf(art, in.director)
	if err != nil {
		return nil, err
	}
	report.Truth = truthOf(commands, in.director)
	if art == nil {
		report.Run.Seed = directorSeed(in.director)
	}
	return report, nil
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
		Journal: art.JournalDigest, SealedJournal: art.SealedJournalDigest}
}

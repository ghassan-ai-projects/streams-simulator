package domain

import (
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	foreignArtifact = "run.json belongs to another run: "
	foreignSealed   = "sealed/journal.jsonl belongs to another run: "
)

func Reconcile(folder Folder) Folder {
	if folder.Artifact == nil || !folder.Observer.Present {
		return folder
	}
	first, found := observerOpened(folder.Observer)
	if !found {
		return foreign(folder, foreignArtifact+"journal.jsonl holds no run.opened record to check it against")
	}
	if differing := artifactDifferences(folder.Artifact, first); len(differing) > 0 {
		return foreign(folder, foreignArtifact+strings.Join(differing, ", ")+" differ from journal.jsonl")
	}
	return folder
}

func ReconcileSealed(folder Folder) Folder {
	if folder.Mismatch != "" || !folder.Director.Present {
		return folder
	}
	records := folder.Director.scan(DirectorJournal).Records
	if len(records) == 0 || records[0].Type != journal.TypeRunOpened {
		return folder
	}
	if differing := sealedDifferences(folder, records); len(differing) > 0 {
		folder.Director, folder.Ledger = Document{}, Document{}
		folder.Mismatch = foreignSealed + strings.Join(differing, ", ") + " differ from the run of the journals"
	}
	return folder
}

func foreign(folder Folder, reason string) Folder {
	folder.Artifact, folder.Mismatch = nil, reason
	return folder
}

func observerOpened(doc Document) (opened, bool) {
	records := doc.scan(ObserverJournal).Records
	var first opened
	if len(records) == 0 || records[0].Type != journal.TypeRunOpened {
		return first, false
	}
	decodeFields(records[0].Fields, &first)
	return first, true
}

func artifactDifferences(art *model.RunArtifact, first opened) []string {
	return differing(map[string]bool{
		"run_id":      art.RunID == first.RunID,
		"sim_version": art.SimVersion == first.SimVersion,
		"domain":      art.Domain == first.Domain,
		"adapter":     art.Adapter == first.Adapter,
	})
}

func sealedDifferences(folder Folder, records []journal.Record) []string {
	runID, known := expectedRunID(folder)
	same := map[string]bool{}
	if known {
		same["run_id"] = records[0].Fields["run_id"] == runID
	}
	if folder.Artifact != nil {
		seed := directorSeed(records)
		same["seed"] = seed != nil && *seed == folder.Artifact.Seed
	}
	return differing(same)
}

func expectedRunID(folder Folder) (string, bool) {
	if folder.Artifact != nil {
		return folder.Artifact.RunID, true
	}
	first, found := observerOpened(folder.Observer)
	return first.RunID, found
}

func differing(same map[string]bool) []string {
	var names []string
	for _, name := range sortedKeys(same) {
		if !same[name] {
			names = append(names, name)
		}
	}
	return names
}

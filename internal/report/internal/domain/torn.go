package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func tornFiles(folder Folder) []string {
	var torn []string
	for _, entry := range []struct {
		name string
		doc  Document
		kind Kind
	}{{"journal.jsonl", folder.Observer, ObserverJournal}, {"sealed/journal.jsonl", folder.Director, DirectorJournal},
		{model.FileLedger, folder.Ledger, LedgerRows}} {
		if entry.doc.scan(entry.kind).Torn {
			torn = append(torn, entry.name)
		}
	}
	return torn
}

func tornLines(torn []string) []string {
	lines := make([]string, len(torn))
	for i, name := range torn {
		lines[i] = "- " + escapeText(name) + ": the last line was cut short and dropped."
	}
	return lines
}

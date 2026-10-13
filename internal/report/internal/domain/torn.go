package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func tornFiles(folder Folder) []string {
	var torn []string
	for _, entry := range []struct {
		name string
		doc  Document
	}{{"journal.jsonl", folder.Observer}, {"sealed/journal.jsonl", folder.Director}, {model.FileLedger, folder.Ledger}} {
		if hasTornTail(entry.doc) {
			torn = append(torn, entry.name)
		}
	}
	return torn
}

func hasTornTail(doc Document) bool {
	return len(doc.Raw) > 0 && doc.Raw[len(doc.Raw)-1] != '\n'
}

func tornLines(torn []string) []string {
	lines := make([]string, len(torn))
	for i, name := range torn {
		lines[i] = "- " + escapeText(name) + ": the last line was cut short and dropped."
	}
	return lines
}

package domain

import (
	"strings"
	"testing"
)

func TestFileAndInMemoryEntryPointsFormatTheSameFailureDifferently(t *testing.T) {
	t.Parallel()
	doc := []byte(`{"id": "x"}`)
	_, fromFile := DecodeFile(doc, "a.adapter.json")
	_, fromMemory := LoadBytes(doc, "memory")
	if fromFile == nil || fromMemory == nil {
		t.Fatal("an incomplete adapter must be refused by both entry points")
	}
	if !strings.Contains(fromFile.Error(), ":\n  ") || strings.Contains(fromMemory.Error(), ":\n  ") {
		t.Fatalf("file layout is multi-line, memory single-line:\nfile:   %q\nmemory: %q", fromFile, fromMemory)
	}
}

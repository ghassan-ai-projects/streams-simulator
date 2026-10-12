package domain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func FuzzWireParse(f *testing.F) {
	files, _ := filepath.Glob(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", "*.json"))
	for _, file := range files {
		if raw, err := os.ReadFile(file); err == nil {
			f.Add(raw)
		}
	}
	f.Fuzz(func(t *testing.T, line []byte) {
		in, ok := Decode(line)
		if ok && in.Invalid == nil && in.Type == "" {
			t.Fatalf("a record without a type passed: %s", line)
		}
	})
}

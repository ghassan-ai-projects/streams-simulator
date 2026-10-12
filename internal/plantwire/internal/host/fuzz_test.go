package host

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func FuzzTheHostSurvivesAnyLine(f *testing.F) {
	for _, folder := range []string{"valid", "invalid", "must-ignore"} {
		files, _ := filepath.Glob(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", folder, "*.json"))
		for _, file := range files {
			if raw, err := os.ReadFile(file); err == nil {
				f.Add(bytes.Join(bytes.Fields(raw), []byte(" ")))
			}
		}
	}
	f.Fuzz(func(t *testing.T, line []byte) {
		h := New(Options{Environment: "simulated", ProtocolMax: 1, Script: []Scripted{{Device: "P1", Command: "start", Verify: &Verify{Metric: "flow", WithinNS: 1}}}})
		for _, part := range bytes.Split(line, []byte("\n")) {
			if in, ok := domain.Decode(part); ok {
				h.Receive(in)
			}
		}
	})
}

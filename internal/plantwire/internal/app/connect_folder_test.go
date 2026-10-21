package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAConnectRefusesARunFolderThatHoldsARunAndLeavesItUntouched(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"run.json", "interrupts.jsonl", "sealed/journal.jsonl"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			out := t.TempDir()
			if err := os.MkdirAll(filepath.Join(out, "sealed"), 0o700); err != nil {
				t.Fatal(err)
			}
			testsupport.WriteFileIn(t, out, name, []byte("earlier session\n"))
			_, err := Connect(t.Context(), connectOptions(t, planText(testsupport.SocketPath(t), ""), out))
			if !errors.Is(err, domain.ErrRunFolderTaken) {
				t.Fatalf("Connect into a folder holding %s = %v, want ErrRunFolderTaken", name, err)
			}
			if raw, readErr := os.ReadFile(filepath.Join(out, name)); readErr != nil || string(raw) != "earlier session\n" {
				t.Fatalf("%s after the refusal = %q, %v, want it untouched", name, raw, readErr)
			}
			if _, statErr := os.Stat(filepath.Join(out, "journal.jsonl")); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("journal.jsonl after the refusal: %v, want nothing written", statErr)
			}
		})
	}
}

func TestAFolderWithoutARunIsAcceptedButASecondAttemptInItIsRefused(t *testing.T) {
	t.Parallel()
	out := t.TempDir()
	testsupport.WriteFileIn(t, out, "notes.txt", []byte("bench m5\n"))
	absentHost := testsupport.SocketPath(t)
	_, err := Connect(t.Context(), connectOptions(t, planText(absentHost, ""), out))
	if err == nil || errors.Is(err, domain.ErrRunFolderTaken) {
		t.Fatalf("first attempt = %v, want it to start and then fail on the absent host", err)
	}
	if _, statErr := os.Stat(filepath.Join(out, "run.json")); statErr != nil {
		t.Fatalf("run.json after the first attempt: %v, want the failed session recorded", statErr)
	}
	if _, err := Connect(t.Context(), connectOptions(t, planText(absentHost, ""), out)); !errors.Is(err, domain.ErrRunFolderTaken) {
		t.Fatalf("second attempt in the same folder = %v, want ErrRunFolderTaken", err)
	}
}

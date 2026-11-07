package files

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type failingReader struct{}

var errBrokenPipe = errors.New("broken pipe")

func (failingReader) Read([]byte) (int, error) { return 0, errBrokenPipe }

func TestPopulationOfFileAndStreamNameTheSameInstances(t *testing.T) {
	t.Parallel()
	path := testsupport.Domain("host-system-health")
	fromFile, err := PopulationOfFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fromStream, err := PopulationOfStream(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if fromFile.Domain != "host-system-health" || len(fromFile.Sets) != 1 || len(fromFile.Sets[0].IDs) != 12 {
		t.Fatalf("population = %+v", fromFile)
	}
	if fromStream.Sets[0].IDs[11] != "rack-a/server-12" {
		t.Fatalf("stream population = %+v", fromStream)
	}
}

func TestPopulationOfFileNamesAMissingOrInvalidFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "bad.domain.json", `{"id":"bad"}`)
	for name, path := range map[string]string{"missing": filepath.Join(dir, "absent.domain.json"), "invalid": filepath.Join(dir, "bad.domain.json")} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := PopulationOfFile(path)
			if err == nil || !strings.Contains(err.Error(), filepath.Base(path)) {
				t.Fatalf("err = %v, want it to name %s", err, path)
			}
		})
	}
}

func TestPopulationOfStreamRefusesABrokenOversizedOrInvalidInput(t *testing.T) {
	t.Parallel()
	_, err := PopulationOfStream(failingReader{})
	if !errors.Is(err, errBrokenPipe) {
		t.Fatalf("broken: err = %v", err)
	}
	_, err = PopulationOfStream(bytes.NewReader(bytes.Repeat([]byte(" "), maxStreamBytes+1)))
	if err == nil || !strings.Contains(err.Error(), "more than") {
		t.Fatalf("oversized: err = %v", err)
	}
	_, err = PopulationOfStream(strings.NewReader(`{"id":"bad"}`))
	if err == nil || !strings.Contains(err.Error(), streamSource) {
		t.Fatalf("invalid: err = %v", err)
	}
}

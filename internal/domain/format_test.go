package domain_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryShippedDomainIsAlreadyCanonicallyFormatted(t *testing.T) {
	t.Parallel()
	for id := range v01RuntimeDigests {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			onDisk, err := os.ReadFile(testsupport.Domain(id))
			if err != nil {
				t.Fatal(err)
			}
			copyPath := filepath.Join(t.TempDir(), id+".domain.json")
			if err := os.WriteFile(copyPath, onDisk, 0o600); err != nil { //nolint:gosec // G703: the path is under t.TempDir and the name is a shipped domain id
				t.Fatal(err)
			}
			result, err := domain.Format(copyPath)
			if err != nil || result.Changed || result.Digest != shipped(t, id).Digest {
				t.Fatalf("Format = %+v (%v), want an unchanged file with the loaded digest", result, err)
			}
		})
	}
}

func TestFormatKeepsTheDigestAndTheRuntimeWorldWhateverTheLayout(t *testing.T) {
	t.Parallel()
	loaded := shipped(t, "k8s-cluster")
	messy := filepath.Join(t.TempDir(), "k8s.domain.json")
	if err := os.WriteFile(messy, compactJSON(t, loaded.Raw), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := domain.Format(messy)
	if err != nil || !result.Changed || result.Digest != loaded.Digest {
		t.Fatalf("Format = %+v (%v), want a changed file with the same digest", result, err)
	}
	reloaded, err := domain.Load(messy)
	if err != nil || reloaded.Digest != loaded.Digest || string(reloaded.Raw) != string(loaded.Raw) {
		t.Fatalf("reloaded = %v (%v), want the shipped canonical bytes", reloaded, err)
	}
}

func TestFormatRefusesAnInvalidDomain(t *testing.T) {
	t.Parallel()
	bad := filepath.Join(t.TempDir(), "bad.domain.json")
	if err := os.WriteFile(bad, []byte(`{"id":"X"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var invalid *model.ValidationError
	if _, err := domain.Format(bad); !errors.As(err, &invalid) || invalid.Source != bad {
		t.Fatalf("Format of an invalid domain = %v, want a validation error naming the file", err)
	}
	if _, err := domain.Format(filepath.Join(t.TempDir(), "absent.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Format of a missing file = %v, want ErrNotExist", err)
	}
}

func compactJSON(t *testing.T, raw []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, raw); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

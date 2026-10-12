package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestLoadCapabilitiesReadsTheShippedCatalog(t *testing.T) {
	t.Parallel()
	caps, err := LoadCapabilities(testsupport.DeviceCapabilities())
	if err != nil || caps.Digest() == "" {
		t.Fatalf("capabilities = %v, %v", caps, err)
	}
}

func TestLoadCapabilitiesNamesWhatItCannotReadOrLoad(t *testing.T) {
	t.Parallel()
	_, err := LoadCapabilities(filepath.Join(t.TempDir(), "absent.json"))
	if !errors.Is(err, fs.ErrNotExist) || !strings.HasPrefix(err.Error(), "read capabilities: ") {
		t.Fatalf("a missing file: err = %v", err)
	}
	broken := filepath.Join(t.TempDir(), "caps.json")
	if err := os.WriteFile(broken, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCapabilities(broken); err == nil || !strings.HasPrefix(err.Error(), "load capabilities: ") {
		t.Fatalf("a malformed catalog: err = %v", err)
	}
}

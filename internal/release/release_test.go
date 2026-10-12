package release_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/release"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestPublishWritesAnUnsignedManifestOfTheShippedInputs(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "release-manifest.json")
	result, err := release.Publish(release.Options{Out: out, DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir(),
		Author: "A <a@x>", Reviewer: "R <r@x>", Version: "v-test", Commit: "c-test", Now: time.Now})
	if err != nil || result.Path != out || result.Signed || result.Sim != "v-test@c-test" {
		t.Fatalf("result = %+v, %v", result, err)
	}
	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		t.Fatalf("manifest file = %v, %v", info, err)
	}
}

func TestPublishRefusesAMissingInputDirectory(t *testing.T) {
	t.Parallel()
	_, err := release.Publish(release.Options{Out: filepath.Join(t.TempDir(), "m.json"), DomainsDir: filepath.Join(t.TempDir(), "absent"),
		AdaptersDir: testsupport.AdaptersDir(), Now: time.Now})
	if !errors.Is(err, fs.ErrNotExist) || !strings.HasPrefix(err.Error(), "manifest: domain dir: ") {
		t.Fatalf("a missing domains directory: err = %v", err)
	}
}

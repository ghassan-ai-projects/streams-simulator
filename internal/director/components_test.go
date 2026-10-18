package director_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestADirectorStartedFromDirectoriesAlwaysServesTheBuiltInLibrary(t *testing.T) {
	t.Parallel()
	d, err := director.NewFromDirs(t.Context(), testsupport.DomainsDir(), testsupport.AdaptersDir(), t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := d.ComponentList()
	if err != nil || listed["components"] == nil {
		t.Fatalf("component list = %v, %v", listed, err)
	}
	detail, err := d.ComponentDescribe("streamsim/chiller")
	if err != nil || detail == nil {
		t.Fatalf("describe = %v, %v", detail, err)
	}
}

func TestADirectorRefusesToStartWithACustomerLibraryDirectoryThatDoesNotExist(t *testing.T) {
	t.Parallel()
	missing := []string{filepath.Join(t.TempDir(), "absent")}
	_, err := director.NewFromDirs(t.Context(), testsupport.DomainsDir(), testsupport.AdaptersDir(), t.TempDir(), missing, nil)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("error = %v, want a not-exist error", err)
	}
}

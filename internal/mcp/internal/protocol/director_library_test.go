package protocol

import (
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheDirectorServesItsComponentLibraryWhenGivenOne(t *testing.T) {
	t.Parallel()
	options := ServeOptions{DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir(), OutDir: t.TempDir(),
		ComponentsDirs: []string{filepath.Join(testsupport.RepositoryRoot(), "components")}}
	d, err := newDirector(t.Context(), options)
	if err != nil {
		t.Fatal(err)
	}
	tools := directorTools{d}
	if list, err := tools.handleComponentList(t.Context(), nil); err != nil || list == nil {
		t.Fatalf("component list = %v, %v", list, err)
	}
	if _, err := tools.handleComponentDescribe(t.Context(), map[string]any{"ref": "streamsim/chiller"}); err != nil {
		t.Fatal(err)
	}
	options.ComponentsDirs = []string{filepath.Join(t.TempDir(), "missing")}
	if _, err := newDirector(t.Context(), options); err == nil {
		t.Fatal("a missing library directory must stop the director from starting")
	}
}

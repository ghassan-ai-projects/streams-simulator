package app

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestGradeUsesTheRecordedSpecsAfterTheDomainFileChanges(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs, replays and grades one world on one goroutine")
	dir := t.TempDir()
	_, gt := closedLoopRun(t, dir)
	label := writeLabel(t, dir, gt)
	before, err := Grade(t.Context(), gradeOptions(dir, label))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	edited := bytes.Replace(raw, []byte(`"initial": 6.5,`), []byte(`"initial": 6.4,`), 1)
	if bytes.Equal(raw, edited) {
		t.Fatal("the pond domain has no initial 6.5 to edit")
	}
	domains := t.TempDir()
	// #nosec G703 -- the path is a file in the test's own temporary directory
	if err := os.WriteFile(filepath.Join(domains, "aquaculture-pond.domain.json"), edited, 0o600); err != nil {
		t.Fatal(err)
	}
	options := gradeOptions(dir, label)
	options.DomainsDir = domains
	after, err := Grade(t.Context(), options)
	if err != nil || !reflect.DeepEqual(after, before) {
		t.Fatalf("after the domain edit: %v\nafter:  %+v\nbefore: %+v", err, after, before)
	}
}

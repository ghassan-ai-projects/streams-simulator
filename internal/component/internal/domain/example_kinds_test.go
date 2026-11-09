package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const beltComposition = `{"id":"belt","title":"Belt","equipment":[
	{"use":"x/feeder@1","names":["F-1"],"implements":"feeder","board_type":"controller"},
	{"use":"x/drive@1","names":["D-1"],"implements":"drive"},
	{"use":"x/drive@1","names":["D-2"],"implements":"drive"},
	{"use":"x/pump@1","names":["P-1"]}]}`

func exampleKindPool() []File {
	return []File{
		{Path: "pool/feeder.kind.json", Raw: []byte(`{"kind":"feeder"}`)},
		{Path: "pool/drive.kind.json", Raw: []byte(`{"kind":"drive"}`)},
		{Path: "pool/unused.kind.json", Raw: []byte(`{"kind":"unused"}`)},
		{Path: "pool/controller.board.json", Raw: []byte(`{"board_type":"controller"}`)},
	}
}

func kindFileNames(files []KindFile) []string {
	names := []string{}
	for _, file := range files {
		names = append(names, file.Name)
	}
	return names
}

func TestAnExampleCarriesTheKindFilesAndBoardsItsEntriesImplement(t *testing.T) {
	t.Parallel()
	examples, err := LoadExamples([]File{{Path: "belt.composition.json", Raw: []byte(beltComposition)}}, exampleKindPool())
	if err != nil {
		t.Fatal(err)
	}
	found, err := examples.Find("belt")
	want := []string{"controller.board.json", "drive.kind.json", "feeder.kind.json"}
	if err != nil || !slices.Equal(kindFileNames(found.KindFiles), want) {
		t.Fatalf("Find(belt) kind files = %v, %v, want %v", kindFileNames(found.KindFiles), err, want)
	}
	if string(found.KindFiles[0].Document) != `{"board_type":"controller"}` {
		t.Fatalf("board document = %s", found.KindFiles[0].Document)
	}
}

func TestAnExampleThatImplementsAKindFileItDoesNotCarryIsRefused(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"a kind file": `{"id":"e","title":"E","equipment":[{"implements":"absent"}]}`,
		"a board":     `{"id":"e","title":"E","equipment":[{"implements":"drive","board_type":"absent"}]}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			examples, err := LoadExamples([]File{{Path: "e.composition.json", Raw: []byte(raw)}}, exampleKindPool())
			if err != nil {
				t.Fatal(err)
			}
			if _, err := examples.Find("e"); !errors.Is(err, ErrExampleKindFileMissing) || !strings.Contains(err.Error(), `"absent"`) {
				t.Fatalf("Find(e) error = %v", err)
			}
		})
	}
}

func TestAnExampleKindFileNeedsJsonAndAKindOrBoardType(t *testing.T) {
	t.Parallel()
	cases := map[string]File{
		"not json":            {Path: "x.kind.json", Raw: []byte(`{`)},
		"no kind":             {Path: "x.kind.json", Raw: []byte(`{"board_type":"b"}`)},
		"no board type":       {Path: "x.board.json", Raw: []byte(`{"kind":"k"}`)},
		"neither kind suffix": {Path: "x.json", Raw: []byte(`{"kind":"k"}`)},
	}
	for name, file := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := LoadExamples(nil, []File{file}); err == nil || !strings.Contains(err.Error(), "x.") {
				t.Fatalf("LoadExamples error = %v", err)
			}
		})
	}
}

func TestAnExampleWhoseCompositionIsNotJsonHasNoKindFiles(t *testing.T) {
	t.Parallel()
	if _, err := (exampleKindFiles{}).of("e", []byte(`{`)); err == nil || !strings.Contains(err.Error(), "example e") {
		t.Fatalf("of error = %v", err)
	}
}

func TestInstallingKindFilesWritesTheMissingKeepsTheSameAndRefusesTheDifferent(t *testing.T) {
	t.Parallel()
	files := []KindFile{
		{Name: "a.kind.json", Document: []byte(`{"kind":"a","metrics":[]}`)},
		{Name: "b.kind.json", Document: []byte(`{"kind":"b"}`)},
	}
	installed, writes, err := PlanKindFileInstall("e", files, map[string][]byte{"a.kind.json": []byte("{\n  \"metrics\": [],\n  \"kind\": \"a\"\n}\n")})
	if err != nil || !slices.Equal(installed.Written, []string{"b.kind.json"}) || !slices.Equal(installed.Unchanged, []string{"a.kind.json"}) {
		t.Fatalf("install = %+v, %v", installed, err)
	}
	if len(writes) != 1 || writes[0].Name != "b.kind.json" {
		t.Fatalf("writes = %+v", writes)
	}
	nothing, none, err := PlanKindFileInstall("e", nil, nil)
	if err != nil || nothing.Written == nil || nothing.Unchanged == nil || len(none) != 0 {
		t.Fatalf("empty install = %+v, %v", nothing, err)
	}
}

func TestInstallingKindFilesOverADifferentFileWritesNothing(t *testing.T) {
	t.Parallel()
	files := []KindFile{
		{Name: "a.kind.json", Document: []byte(`{"kind":"a"}`)},
		{Name: "b.kind.json", Document: []byte(`{"kind":"b"}`)},
		{Name: "c.kind.json", Document: []byte(`{"kind":"c"}`)},
	}
	present := map[string][]byte{"a.kind.json": []byte(`{"kind":"other"}`), "c.kind.json": []byte(`not json`)}
	installed, writes, err := PlanKindFileInstall("belt", files, present)
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractKindFile || writes != nil || len(invalid.Issues) != 2 {
		t.Fatalf("install = %+v, %v", installed, err)
	}
	if issue := invalid.Issues[0]; issue.Keyword != kindFileDiffers || issue.Pointer != "/a.kind.json" || !strings.Contains(issue.Message, "example belt") {
		t.Fatalf("issue = %+v", issue)
	}
	if !strings.Contains(err.Error(), "a.kind.json, c.kind.json") {
		t.Fatalf("summary = %v", err)
	}
}

package domain

import (
	"errors"
	"io/fs"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func pathsOf(sources []KindSource) []string {
	paths := []string{}
	for _, source := range sources {
		paths = append(paths, source.Path)
	}
	return paths
}

func TestSelectKindSourcesKeepsOnlyTheFilesOfTheWantedKinds(t *testing.T) {
	t.Parallel()
	unreadable := KindSource{Path: "kinds/pump.kind.json", Err: fs.ErrPermission}
	cases := []struct {
		name    string
		sources []KindSource
		want    []string
	}{
		{"a file is known by the kind it declares", []KindSource{{Path: "kinds/renamed.kind.json", Raw: []byte(`{"kind": "pump"}`)}, {Path: "kinds/other.kind.json", Raw: []byte(`{"kind": "valve"}`)}}, []string{"kinds/renamed.kind.json"}},
		{"a file that is not JSON is known by its file name", []KindSource{{Path: "kinds/pump.kind.json", Raw: []byte("{")}, {Path: "kinds/valve.kind.json", Raw: []byte("{")}}, []string{"kinds/pump.kind.json"}},
		{"a file that declares no kind is known by its file name", []KindSource{{Path: "kinds/pump.kind.json", Raw: []byte(`{"kind": 7}`)}}, []string{"kinds/pump.kind.json"}},
		{"a file that could not be read is known by its file name", []KindSource{unreadable, {Path: "kinds/valve.kind.json", Err: fs.ErrPermission}}, []string{"kinds/pump.kind.json"}},
		{"the declared kind wins over the file name", []KindSource{{Path: "kinds/pump.kind.json", Raw: []byte(`{"kind": "valve"}`)}}, []string{}},
		{"nothing to select from", nil, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := pathsOf(SelectKindSources(tc.sources, map[string]bool{"pump": true}))
			if !slices.Equal(got, tc.want) {
				t.Fatalf("SelectKindSources = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseKindFilesReturnsTheReadFailureOfASourceItsEdgeCouldNotRead(t *testing.T) {
	t.Parallel()
	_, _, err := ParseKindFiles([]KindSource{{Path: "kinds/pump.kind.json", Err: fs.ErrPermission}})
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("ParseKindFiles of an unreadable source = %v, want the read failure", err)
	}
}

func TestImplementedKindsNamesEachKindFileTheDomainImplementsOnce(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{Kinds: []model.Kind{{Name: "a", Implements: "pump"}, {Name: "b"}, {Name: "c", Implements: "pump"}, {Name: "d", Implements: "valve"}}}
	got := ImplementedKinds(spec)
	if len(got) != 2 || !got["pump"] || !got["valve"] {
		t.Fatalf("ImplementedKinds = %v, want pump and valve", got)
	}
}

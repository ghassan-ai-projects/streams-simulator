package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestExamplePlantsAreListedByTitleAndFoundById(t *testing.T) {
	t.Parallel()
	examples, err := LoadExamples([]File{
		{Path: "b.composition.json", Raw: []byte(`{"id":"b","title":"Beta","description":"second"}`)},
		{Path: "a.composition.json", Raw: []byte(`{"id":"a","title":"Alpha"}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	list := examples.List()
	if len(list) != 2 || list[0] != (ExampleSummary{ID: "a", Title: "Alpha"}) || list[1].Description != "second" {
		t.Fatalf("List = %+v", list)
	}
	document, err := examples.Find("b")
	if err != nil || !strings.Contains(string(document), `"Beta"`) {
		t.Fatalf("Find(b) = %s, %v", document, err)
	}
	if _, err := examples.Find("c"); !errors.Is(err, ErrUnknownExample) || !strings.Contains(err.Error(), `"c"`) {
		t.Fatalf("Find(c) error = %v", err)
	}
}

func TestAnExamplePlantNeedsJsonAndAnId(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		raw  string
		want string
	}{
		"not json": {raw: `{`, want: "example x.composition.json: unexpected end of JSON input"},
		"no id":    {raw: `{"title":"T"}`, want: "example x.composition.json has no id"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadExamples([]File{{Path: "x.composition.json", Raw: []byte(c.raw)}})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("LoadExamples error = %v, want %q", err, c.want)
			}
		})
	}
}

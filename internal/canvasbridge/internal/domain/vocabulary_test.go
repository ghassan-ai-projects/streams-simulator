package domain

import "testing"

func TestADefinitionCountsTheInstancesItExpandsTo(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		content string
		want    int
	}{
		"named instances":       {`{"instances":[{"kind":"a","names":["A-1","A-2"]},{"kind":"b","names":["B-1"]}]}`, 3},
		"a templated set":       {`{"instances":[{"kind":"a","id_template":"A-{n}","count":{"default":4}},{"kind":"b","names":["B-1"]}]}`, 5},
		"composition equipment": {`{"equipment":[{"use":"x/a@1","names":["A-1","A-2","A-3"]},{"use":"x/b@1","id_template":"B-{n}","count":{"default":2}}]}`, 5},
		"no instances":          {`{"id":"p"}`, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := DefinitionOf("p.json", []byte(tc.content)).Instances; got != tc.want {
				t.Fatalf("Instances = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestADefinitionReadsIDAndTitleAndTheUnreadableOnesKeepOnlyTheirPath(t *testing.T) {
	t.Parallel()
	got := DefinitionOf("listing/p.json", []byte(`{"id":"p","title":"P"}`))
	if got != (Definition{Path: "listing/p.json", ID: "p", Title: "P"}) {
		t.Fatalf("DefinitionOf = %+v", got)
	}
	if got := DefinitionOf("x.json", []byte("{")); got != (Definition{Path: "x.json"}) {
		t.Fatalf("unreadable = %+v", got)
	}
}

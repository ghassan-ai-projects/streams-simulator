package domain

import "testing"

func TestADefinitionReadsIDTitleAndInstanceCount(t *testing.T) {
	t.Parallel()
	got := DefinitionOf("listing/p.json", []byte(`{"id":"p","title":"P","instances":[{},{},{}]}`))
	if got != (Definition{Path: "listing/p.json", ID: "p", Title: "P", Instances: 3}) {
		t.Fatalf("DefinitionOf = %+v", got)
	}
	if got := DefinitionOf("x.json", []byte("{")); got != (Definition{Path: "x.json"}) {
		t.Fatalf("unreadable = %+v", got)
	}
}

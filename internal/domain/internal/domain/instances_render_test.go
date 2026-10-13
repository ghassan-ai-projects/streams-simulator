package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestInstanceRendersOnlyIdsTheSetCanPopulate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		set     model.InstanceSet
		id      string
		renders bool
	}{
		{"a declared name", model.InstanceSet{Names: []string{"K-1"}}, "K-1", true},
		{"a name the set does not declare", model.InstanceSet{Names: []string{"K-1"}}, "K-2", false},
		{"the last default index", template("line.{n}", 2, 0), "line.2", true},
		{"an index beyond the default without a maximum", template("line.{n}", 2, 0), "line.3", false},
		{"an index up to the maximum", template("line.{n}", 2, 5), "line.5", true},
		{"an index beyond the maximum", template("line.{n}", 2, 5), "line.6", false},
		{"index zero", template("line.{n}", 2, 0), "line.0", false},
		{"a leading zero the template does not render", template("line.{n}", 2, 0), "line.02", false},
		{"a named placeholder takes any id characters", template("{site}.{n}", 2, 0), "north.1", true},
		{"another literal", template("line.{n}", 2, 0), "row.1", false},
		{"a template with no index renders its one instance", template("main", 1, 0), "main", true},
		{"a repeated index must agree", template("{n}-{n}", 3, 0), "1-2", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := InstanceRenders(tc.set, tc.id); got != tc.renders {
				t.Fatalf("InstanceRenders(%+v, %q) = %v, want %v", tc.set, tc.id, got, tc.renders)
			}
		})
	}
}

func template(text string, initial, limit int) model.InstanceSet {
	return model.InstanceSet{IDTemplate: text, Count: model.EntityCount{Default: initial, Max: limit}}
}

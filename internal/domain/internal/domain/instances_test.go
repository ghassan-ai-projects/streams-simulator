package domain

import (
	"fmt"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestExplicitEntityIDsMustFitTheInstanceTemplateAndCountBounds(t *testing.T) {
	t.Parallel()
	pond, err := Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	pumps, err := Load(testsupport.Domain("rotating-machinery"))
	if err != nil {
		t.Fatal(err)
	}
	tooMany := make([]string, 65)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf("site-a/pond-%d", i+1)
	}
	cases := []struct {
		name    string
		compile *Compiled
		ids     []string
		fits    bool
	}{
		{"none selects the default population", pond, nil, true},
		{"ids that fit the template", pond, []string{"site-a/pond-1", "site-a/pond-12"}, true},
		{"a template with a named placeholder", pumps, []string{"site-b/pump-3"}, true},
		{"an id outside the template", pond, []string{"site-a/pond-1", "tank-2"}, false},
		{"a non-numeric index", pond, []string{"site-a/pond-x"}, false},
		{"a trailing suffix", pond, []string{"site-a/pond-1/extra"}, false},
		{"more ids than the maximum", pond, tooMany, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.compile.CheckEntityIDs(tc.ids); (err == nil) != tc.fits {
				t.Fatalf("CheckEntityIDs(%v) = %v, want fits=%v", tc.ids, err, tc.fits)
			}
		})
	}
}

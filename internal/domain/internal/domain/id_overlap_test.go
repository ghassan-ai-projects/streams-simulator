package domain

import "testing"

func TestTwoIDTemplatesOverlapWhenOneIDFitsBoth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b    string
		overlap bool
	}{
		{"pump-{n}", "pump-{n}", true},
		{"pump-{n}", "pump-1{n}", true},
		{"pump-{n}", "pump-{n}x", false},
		{"pump-{n}", "tank-{n}", false},
		{"p-{n}", "p-{site}", true},
		{"p-{site}-{n}", "p-{n}", false},
		{"p-{site}-{n}", "p-a-{n}", true},
		{"{n}", "{site}", true},
		{"a{n}", "{site}b", false},
		{"cell-{n}", "cell-0", true},
	}
	for _, tc := range cases {
		if got := templatesOverlap(tc.a, tc.b); got != tc.overlap {
			t.Fatalf("%q and %q overlap = %v, want %v", tc.a, tc.b, got, tc.overlap)
		}
		if got := templatesOverlap(tc.b, tc.a); got != tc.overlap {
			t.Fatalf("%q and %q overlap = %v (reversed), want %v", tc.b, tc.a, got, tc.overlap)
		}
	}
}

func TestAnInstanceSetMayNotRenderAnotherSetsIDs(t *testing.T) {
	t.Parallel()
	runViolations(t, []violationCase{
		{"a second set rendering pump ids", func(d map[string]any) {
			put(d, append(d["instances"].([]any), map[string]any{"kind": "pump", "id_template": "pump-1{n}", "count": map[string]any{"default": 1}}), "instances")
		}, "/instances/5/id_template"},
	})
}

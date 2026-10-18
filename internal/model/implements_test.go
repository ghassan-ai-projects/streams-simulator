package model

import "testing"

func TestHasImplementsTellsADomainThatNamesAKindFileFromOneThatDoesNot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		kinds []Kind
		want  bool
	}{
		{"no kinds", nil, false},
		{"kinds that implement nothing", []Kind{{Name: "pump"}, {Name: "fan"}}, false},
		{"one kind that implements a kind file", []Kind{{Name: "pump"}, {Name: "fan", Implements: "fan"}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := (&DomainSpec{Kinds: tc.kinds}).HasImplements(); got != tc.want {
				t.Fatalf("HasImplements = %v, want %v", got, tc.want)
			}
		})
	}
}

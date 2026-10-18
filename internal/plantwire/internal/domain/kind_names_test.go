package domain

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var workspaceSegment = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

func TestAKindFileStemIsTheKindNameWithOnlyTheColonReplacedAndFitsAWorkspacePathSegment(t *testing.T) {
	t.Parallel()
	longest := "k" + strings.Repeat("a", 127)
	cases := []struct{ kind, stem string }{
		{"pump", "pump"},
		{"thermal_sensor", "thermal_sensor"},
		{"Fan-2.v1", "Fan-2.v1"},
		{"board:part:7", "board_part_7"},
		{longest, longest[:maxStemBytes]},
	}
	for _, tc := range cases {
		if got := KindFileStem(tc.kind); got != tc.stem || !workspaceSegment.MatchString(got+KindFileSuffix) {
			t.Errorf("KindFileStem(%q) = %q (file %q), want %q within a workspace path segment", tc.kind, got, got+KindFileSuffix, tc.stem)
		}
	}
}

func TestKindsThatDifferOnlyInCaseAreOneKindForTheDuplicateRule(t *testing.T) {
	t.Parallel()
	named := func(names ...string) []any {
		kinds := []any{}
		for _, name := range names {
			kinds = append(kinds, map[string]any{"kind": name})
		}
		return kinds
	}
	cases := []struct {
		name  string
		kinds []any
		want  []int
	}{
		{"distinct names", named("pump", "valve"), nil},
		{"the same name", named("pump", "valve", "pump"), []int{2}},
		{"another case", named("pump", "PUMP", "Pump"), []int{1, 2}},
	}
	for _, tc := range cases {
		if got := DuplicateKinds(tc.kinds); !reflect.DeepEqual(got, tc.want) || KindsDuplicated(tc.kinds) != (len(tc.want) > 0) {
			t.Errorf("%s: DuplicateKinds = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestABirthAnnouncingKindsThatDifferOnlyInCaseIsRefusedAtTheSecond(t *testing.T) {
	t.Parallel()
	kinds := birthKinds()
	shouting := map[string]any{"kind": "PUMP", "metrics": []any{}, "commands": []any{}}
	_, err := parse(birthTextOf(t, append(kinds, shouting)))
	invalid := violation(t, err, ErrBirthRefused)
	if len(invalid.Issues) != 1 || invalid.Issues[0].Keyword != "kind_duplicate" || invalid.Issues[0].Pointer != "/kinds/2/kind" {
		t.Fatalf("issues = %+v, want the one kind_duplicate at /kinds/2/kind", invalid.Issues)
	}
}

func TestEchoedNamesAreClippedToTheirLimitAndPointerTokensAreEscaped(t *testing.T) {
	t.Parallel()
	if got := pointerToken("a/b~c"); got != "a~1b~0c" {
		t.Errorf("pointerToken = %q, want the RFC 6901 escape a~1b~0c", got)
	}
	long := strings.Repeat("é", 200)
	for _, got := range []string{clip(long), pointerToken(long)} {
		if len(got) > maxEchoed || !strings.HasPrefix(long, got) || len(got) == 0 {
			t.Errorf("clipped to %d bytes %q, want a valid prefix of at most %d bytes", len(got), got, maxEchoed)
		}
	}
	if clip("short") != "short" {
		t.Error("a short name was changed")
	}
}

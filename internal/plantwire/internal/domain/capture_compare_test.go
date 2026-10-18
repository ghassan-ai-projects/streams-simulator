package domain

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func capturedBirth(t *testing.T) Capture {
	t.Helper()
	captured, err := parse(birthLine(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	return captured
}

type workspaceFile struct{ stem, raw string }

func comparedWith(t *testing.T, captured Capture, files ...workspaceFile) Capture {
	t.Helper()
	var sources []KindSource
	for _, file := range files {
		sources = append(sources, KindSource{Path: "kinds/" + file.stem + ".kind.json", Raw: []byte(file.raw)})
	}
	workspace := LoadWorkspaceKinds(SelectFolded(sources, KindNames(captured)))
	compared, err := CompareWithFiles(captured, workspace, "kinds")
	if err != nil {
		t.Fatal(err)
	}
	return compared
}

func draftText(t *testing.T, kind CapturedKind) string {
	t.Helper()
	raw, err := json.Marshal(kind.Draft)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestACapturedKindIsNewSameOrDifferentFromTheWorkspaceFileOfItsName(t *testing.T) {
	t.Parallel()
	captured := capturedBirth(t)
	valve := `{"kind":"valve","metrics":[{"name":"position","data_type":"number"}],"commands":[]}`
	compared := comparedWith(t, captured, workspaceFile{"pump", draftText(t, captured.Kinds[0])}, workspaceFile{"valve", valve})
	pump, other := compared.Kinds[0].Comparison, compared.Kinds[1].Comparison
	wantPump := &KindComparison{Status: CompareSame, Target: "kinds/pump.kind.json", File: "kinds/pump.kind.json", FileDigest: captured.Kinds[0].KindDigest}
	if !reflect.DeepEqual(pump, wantPump) {
		t.Fatalf("pump comparison = %+v, want %+v", pump, wantPump)
	}
	if other.Status != CompareDiffers || other.File != "kinds/valve.kind.json" || other.Target != other.File || other.DiffersAt != "/metrics/0" ||
		other.FileDigest == captured.Kinds[1].KindDigest || other.FileDigest == "" {
		t.Fatalf("valve comparison = %+v, want a difference at the metric only the workspace file has, aimed at that file", other)
	}
	if captured.Kinds[0].Comparison != nil || captured.Kinds[1].Comparison != nil {
		t.Fatal("CompareWithFiles changed the capture it was given")
	}
}

func TestAKindNoWorkspaceFileDefinesIsNewAndAimedAtTheFileItWouldBecome(t *testing.T) {
	t.Parallel()
	compared := comparedWith(t, capturedBirth(t))
	for i, want := range []*KindComparison{{Status: CompareNew, Target: "kinds/pump.kind.json"}, {Status: CompareNew, Target: "kinds/valve.kind.json"}} {
		if !reflect.DeepEqual(compared.Kinds[i].Comparison, want) {
			t.Errorf("kind %d comparison = %+v, want %+v", i, compared.Kinds[i].Comparison, want)
		}
	}
	colon := Capture{Kinds: []CapturedKind{{Kind: "board:part.1", KindDigest: "sha256:x", Draft: map[string]any{}}}}
	moved, err := CompareWithFiles(colon, nil, filepath.Join("ws", "kinds"))
	if want := filepath.Join("ws", "kinds", "board_part.1.kind.json"); err != nil || moved.Kinds[0].Comparison.Target != want {
		t.Fatalf("target = %+v, %v, want %s: the stem is the kind name with its colon made safe", moved.Kinds[0].Comparison, err, want)
	}
}

func TestTheFirstDifferingPointerIsTheOneTheImplementsCheckWouldName(t *testing.T) {
	t.Parallel()
	captured := capturedBirth(t)
	cases := []struct{ name, from, to, at string }{
		{"a unit", `"m3/h"`, `"l/s"`, "/metrics/0/unit"},
		{"a command's risk", `"low"`, `"high"`, "/commands/0/risk_class"},
		{"a flag", `"writable":false`, `"writable":true`, "/metrics/0/writable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			text := replaceOnce(t, draftText(t, captured.Kinds[0]), tc.from, tc.to)
			compared := comparedWith(t, captured, workspaceFile{"pump", text})
			if got := compared.Kinds[0].Comparison; got.Status != CompareDiffers || got.DiffersAt != tc.at {
				t.Fatalf("comparison = %+v, want a difference at %s", got, tc.at)
			}
		})
	}
}

func TestTheFirstDifferingPointerEscapesTheMemberNamesItPasses(t *testing.T) {
	t.Parallel()
	draft := `{"kind":"pump","metrics":[],"commands":[{"name":"start","args_schema":{"properties":{"a/b":{"maximum":5},"c~d":{}}},"risk_class":"low","safe_stop":false}]}`
	changed := strings.Replace(draft, `"maximum":5`, `"maximum":6`, 1)
	var entry map[string]any
	if err := json.Unmarshal([]byte(changed), &entry); err != nil {
		t.Fatal(err)
	}
	captured := Capture{Kinds: []CapturedKind{{Kind: "pump", KindDigest: "sha256:x", Draft: entry}}}
	compared := comparedWith(t, captured, workspaceFile{"pump", draft})
	if got := compared.Kinds[0].Comparison.DiffersAt; got != "/commands/0/args_schema/properties/a~1b/maximum" {
		t.Fatalf("differs_at = %q, want the slash of the member name escaped", got)
	}
}

func TestANameThatDiffersOnlyInCaseCollidesWithTheWorkspaceFileAndIsNotNew(t *testing.T) {
	t.Parallel()
	captured, err := parse(birthTextOf(t, []any{map[string]any{"kind": "Pump", "metrics": []any{}, "commands": []any{}}}))
	if err != nil {
		t.Fatal(err)
	}
	existing := `{"kind":"pump","metrics":[],"commands":[]}`
	compared := comparedWith(t, captured, workspaceFile{"pump", existing})
	got := compared.Kinds[0]
	if got.Comparison.Status != CompareDiffers || got.Comparison.CollidesWith != "kinds/pump.kind.json" || got.Comparison.Target != "kinds/pump.kind.json" ||
		got.Comparison.File != "kinds/pump.kind.json" || got.Comparison.DiffersAt != "/kind" {
		t.Fatalf("comparison = %+v, want a difference at /kind that collides with, and is aimed at, kinds/pump.kind.json", got.Comparison)
	}
	if len(got.Issues) != 1 || got.Issues[0].Keyword != "kind_name_collision" || got.Issues[0].Pointer != "/kind" || got.Issues[0].Member != "kind" {
		t.Fatalf("issues = %+v, want one kind_name_collision at /kind", got.Issues)
	}
}

func TestAnExactNameWinsOverACollidingOneButTheCollisionIsStillReported(t *testing.T) {
	t.Parallel()
	captured := capturedBirth(t)
	compared := comparedWith(t, captured, workspaceFile{"pump", draftText(t, captured.Kinds[0])}, workspaceFile{"PUMP", `{"kind":"PUMP","metrics":[],"commands":[]}`})
	pump := compared.Kinds[0]
	if pump.Comparison.Status != CompareSame || pump.Comparison.Target != "kinds/pump.kind.json" || len(pump.Issues) != 1 || pump.Issues[0].Keyword != "kind_name_collision" {
		t.Fatalf("pump = %+v, want the exact file compared and the collision with kinds/PUMP.kind.json reported", pump)
	}
}

func TestNamesThatOnlyLookLikeAKindNameAreRefusedByTheSchemaBeforeAnyComparison(t *testing.T) {
	t.Parallel()
	for name, kind := range map[string]string{"a trailing space": "pump ", "a leading space": " pump", "a cyrillic lookalike": "рump", "an accent": "pumṕ"} {
		_, err := parse(birthLine(t, func(b map[string]any) { firstKind(b)["kind"] = kind }))
		invalid := birthViolationOf(t, err)
		if invalid.Issues[0].Pointer != "/kinds/0/kind" || invalid.Issues[0].Keyword != "pattern" {
			t.Errorf("%s: issues = %+v, want a pattern refusal at /kinds/0/kind", name, invalid.Issues)
		}
	}
}

func TestAWorkspaceFileThatCannotBeReadDegradesToAnIssueOfItsKindAndSpoilsNoOtherKind(t *testing.T) {
	t.Parallel()
	captured := capturedBirth(t)
	twice := `{"kind":"valve","metrics":[],"commands":[]}`
	compared := comparedWith(t, captured, workspaceFile{"pump", "{"}, workspaceFile{"valve", twice}, workspaceFile{"valve-again", twice})
	pump, valve := compared.Kinds[0], compared.Kinds[1]
	if pump.Comparison.Status != CompareUnreadable || pump.Comparison.File != "kinds/pump.kind.json" || pump.Comparison.Target != pump.Comparison.File ||
		len(pump.Issues) != 1 || pump.Issues[0].Keyword != "kind_file_unreadable" || !strings.Contains(pump.Issues[0].Message, "kinds/pump.kind.json") {
		t.Fatalf("pump = %+v %+v, want its comparison unreadable with one kind_file_unreadable issue", pump.Comparison, pump.Issues)
	}
	if valve.Comparison.Status != CompareUnreadable || len(valve.Issues) != 1 || !strings.Contains(valve.Issues[0].Message, "defined by") {
		t.Fatalf("valve = %+v %+v, want two files defining one kind reported as unreadable", valve.Comparison, valve.Issues)
	}
}

func replaceOnce(t *testing.T, text, from, to string) string {
	t.Helper()
	if !strings.Contains(text, from) {
		t.Fatalf("%q holds no %q", text, from)
	}
	return strings.Replace(text, from, to, 1)
}

func TestKindNamesListsEachCapturedKindInTheOrderOfTheBirth(t *testing.T) {
	t.Parallel()
	if got, want := KindNames(capturedBirth(t)), []string{"pump", "valve"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("KindNames = %v, want %v", got, want)
	}
}

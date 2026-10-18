package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func capturedBirth(t *testing.T) Capture {
	t.Helper()
	captured, err := ParseCapture(birthLine(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	return captured
}

func workspaceKinds(t *testing.T, files map[string]string) (KindFiles, []KindSummary) {
	t.Helper()
	var sources []KindSource
	for _, name := range []string{"pump", "valve"} {
		if raw, present := files[name]; present {
			sources = append(sources, KindSource{Path: "kinds/" + name + ".kind.json", Raw: []byte(raw)})
		}
	}
	return parsedKinds(t, sources...)
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
	files, summaries := workspaceKinds(t, map[string]string{"pump": draftText(t, captured.Kinds[0]), "valve": valve})
	compared, err := CompareWithFiles(captured, files, summaries)
	if err != nil {
		t.Fatal(err)
	}
	pump, other := compared.Kinds[0].Comparison, compared.Kinds[1].Comparison
	wantPump := &KindComparison{Status: CompareSame, File: "kinds/pump.kind.json", FileDigest: captured.Kinds[0].KindDigest}
	if !reflect.DeepEqual(pump, wantPump) {
		t.Fatalf("pump comparison = %+v, want %+v", pump, wantPump)
	}
	if other.Status != CompareDiffers || other.File != "kinds/valve.kind.json" || other.DiffersAt != "/metrics/0" || other.FileDigest == captured.Kinds[1].KindDigest || other.FileDigest == "" {
		t.Fatalf("valve comparison = %+v, want a difference at the metric only the workspace file has", other)
	}
	if captured.Kinds[0].Comparison != nil || captured.Kinds[1].Comparison != nil {
		t.Fatal("CompareWithFiles changed the capture it was given")
	}
}

func TestAKindNoWorkspaceFileDefinesIsNewAndNamesNoFile(t *testing.T) {
	t.Parallel()
	captured := capturedBirth(t)
	files, summaries := workspaceKinds(t, nil)
	compared, err := CompareWithFiles(captured, files, summaries)
	if err != nil {
		t.Fatal(err)
	}
	for i, kind := range compared.Kinds {
		if want := (&KindComparison{Status: CompareNew}); !reflect.DeepEqual(kind.Comparison, want) {
			t.Errorf("kind %d comparison = %+v, want %+v", i, kind.Comparison, want)
		}
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
			files, summaries := workspaceKinds(t, map[string]string{"pump": text})
			compared, err := CompareWithFiles(captured, files, summaries)
			if err != nil || compared.Kinds[0].Comparison.Status != CompareDiffers || compared.Kinds[0].Comparison.DiffersAt != tc.at {
				t.Fatalf("comparison = %+v, %v, want a difference at %s", compared.Kinds[0].Comparison, err, tc.at)
			}
		})
	}
}

func replaceOnce(t *testing.T, text, from, to string) string {
	t.Helper()
	if !strings.Contains(text, from) {
		t.Fatalf("%q holds no %q", text, from)
	}
	return strings.Replace(text, from, to, 1)
}

func TestKindNamesListsEachCapturedKindOnce(t *testing.T) {
	t.Parallel()
	if got, want := KindNames(capturedBirth(t)), map[string]bool{"pump": true, "valve": true}; !reflect.DeepEqual(got, want) {
		t.Fatalf("KindNames = %v, want %v", got, want)
	}
}

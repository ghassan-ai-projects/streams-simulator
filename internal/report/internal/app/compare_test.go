package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func variantFolder(t *testing.T) string {
	t.Helper()
	dir := copyFolder(t, testsupport.ReportRun())
	editFile(t, filepath.Join(dir, "sealed", "journal.jsonl"), func(raw string) string {
		kept := []string{}
		for _, line := range strings.Split(raw, "\n") {
			if !strings.Contains(line, `"act-2"`) {
				kept = append(kept, line)
			}
		}
		return strings.Replace(strings.Join(kept, "\n"), `"cleared_ns":null,"cleared_by":null`, `"cleared_ns":1767225690000000000,"cleared_by":"command"`, 1)
	})
	editFile(t, filepath.Join(dir, "run.json"), dropSecondCall)
	return dir
}

func dropSecondCall(raw string) string {
	var artifact map[string]any
	if err := model.DecodeBytes([]byte(raw), &artifact); err != nil {
		panic(err)
	}
	kept := []any{}
	for _, command := range artifact["command_log"].([]any) {
		args, _ := command.(map[string]any)["args"].(map[string]any)
		if args["command_id"] != "act-2" {
			kept = append(kept, command)
		}
	}
	artifact["command_log"] = kept
	out, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(out)
}

func editFile(t *testing.T, path string, edit func(string) string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(edit(string(raw))), 0o600); err != nil {
		t.Fatal(err)
	}
}

func jsonText(t *testing.T, value any) string {
	t.Helper()
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func comparisonText(t *testing.T, a, b, format string) []byte {
	t.Helper()
	built, err := Compare(t.Context(), a, b)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	out, err := Render(built, format)
	if err != nil {
		t.Fatalf("render %s: %v", format, err)
	}
	return out
}

func TestTheComparisonOfThePinnedRunWithAVariantMatchesItsGoldenDocuments(t *testing.T) {
	t.Parallel()
	variant := variantFolder(t)
	for _, tc := range []struct{ format, golden string }{{JSONFormat, "compare.golden.json"}, {Markdown, "compare.golden.md"}} {
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()
			got := comparisonText(t, testsupport.ReportRun(), variant, tc.format)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				if err := os.WriteFile(testsupport.ReportGolden(tc.golden), got, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(testsupport.ReportGolden(tc.golden))
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("%s differs from %s (%v):\n%s", tc.format, tc.golden, err, got)
			}
		})
	}
}

func TestTheComparisonValidatesAgainstItsContractAndIsTheSameBytesEverywhere(t *testing.T) {
	t.Parallel()
	variant := variantFolder(t)
	schema, err := jsonschema.CompileJSON(schemas.RunComparison())
	if err != nil {
		t.Fatal(err)
	}
	first := comparisonText(t, testsupport.ReportRun(), variant, JSONFormat)
	var doc any
	if err := model.DecodeBytes(first, &doc); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		t.Fatalf("comparison does not match run-comparison-v0.1:\n%s", jsonschema.FormatErrors(errs))
	}
	second := comparisonText(t, copyFolder(t, testsupport.ReportRun()), copyFolder(t, variant), JSONFormat)
	if !bytes.Equal(first, second) {
		t.Fatal("the comparison of two copies of the same folders differs")
	}
}

func TestTheVariantShowsTheClearedFaultTheMissingCallAndTheUnverifiedFilesAsTheyAre(t *testing.T) {
	t.Parallel()
	built, err := Compare(t.Context(), testsupport.ReportRun(), variantFolder(t))
	if err != nil {
		t.Fatal(err)
	}
	plan := built.Plan
	if len(plan.Faults.Different) != 1 || len(plan.Faults.Different[0].Differences) != 2 || len(plan.EffectorCalls.OnlyInA) != 1 || len(plan.EffectorCalls.Identical) != 1 {
		t.Fatalf("plan = %+v, want the fault cleared in B (cleared_by and cleared_ns), act-2 only in A and act-1 identical", plan)
	}
	codes := []string{}
	for _, limit := range built.Limits {
		codes = append(codes, limit.Code+limit.Side)
	}
	if !strings.Contains(strings.Join(codes, " "), "files_unverifiedb") || strings.Contains(strings.Join(codes, " "), "files_unverifieda") {
		t.Fatalf("limits = %v, want files_unverified for b only", codes)
	}
}

func TestARunThatHasNotEndedCannotBeCompared(t *testing.T) {
	t.Parallel()
	open := copyFolder(t, testsupport.ReportRun())
	if err := os.Remove(filepath.Join(open, "run.json")); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, a, b, side string }{
		{"the reference is open", open, testsupport.ReportRun(), "run A"},
		{"the other run is open", testsupport.ReportRun(), open, "run B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Compare(t.Context(), tc.a, tc.b)
			if !errors.Is(err, domain.ErrNotEnded) || !strings.Contains(err.Error(), tc.side) || !strings.Contains(err.Error(), domain.TruthSealed) {
				t.Fatalf("err = %v, want ErrNotEnded naming %s and why", err, tc.side)
			}
		})
	}
}

func TestAFolderThatIsNotARunFailsWithTheMissingFileAndNamesTheSide(t *testing.T) {
	t.Parallel()
	_, err := Compare(t.Context(), testsupport.ReportRun(), t.TempDir())
	if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "run B") {
		t.Fatalf("err = %v, want fs.ErrNotExist naming run B", err)
	}
}

func TestRenderComparisonRefusesAFormatItDoesNotKnow(t *testing.T) {
	t.Parallel()
	built, err := Compare(t.Context(), testsupport.ReportRun(), testsupport.ReportRun())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Render(built, "html"); !errors.Is(err, ErrUnknownFormat) {
		t.Fatalf("err = %v, want ErrUnknownFormat", err)
	}
}

func TestAComparisonThatBreaksItsOwnContractIsNeverReturned(t *testing.T) {
	t.Parallel()
	_, err := validated(&domain.Comparison{})
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractRunComparison {
		t.Fatalf("err = %v, want a violation of %s", err, model.ContractRunComparison)
	}
	unencodable := &domain.Comparison{}
	unencodable.Plan.Environment.Identical = []domain.EnvironmentItem{{Params: map[string]any{"bad": make(chan int)}}}
	if _, err := validated(unencodable); err == nil || !strings.Contains(err.Error(), "encode comparison") {
		t.Fatalf("err = %v, want the encode failure named", err)
	}
}

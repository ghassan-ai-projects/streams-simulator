package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const capturedPump = `{"kind":"pump","metrics":[{"name":"flow","data_type":"number","unit":"m3/h","writable":false}],"commands":[]}`

func birthText(t *testing.T) string {
	t.Helper()
	templates := []any{map[string]any{"kind": "pump", "metrics": []any{map[string]any{"name": "flow", "data_type": "number", "unit": "m3/h", "writable": false}}, "commands": []any{}},
		map[string]any{"kind": "valve", "metrics": []any{}, "commands": []any{}}}
	record, err := domain.NodeBirthOf(domain.NodeIdentity{Environment: "hardware", ProtocolMax: 1, PriorLastSeq: -1}, templates, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := domain.NewTransmitter("s-1")
	out.Queue(record)
	raw, err := json.Marshal(out.Drain()[0])
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func statuses(captured domain.Capture) map[string]string {
	found := map[string]string{}
	for _, kind := range captured.Kinds {
		found[kind.Kind] = kind.Comparison.Status
	}
	return found
}

func TestCaptureStreamComparesEachKindWithTheWorkspaceFileOfItsName(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": capturedPump, "valve.kind.json": `{"kind":"valve","metrics":[{"name":"position","data_type":"number"}],"commands":[]}`})
	captured, err := CaptureStream(strings.NewReader(birthText(t)), []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	got := statuses(captured)
	if got["pump"] != domain.CompareSame || got["valve"] != domain.CompareDiffers || captured.Kinds[1].Comparison.File != filepath.Join(dir, "valve.kind.json") {
		t.Fatalf("comparison = %v, want pump the same and valve different, naming the workspace file", got)
	}
	empty, err := CaptureStream(strings.NewReader(birthText(t)), []string{t.TempDir()})
	if got := statuses(empty); err != nil || got["pump"] != domain.CompareNew || got["valve"] != domain.CompareNew {
		t.Fatalf("comparison with an empty directory = %v, %v, want every kind new", got, err)
	}
}

func TestCaptureWithoutAKindsDirectoryComparesNothing(t *testing.T) {
	t.Parallel()
	captured, err := CaptureStream(strings.NewReader(birthText(t)), nil)
	if err != nil || len(captured.Kinds) != 2 || captured.Kinds[0].Comparison != nil || captured.Kinds[1].Comparison != nil {
		t.Fatalf("CaptureStream = %+v, %v, want the two kinds without a comparison", captured, err)
	}
}

func TestCaptureReadsAnythingTheWorkspaceHoldsAboutAKindOnlyForTheKindsItCaptured(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": capturedPump, "other.kind.json": "{"})
	if _, err := CaptureStream(strings.NewReader(birthText(t)), []string{dir}); err != nil {
		t.Fatalf("CaptureStream = %v, want a broken file of another kind left unread", err)
	}
	broken := kindsDirectoryWith(t, map[string]string{"pump.kind.json": "{"})
	_, err := CaptureStream(strings.NewReader(birthText(t)), []string{broken})
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractKindFile || invalid.Source != filepath.Join(broken, "pump.kind.json") {
		t.Fatalf("CaptureStream = %v, want the kind file violation of the broken pump file", err)
	}
	if _, err := CaptureStream(strings.NewReader(birthText(t)), []string{filepath.Join(broken, "absent")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("CaptureStream of a missing kinds directory = %v, want fs.ErrNotExist", err)
	}
}

func TestCaptureNamesItsSourceInTheViolationsItReports(t *testing.T) {
	t.Parallel()
	_, err := CaptureStream(strings.NewReader(`{"type":"DDATA"}`), nil)
	invalid, ok := model.AsValidationError(err)
	if !ok || !errors.Is(err, domain.ErrNoBirth) || invalid.Source != files.StreamName || !strings.HasPrefix(err.Error(), files.StreamName+": ") {
		t.Fatalf("CaptureStream = %v, want ErrNoBirth from %s", err, files.StreamName)
	}
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	testsupport.WriteFileIn(t, filepath.Dir(path), "journal.jsonl", []byte(`{"type":"DDATA"}`))
	_, err = CaptureFile(path, nil)
	if invalid, ok = model.AsValidationError(err); !ok || !errors.Is(err, domain.ErrNoBirth) || invalid.Source != path {
		t.Fatalf("CaptureFile = %v, want ErrNoBirth from %s", err, path)
	}
}

func TestCaptureFileReadsAJournalBeyondTheSizeOfAKindFileButNotBeyondItsOwnLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	wrapped := `{"stream":"observer","seq":2,"type":"wire","at_ns":1,"direction":"node","record":` + birthText(t) + "}\n"
	padding := strings.Repeat(`{"stream":"observer","type":"observation","at_ns":1,"line":"x"}`+"\n", (files.MaxKindFileBytes/60)+1)
	testsupport.WriteFileIn(t, dir, "journal.jsonl", []byte(padding+wrapped))
	captured, err := CaptureFile(filepath.Join(dir, "journal.jsonl"), nil)
	if err != nil || len(captured.Kinds) != 2 {
		t.Fatalf("CaptureFile of a journal over %d bytes = %+v, %v, want its two kinds", files.MaxKindFileBytes, captured, err)
	}
	if _, err := CaptureFile(filepath.Join(dir, "absent.jsonl"), nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("CaptureFile of a missing file = %v, want fs.ErrNotExist", err)
	}
	if _, err := CaptureFile(dir, nil); !errors.Is(err, kernel.ErrNotRegular) {
		t.Fatalf("CaptureFile of a directory = %v, want kernel.ErrNotRegular", err)
	}
}

func TestCaptureStreamHoldsInputToTheLimitOfAKindFile(t *testing.T) {
	t.Parallel()
	over := strings.NewReader(birthText(t) + strings.Repeat(" ", files.MaxKindFileBytes))
	if _, err := CaptureStream(over, nil); !errors.Is(err, kernel.ErrTooLarge) || !strings.Contains(err.Error(), "<stdin> holds more than 1048576 bytes") {
		t.Fatalf("CaptureStream of input over the limit = %v, want kernel.ErrTooLarge naming the limit", err)
	}
}

func TestCaptureFileLeavesTheFileAsItFoundIt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "birth.json", []byte(birthText(t)))
	before, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CaptureFile(filepath.Join(dir, "birth.json"), nil); err != nil {
		t.Fatal(err)
	}
	if after, err := os.ReadDir(dir); err != nil || len(after) != len(before) {
		t.Fatalf("directory holds %d entries after the capture, %v, want the %d it held", len(after), err, len(before))
	}
}

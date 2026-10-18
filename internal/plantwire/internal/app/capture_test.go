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

func comparisons(captured domain.Capture) map[string]domain.KindComparison {
	found := map[string]domain.KindComparison{}
	for _, kind := range captured.Kinds {
		found[kind.Kind] = *kind.Comparison
	}
	return found
}

func TestCaptureStreamComparesEachKindWithTheWorkspaceFileOfItsNameAndAimsItAtAFile(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": capturedPump, "valve.kind.json": `{"kind":"valve","metrics":[{"name":"position","data_type":"number"}],"commands":[]}`})
	captured, err := CaptureStream(strings.NewReader(birthText(t)), []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	got := comparisons(captured)
	pump, valve := filepath.Join(dir, "pump.kind.json"), filepath.Join(dir, "valve.kind.json")
	if got["pump"].Status != domain.CompareSame || got["pump"].Target != pump || got["valve"].Status != domain.CompareDiffers || got["valve"].File != valve || got["valve"].Target != valve {
		t.Fatalf("comparison = %+v, want pump the same and valve different, each aimed at its workspace file", got)
	}
	emptyDir := t.TempDir()
	empty, err := CaptureStream(strings.NewReader(birthText(t)), []string{emptyDir})
	if got := comparisons(empty); err != nil || got["pump"].Status != domain.CompareNew || got["valve"].Target != filepath.Join(emptyDir, "valve.kind.json") {
		t.Fatalf("comparison with an empty directory = %+v, %v, want every kind new and aimed at the file it would become", got, err)
	}
}

func TestCaptureWithoutAKindsDirectoryComparesNothing(t *testing.T) {
	t.Parallel()
	captured, err := CaptureStream(strings.NewReader(birthText(t)), nil)
	if err != nil || len(captured.Kinds) != 2 || captured.Kinds[0].Comparison != nil || captured.Kinds[1].Comparison != nil {
		t.Fatalf("CaptureStream = %+v, %v, want the two kinds without a comparison", captured, err)
	}
}

func TestCaptureParsesOnlyTheKindFilesOfTheKindsItCapturedAndALameOneSpoilsOnlyItsOwnKind(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": capturedPump, "other.kind.json": "{"})
	if _, err := CaptureStream(strings.NewReader(birthText(t)), []string{dir}); err != nil {
		t.Fatalf("CaptureStream = %v, want a broken file of another kind left unread", err)
	}
	broken := kindsDirectoryWith(t, map[string]string{"pump.kind.json": "{"})
	captured, err := CaptureStream(strings.NewReader(birthText(t)), []string{broken})
	if err != nil {
		t.Fatalf("CaptureStream = %v, want a broken file of a captured kind to degrade to an issue", err)
	}
	pump, valve := captured.Kinds[0], captured.Kinds[1]
	if pump.Comparison.Status != domain.CompareUnreadable || len(pump.Issues) != 1 || pump.Issues[0].Keyword != "kind_file_unreadable" || valve.Comparison.Status != domain.CompareNew {
		t.Fatalf("pump = %+v %+v, valve = %+v, want pump unreadable with its issue and valve untouched", pump.Comparison, pump.Issues, valve.Comparison)
	}
}

func TestCaptureTakesAMissingKindsDirectoryForNoSavedKindsAndAimsEveryKindAtIt(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "kinds")
	captured, err := CaptureStream(strings.NewReader(birthText(t)), []string{missing})
	if err != nil {
		t.Fatalf("CaptureStream = %v, want a missing kinds directory to hold no kind file", err)
	}
	got := comparisons(captured)
	if got["pump"].Status != domain.CompareNew || got["pump"].Target != filepath.Join(missing, "pump.kind.json") || got["valve"].Target != filepath.Join(missing, "valve.kind.json") {
		t.Fatalf("comparison = %+v, want every kind new and aimed at a file of the missing directory", got)
	}
	if _, statErr := os.Stat(missing); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("stat = %v, want the capture to create nothing", statErr)
	}
}

func TestCaptureNamesItsSourceInTheViolationsItReports(t *testing.T) {
	t.Parallel()
	_, err := CaptureStream(strings.NewReader(`{"type":"DDATA"}`), nil)
	invalid, ok := model.AsValidationError(err)
	if !ok || !errors.Is(err, domain.ErrNoBirth) || invalid.Source != files.StreamName || !strings.HasPrefix(err.Error(), files.StreamName+": ") {
		t.Fatalf("CaptureStream = %v, want ErrNoBirth from %s", err, files.StreamName)
	}
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "journal.jsonl", []byte(`{"type":"DDATA"}`))
	path := filepath.Join(dir, "journal.jsonl")
	_, err = CaptureFile(path, nil)
	if invalid, ok = model.AsValidationError(err); !ok || !errors.Is(err, domain.ErrNoBirth) || invalid.Source != path {
		t.Fatalf("CaptureFile = %v, want ErrNoBirth from %s", err, path)
	}
}

func writeLargeJournal(t *testing.T, birth string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	line := append(bytes.Repeat([]byte("x"), 1023), '\n')
	chunk := bytes.Repeat(line, 1024)
	_, wrote := file.WriteString(`{"stream":"observer","seq":2,"type":"wire","at_ns":1,"direction":"node","record":` + birth + "}\n")
	for written := 0; wrote == nil && written <= domain.MaxScanBytes+len(chunk); written += len(chunk) {
		_, wrote = file.Write(chunk)
	}
	if err := errors.Join(wrote, file.Close()); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCaptureFileStreamsAJournalLargerThanTheScanLimitAndFindsItsEarlyBirth(t *testing.T) {
	t.Parallel()
	path := writeLargeJournal(t, birthText(t))
	if info, err := os.Stat(path); err != nil || info.Size() <= domain.MaxScanBytes {
		t.Fatalf("the synthetic journal is %v bytes, %v, want more than the scan limit", info, err)
	}
	captured, err := CaptureFile(path, nil)
	if err != nil || len(captured.Kinds) != 2 || captured.Taken.Line != 1 || len(captured.Issues) != 1 || captured.Issues[0].Keyword != "scan_truncated" {
		t.Fatalf("CaptureFile = %+v, %v, want the early birth of a journal over %d bytes and a scan_truncated issue", captured, err, domain.MaxScanBytes)
	}
}

func TestCaptureFileRefusesWhatIsNotARegularFileBeforeReadingIt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "birth.json", []byte(birthText(t)))
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(filepath.Join(dir, "birth.json"), link); err != nil {
		t.Fatal(err)
	}
	if _, err := CaptureFile(link, nil); !errors.Is(err, kernel.ErrNotRegular) {
		t.Fatalf("CaptureFile of a symbolic link = %v, want kernel.ErrNotRegular", err)
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

func TestCaptureFileLeavesTheDirectoryAsItFoundIt(t *testing.T) {
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

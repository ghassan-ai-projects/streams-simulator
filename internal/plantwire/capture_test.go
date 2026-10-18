package plantwire

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func thermalBoard(t *testing.T) *standin.Node {
	t.Helper()
	kinds, err := WithKindFiles(arduinoKindFiles(t), nil, func(files map[string]map[string]any) ([]map[string]any, error) {
		return []map[string]any{files["thermal_sensor"], files["indicator"], files["fan"]}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	node, err := standin.New(standin.Config{Session: "bench-1", Board: "mega-01", Kinds: kinds, Devices: []standin.Device{
		{Name: "zone-01", Kind: "thermal_sensor", BootID: "boot-61"}, {Name: "led-01", Kind: "indicator", BootID: "boot-61"},
		{Name: "fan-01", Kind: "fan", BootID: "boot-61"}}})
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func announcedBirth(t *testing.T) string {
	t.Helper()
	birth, err := thermalBoard(t).Birth()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(birth)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestTheStandInBoardBuiltFromTheThermalKindFilesIsAcceptedByTheReferenceHost(t *testing.T) {
	t.Parallel()
	node, reference := thermalBoard(t), host.New(host.Options{Environment: "hardware", ProtocolMax: 1, Epoch: 1})
	deliver := func(records ...map[string]any) []map[string]any {
		var replies []map[string]any
		for _, record := range records {
			raw, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			inbound, ok := wire.Decode(raw)
			if !ok || inbound.Invalid != nil {
				t.Fatalf("record %s is no valid plant wire record: %v", raw, inbound.Invalid)
			}
			replies = append(replies, reference.Receive(inbound)...)
		}
		return replies
	}
	birth, err := node.Birth()
	if err != nil {
		t.Fatal(err)
	}
	replies := deliver(birth)
	if len(replies) != 1 || replies[0]["online"] != true || !reference.Accepted() {
		t.Fatalf("host answered %v, want it to accept the hardware session", replies)
	}
	births, err := node.Accept(replies[0])
	if err != nil || len(births) != 3 {
		t.Fatalf("Accept = %v, %v, want the births of the three devices", births, err)
	}
	deliver(births...)
	if report := reference.Report(); report.Refused != "" || len(report.Quarantined) != 0 || report.Rebirths != 0 {
		t.Fatalf("host report %+v, want the board commissioned whole", report)
	}
}

var thermalGoldens = map[string]string{
	"fan":            "sha256:fe29bb90a5abc8180f1287608211e678814fb51b86ffeffd4a89f605e2fa379f",
	"indicator":      "sha256:1c44d298974f5121a23acc3a9cb2befa3e5f51cb2fcacb3ff111407f1116eb67",
	"thermal_sensor": "sha256:7d0921b1526ddbb40760f3244b7b5f0a6443a125da593a9f0fa2878a871d9d48",
}

func canonicalOf(t *testing.T, value any) string {
	t.Helper()
	raw, err := canonical.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func canonicalFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var tree any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&tree); err != nil {
		t.Fatal(err)
	}
	return canonicalOf(t, tree)
}

func TestCapturingTheStandInsBirthYieldsTheKindFilesItWasBuiltFrom(t *testing.T) {
	t.Parallel()
	dir := testsupport.ThermalBoardKindsDir()
	captured, err := CaptureStream(strings.NewReader(announcedBirth(t)), []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if captured.Environment != "hardware" || captured.BDSeq != 0 || captured.BirthsSeen != 1 || len(captured.Kinds) != 3 || len(captured.Issues) != 0 {
		t.Fatalf("capture = %+v, want the three kinds of one hardware birth and nothing to report", captured)
	}
	if len(captured.Boards) != 1 || captured.Boards[0].Board != "mega-01" || captured.Boards[0].Environment != "hardware" {
		t.Fatalf("boards = %+v, want the board mega-01 echoed with the environment it claims, hardware", captured.Boards)
	}
	for _, kind := range captured.Kinds {
		comparison, file := kind.Comparison, filepath.Join(dir, kind.Kind+".kind.json")
		if comparison == nil || comparison.Status != "same" || comparison.Target != file || comparison.File != file || len(kind.Issues) != 0 {
			t.Fatalf("kind %s = %+v (%+v), want the same as %s", kind.Kind, kind, comparison, file)
		}
		if kind.KindDigest != thermalGoldens[kind.Kind] || comparison.FileDigest != thermalGoldens[kind.Kind] {
			t.Errorf("kind %s digests to %s (file %s), want the hard-coded %s", kind.Kind, kind.KindDigest, comparison.FileDigest, thermalGoldens[kind.Kind])
		}
		if got, want := canonicalOf(t, kind.Draft), canonicalFile(t, file); got != want {
			t.Errorf("kind %s: the captured draft is not the kind file:\n%s\n%s", kind.Kind, got, want)
		}
	}
}

func TestCapturingAgainstADriftedWorkspaceNamesTheFirstDifferingMember(t *testing.T) {
	t.Parallel()
	dir := driftedKinds(t, "thermal_sensor.kind.json", `"Cel"`, `"degC"`)
	captured, err := CaptureFile(writeBirth(t, announcedBirth(t)), []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, kind := range captured.Kinds {
		got[kind.Kind] = kind.Comparison.Status
	}
	if got["thermal_sensor"] != "differs" || got["indicator"] != "same" || got["fan"] != "same" {
		t.Fatalf("comparison = %v, want only the sensor to differ", got)
	}
	if sensor := captured.Kinds[0].Comparison; sensor.DiffersAt != "/metrics/0/unit" || sensor.FileDigest == captured.Kinds[0].KindDigest {
		t.Fatalf("sensor comparison = %+v, want a difference at the unit", sensor)
	}
}

func writeBirth(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "journal.jsonl", []byte(text))
	return filepath.Join(dir, "journal.jsonl")
}

func TestABirthWhoseKindWasChangedAfterItWasDigestedIsRefusedAsAClaimThatDoesNotHold(t *testing.T) {
	t.Parallel()
	tampered := strings.Replace(announcedBirth(t), `"Cel"`, `"degC"`, 1)
	_, err := CaptureStream(strings.NewReader(tampered), nil)
	invalid, ok := model.AsValidationError(err)
	if !errors.Is(err, ErrBirthRefused) || !ok || len(invalid.Issues) != 1 || invalid.Issues[0].Keyword != "kind_digest_mismatch" {
		t.Fatalf("CaptureStream = %v, want ErrBirthRefused with one kind_digest_mismatch issue", err)
	}
}

func TestCaptureFindsTheBirthInAJournalAndRefusesTheSealedStreamAndNoBirth(t *testing.T) {
	t.Parallel()
	wrapped := func(stream string) string {
		return `{"stream":"` + stream + `","seq":2,"type":"wire","at_ns":5,"direction":"node","record":` + announcedBirth(t) + "}\n"
	}
	observation := `{"stream":"observer","seq":1,"type":"observation","at_ns":1,"line":"x"}` + "\n"
	if captured, err := CaptureFile(writeBirth(t, observation+wrapped("observer")), nil); err != nil || len(captured.Kinds) != 3 {
		t.Fatalf("CaptureFile of a journal = %+v, %v, want the three kinds", captured, err)
	}
	if _, err := CaptureFile(writeBirth(t, observation+wrapped("director")), nil); !errors.Is(err, ErrSealedStream) {
		t.Fatalf("CaptureFile of the sealed stream = %v, want ErrSealedStream", err)
	}
	if _, err := CaptureFile(writeBirth(t, observation), nil); !errors.Is(err, ErrNoBirth) {
		t.Fatalf("CaptureFile of a journal without a birth = %v, want ErrNoBirth", err)
	}
}

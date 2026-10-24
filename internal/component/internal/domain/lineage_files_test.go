package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"

	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestLineageNamesTheFilesItCannotUseInsteadOfDroppingThem(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	workspace := Workspace{Domains: []NamedFile{composedPlant(t, library, compositionJSON)},
		Kinds:        []NamedFile{workspaceFile("bad.kind.json", "{"), {Name: "denied.kind.json", Err: os.ErrPermission}, workspaceFile("empty.kind.json", "{}")},
		Twins:        []NamedFile{workspaceFile("bad.twin.json", `{"id": 1}`), {Name: "denied.twin.json", Err: os.ErrPermission}},
		Compositions: []NamedFile{workspaceFile("bad.composition.json", "["), {Name: "denied.composition.json", Err: os.ErrPermission}, workspaceFile("nameless.composition.json", "{}")},
		Runs:         []RunFolder{{ID: "r-bad", RunJSON: []byte("{")}, {ID: "r-none"}, {ID: "r-denied", Err: os.ErrPermission}, {ID: "r-odd", RunJSON: []byte(`{"domain":{"id":"x","version":"1","digest":"nope"}}`)}}}
	got := lineageOf(t, library, workspace, "")
	var codes []string
	for _, entry := range got.Unreadable {
		codes = append(codes, entry.Code+" "+entry.File)
	}
	want := []string{"kind_file_unreadable kinds/bad.kind.json", "kind_file_unreadable kinds/denied.kind.json", "kind_file_unreadable kinds/empty.kind.json",
		"twin_file_unreadable twins/bad.twin.json", "twin_file_unreadable twins/denied.twin.json",
		"composition_file_unreadable compositions/bad.composition.json", "composition_file_unreadable compositions/denied.composition.json",
		"composition_file_unreadable compositions/nameless.composition.json", "run_unreadable runs/r-bad", "run_unreadable runs/r-denied", "run_unreadable runs/r-none", "run_unreadable runs/r-odd"}
	if !slices.Equal(codes, want) {
		t.Fatalf("unreadable = %v, want %v", codes, want)
	}
}

func TestLineageOfOneDomainKeepsOnlyItsPlantAndTheKindsItImplements(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	raw, err := os.ReadFile(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	workspace := Workspace{Domains: []NamedFile{{Name: "aquaculture-pond.domain.json", Raw: raw}, twinPlant(t, library)},
		Twins: []NamedFile{workspaceFile("fan-zone.twin.json", fanZoneChoicesJSON)}, Kinds: []NamedFile{workspaceFile("zone.kind.json", zoneKindJSON), workspaceFile("fan.kind.json", fanKindJSON)}}
	workspace.Kinds = append(workspace.Kinds, workspaceFile("pump.kind.json", `{"kind": "pump"}`))
	all, only := lineageOf(t, library, workspace, ""), lineageOf(t, library, workspace, "fan-zone-twin")
	if len(all.Plants) != 2 || all.Plants[0].Domain.ID != "aquaculture-pond" || len(all.DeviceKinds) != 3 || len(only.Plants) != 1 || len(only.DeviceKinds) != 2 {
		t.Fatalf("all = %d plants first %s, only = %d plants and %d kinds, want plants ordered by id and the filter to keep one", len(all.Plants), all.Plants[0].Domain.ID, len(only.Plants), len(only.DeviceKinds))
	}
	if _, err := DescribeLineage(library, workspace, "no-such-plant"); !errors.Is(err, specs.ErrUnknownDomain) {
		t.Fatalf("DescribeLineage of an unknown domain = %v, want ErrUnknownDomain", err)
	}
}

func TestLineageRefusesADomainFileItCannotReadAndIsTheSameBytesEveryTime(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	_, err := DescribeLineage(library, Workspace{Domains: []NamedFile{workspaceFile("bad.domain.json", "{")}}, "")
	wantViolation(t, err, model.ContractDomainSpec)
	if _, err := DescribeLineage(library, Workspace{Domains: []NamedFile{{Name: "denied.domain.json", Err: os.ErrPermission}}}, ""); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("DescribeLineage of an unreadable domain = %v, want the permission error", err)
	}
	workspace := Workspace{Domains: []NamedFile{twinPlant(t, library), composedPlant(t, library, compositionJSON)}, Twins: []NamedFile{workspaceFile("fan-zone.twin.json", fanZoneChoicesJSON)},
		Kinds: []NamedFile{workspaceFile("zone.kind.json", zoneKindJSON), workspaceFile("fan.kind.json", fanKindJSON)}}
	first, second := encodedLineage(t, lineageOf(t, library, workspace, "")), encodedLineage(t, lineageOf(t, library, workspace, ""))
	if !bytes.Equal(first, second) || model.ValidateDomainLineage(first) != nil {
		t.Fatalf("two documents differ or break their contract: %v", model.ValidateDomainLineage(first))
	}
}

func encodedLineage(t *testing.T, lineage *Lineage) []byte {
	t.Helper()
	raw, err := json.Marshal(lineage)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

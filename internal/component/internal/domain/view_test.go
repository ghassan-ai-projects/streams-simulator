package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func twoMajorLibrary(t *testing.T) *Library {
	t.Helper()
	second := mutate(t, kilnJSON, func(doc map[string]any) { doc["version"] = 2; doc["title"] = "Kiln two" })
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/kiln-2.component.json", second))
	return loadLibrary(t, files.directory("lib"))
}

func TestListSortsByIdThenVersionAndNamesTheKind(t *testing.T) {
	t.Parallel()
	var got []string
	for _, entry := range twoMajorLibrary(t).List() {
		got = append(got, fmt.Sprintf("%s:%s@%d", entry.Kind, entry.ID, entry.Version))
	}
	want := "equipment:acme/kiln@1,equipment:acme/kiln@2,instrument:acme/meter@1,instrument:acme/probe@1"
	if strings.Join(got, ",") != want {
		t.Fatalf("list = %v, want %s", got, want)
	}
}

func TestListDescribesEquipmentAndInstrumentsForSearching(t *testing.T) {
	t.Parallel()
	entries := baseLibrary(t).List()
	kiln, probe := entries[0], entries[2]
	if kiln.Title != "Kiln" || kiln.Class != "static/furnace/kiln" || strings.Join(kiln.Tags, ",") != "thermal" || !strings.HasPrefix(kiln.Digest, "sha256:") {
		t.Errorf("kiln entry = %+v", kiln)
	}
	if probe.Class != "temperature" || strings.Join(probe.Tags, ",") != "TT" {
		t.Errorf("probe entry = %+v, want its family as class and its tag letters as tags", probe)
	}
}

func TestListingsAndDescriptionsNameNoFilePath(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	raw, err := json.Marshal(library.List())
	if err != nil {
		t.Fatal(err)
	}
	detail, err := library.Describe("acme/probe@1")
	if err != nil {
		t.Fatal(err)
	}
	described, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{string(raw), string(described)} {
		if strings.Contains(text, "lib/") || strings.Contains(text, ".json") || strings.Contains(text, `"file"`) {
			t.Errorf("output exposes a file path: %s", text)
		}
	}
}

func TestListGivesEmptyTagsAsAnEmptyList(t *testing.T) {
	t.Parallel()
	untagged := mutate(t, kilnJSON, func(doc map[string]any) { delete(doc, "tags") })
	files := baseFiles()
	files.equipment = []File{file("equipment/kiln.component.json", untagged)}
	if tags := loadLibrary(t, files.directory("lib")).List()[0].Tags; tags == nil || len(tags) != 0 {
		t.Fatalf("tags = %#v, want an empty list that encodes as []", tags)
	}
}

func TestDescribeResolvesAnExactReferenceAndTheHighestMajor(t *testing.T) {
	t.Parallel()
	library := twoMajorLibrary(t)
	cases := map[string]string{"acme/kiln@1": "Kiln", "acme/kiln@2": "Kiln two", "acme/kiln": "Kiln two", "acme/probe": "Probe", "acme/meter@1": "Meter"}
	for ref, title := range cases {
		detail, err := library.Describe(ref)
		if err != nil || detail.Entry.Title != title || detail.Document["title"] != title {
			t.Errorf("Describe(%s) = %+v (%v), want title %s from entry and document", ref, detail, err, title)
		}
	}
}

func TestDescribeRefusesWhatIsNotInTheLibrary(t *testing.T) {
	t.Parallel()
	library := twoMajorLibrary(t)
	cases := map[string]string{
		"acme/forge":   `"acme/forge" is not in the library`,
		"acme/kiln@3":  "acme/kiln@3 is not in the library",
		"acme/kiln@0":  "is not <namespace>/<name>@<major>",
		"acme/kiln@x":  "is not <namespace>/<name>@<major>",
		"acme/kiln@-1": "is not <namespace>/<name>@<major>",
	}
	for ref, message := range cases {
		if _, err := library.Describe(ref); err == nil || !strings.Contains(err.Error(), message) {
			t.Errorf("Describe(%s) error = %v, want %q", ref, err, message)
		}
	}
}

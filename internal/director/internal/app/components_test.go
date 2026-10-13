package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pilotComposition(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "components", "compositions", "chiller-pilot.composition.json"))
	if err != nil {
		t.Fatal(err)
	}
	var composition map[string]any
	if err := json.Unmarshal(raw, &composition); err != nil {
		t.Fatal(err)
	}
	return composition
}

func toolCode(err error) rules.Code {
	var tool *rules.ToolError
	if errors.As(err, &tool) {
		return tool.Code
	}
	return ""
}

func TestADirectorWithoutALibraryRefusesTheComponentTools(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	_, listErr := d.ComponentList()
	_, describeErr := d.ComponentDescribe("streamsim/chiller")
	_, composeErr := d.ComposeDomain(pilotComposition(t), false)
	for _, err := range []error{listErr, describeErr, composeErr} {
		if toolCode(err) != rules.CodeNoLibrary {
			t.Fatalf("want %s, got %v", rules.CodeNoLibrary, err)
		}
	}
}

func TestADirectorComposesAnInstallableDomainFromItsLibrary(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	if err := d.LoadLibrary([]string{filepath.Join(testsupport.RepositoryRoot(), "components")}); err != nil {
		t.Fatal(err)
	}
	if list, err := d.ComponentList(); err != nil || list["components"] == nil {
		t.Fatalf("list = %v, %v", list, err)
	}
	if _, err := d.ComponentDescribe("streamsim/no-such"); toolCode(err) != rules.CodeComponentInvalid {
		t.Fatalf("an unknown ref = %v", err)
	}
	if _, err := d.ComposeDomain(map[string]any{"id": "x"}, false); toolCode(err) != rules.CodeComponentInvalid {
		t.Fatalf("an invalid composition = %v", err)
	}
	composed, err := d.ComposeDomain(pilotComposition(t), true)
	if err != nil || composed["installed"] != true {
		t.Fatalf("compose = %v, %v", composed, err)
	}
	if _, err := d.ComposeDomain(pilotComposition(t), true); err != nil {
		t.Fatalf("installing the same domain twice is idempotent: %v", err)
	}
	res, err := d.CreateWorld(WorldRequest{Domain: "chiller-pilot", Seed: 1, Adapter: "native-jsonl", Sink: model.SinkInproc, TimeMode: model.TimeStepped})
	if err != nil || res["world_id"] == "" {
		t.Fatalf("a world of the installed domain = %v, %v", res, err)
	}
	if _, err := d.CatalogDescribe("aquaculture-pond"); err != nil {
		t.Fatalf("installing keeps the domains already there: %v", err)
	}
}

func TestADifferentDomainCannotTakeAnInstalledID(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	if err := d.LoadLibrary([]string{filepath.Join(testsupport.RepositoryRoot(), "components")}); err != nil {
		t.Fatal(err)
	}
	composition := pilotComposition(t)
	composition["id"] = "aquaculture-pond"
	if _, err := d.ComposeDomain(composition, true); toolCode(err) != rules.CodeDomainInvalid {
		t.Fatalf("taking an installed id = %v", err)
	}
}

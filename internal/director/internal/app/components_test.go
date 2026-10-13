package app

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pilotComposition(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PilotComposition())
	if err != nil {
		t.Fatal(err)
	}
	var composition map[string]any
	if err := json.Unmarshal(raw, &composition); err != nil {
		t.Fatal(err)
	}
	return composition
}

func withBuiltInLibrary(t *testing.T, d *Director) {
	t.Helper()
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	d.library = library
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
	withBuiltInLibrary(t, d)
	if list, err := d.ComponentList(); err != nil || list["components"] == nil {
		t.Fatalf("list = %v, %v", list, err)
	}
	if _, err := d.ComponentDescribe("streamsim/no-such"); toolCode(err) != rules.CodeComponentInvalid {
		t.Fatalf("an unknown ref = %v", err)
	}
	if _, err := d.ComposeDomain(map[string]any{"id": "x"}, false); toolCode(err) != rules.CodeCompositionInvalid || !strings.Contains(err.Error(), "/equipment: missing required property") {
		t.Fatalf("an invalid composition = %v, want composition_invalid keeping the pointers", err)
	}
	composed, err := d.ComposeDomain(pilotComposition(t), true)
	if got, _ := composed.(composedDomain); err != nil || !got.Installed || got.ID != "chiller-pilot" || got.Channels == 0 {
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
	withBuiltInLibrary(t, d)
	composition := pilotComposition(t)
	composition["id"] = "aquaculture-pond"
	if _, err := d.ComposeDomain(composition, true); toolCode(err) != rules.CodeDomainInvalid {
		t.Fatalf("taking an installed id = %v", err)
	}
}

func TestAComposedDomainThatFailsDomainValidationKeepsTheDomainContractAndItsPointer(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	withBuiltInLibrary(t, d)
	composition := pilotComposition(t)
	composition["connections"] = []any{map[string]any{"from": "no-such.port", "to": "no-such.port"}}
	_, err := d.ComposeDomain(composition, false)
	if toolCode(err) != rules.CodeCompositionInvalid || !strings.Contains(err.Error(), "/connections/0/from") {
		t.Fatalf("a connection naming no instance = %v, want composition_invalid at /connections/0/from", err)
	}
}

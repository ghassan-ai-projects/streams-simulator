package contract

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

type catalogPort struct {
	ref, name, typ string
}

func TestCatalogEveryInstrumentAttachesToAMatchingPoint(t *testing.T) {
	t.Parallel()
	lib := loadLibrary(t)
	attached := map[string]bool{}
	for _, doc := range lib.files {
		points, _ := doc["measurement_points"].([]any)
		for _, raw := range points {
			point := raw.(map[string]any)
			attached[point["tag"].(string)+"/"+point["unit_family"].(string)] = true
		}
	}
	for ref, doc := range lib.files {
		if _, instrument := doc["channel"]; instrument && !attached[doc["tag"].(string)+"/"+doc["unit_family"].(string)] {
			t.Errorf("%s (%s, %s) attaches to no measurement point of the library", ref, doc["tag"], doc["unit_family"])
		}
	}
}

func TestCatalogEveryPortTypeHasBothEnds(t *testing.T) {
	t.Parallel()
	providers, consumers := catalogPorts(t)
	for typ := range providers {
		if len(consumers[typ]) == 0 {
			t.Errorf("port type %s is provided by %v and consumed by nothing", typ, providers[typ])
		}
	}
	for typ := range consumers {
		if len(providers[typ]) == 0 {
			t.Errorf("port type %s is consumed by %v and provided by nothing", typ, consumers[typ])
		}
	}
}

func TestCatalogEveryProvidingPortConnectsToEveryConsumingPortOfItsType(t *testing.T) {
	t.Parallel()
	providers, consumers := catalogPorts(t)
	library, err := component.NewLibraryFromDirs([]string{componentsDir()})
	if err != nil {
		t.Fatal(err)
	}
	for typ, from := range providers {
		for _, source := range from {
			for _, sink := range consumers[typ] {
				if err := composeConnection(t, library, source, sink); err != nil {
					t.Errorf("%s.%s -> %s.%s: %v", source.ref, source.name, sink.ref, sink.name, err)
				}
			}
		}
	}
}

func catalogPorts(t *testing.T) (providers, consumers map[string][]catalogPort) {
	t.Helper()
	providers, consumers = map[string][]catalogPort{}, map[string][]catalogPort{}
	lib := loadLibrary(t)
	for _, ref := range sortedKeys(lib.files) {
		ports, _ := lib.files[ref]["ports"].([]any)
		for _, raw := range ports {
			port := raw.(map[string]any)
			entry := catalogPort{ref: ref, name: port["name"].(string), typ: port["type"].(string)}
			if port["direction"] == "provides" {
				providers[entry.typ] = append(providers[entry.typ], entry)
			} else {
				consumers[entry.typ] = append(consumers[entry.typ], entry)
			}
		}
	}
	return providers, consumers
}

func composeConnection(t *testing.T, library *component.Library, source, sink catalogPort) error {
	t.Helper()
	equipment := []any{
		map[string]any{"use": source.ref, "as": "src", "names": []string{"SRC-1"}},
		map[string]any{"use": sink.ref, "as": "dst", "names": []string{"DST-1"}},
	}
	connections := []any{map[string]any{"from": "SRC-1." + source.name, "to": "DST-1." + sink.name}}
	composed, err := component.Compose(library, catalogDocument(t, equipment, connections), "connection.composition.json")
	if err != nil {
		return fmt.Errorf("compose: %w", err)
	}
	if _, err = domain.Parse(composed, "connection.domain.json"); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func TestCatalogPresetsAreSmallMediumLarge(t *testing.T) {
	t.Parallel()
	for _, c := range catalogCases(t) {
		presets, _ := c.doc["presets"].(map[string]any)
		if len(presets) == 0 {
			continue
		}
		sizes := map[string]bool{}
		for name := range presets {
			sizes[name] = true
		}
		if !sizes["small"] || !sizes["medium"] || !sizes["large"] || c.doc["default_preset"] != "medium" {
			t.Errorf("%s: presets must be small, medium, large with medium by default, got %s", c.ref, strings.Join(sortedKeys(presets), ","))
		}
	}
}

package contract

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

func composeCatalogPlant(t *testing.T, c catalogCase) catalogPlant {
	t.Helper()
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	names := catalogEntityNames(c)
	composed, err := component.Compose(library, catalogComposition(t, c, names), "catalog.composition.json")
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	return catalogPlant{catalogCase: c, spec: composed.Domain, entities: names}
}

func catalogEntityNames(c catalogCase) []string {
	if kind, _ := c.doc["kind"].(map[string]any); kind["system"] == true || !usesPeers(kind) {
		return []string{"U-1"}
	}
	return []string{"U-1", "U-2"}
}

func usesPeers(kind map[string]any) bool {
	faults, _ := kind["faults"].([]any)
	return strings.Contains(jsonString(faults), `"peer_residual"`)
}

func catalogComposition(t *testing.T, c catalogCase, names []string) []byte {
	t.Helper()
	entry := map[string]any{"use": c.ref, "names": names}
	if c.preset != "" {
		entry["preset"] = c.preset
	}
	if optional := optionalPoints(c.doc); len(optional) > 0 {
		entry["instruments"] = map[string]any{"add": optional}
	}
	return catalogDocument(t, []any{entry}, nil)
}

func catalogDocument(t *testing.T, equipment, connections []any) []byte {
	t.Helper()
	doc := map[string]any{
		"schema_version": "0.1", "id": "catalog-check", "version": "0.1.0", "title": "Catalog check",
		"stresses": "one component of the library, composed alone, to prove it on its own",
		"axes": map[string]any{"absence": "signal", "actuation": "advisory", "cadence": []string{"periodic"}, "cardinality": "small",
			"consequence": []string{"cost"}, "correlation": []string{"independent"}, "fidelity": []string{"F0", "F1"}, "lateness": "none",
			"rate": "low", "seasonality": []string{"diurnal"}, "time_reference": "wall", "value_shape": []string{"scalar"}},
		"equipment": equipment,
		"profiles": []any{
			map[string]any{"name": "nominal", "duration_s": 7200, "description": "Two hours of the component alone."},
			map[string]any{"name": "sensor_pathology", "duration_s": 7200, "description": "An instrument lies while the process is healthy."},
			map[string]any{"name": "correlated_cascade", "not_applicable": "One component alone has nothing to cascade into."}},
		"ground_truth": map[string]any{"min_per_label": 10, "min_scenarios": 100, "negative_class_fraction": 0.35,
			"pre_degraded_fraction": 0.0},
	}
	if connections != nil {
		doc["connections"] = connections
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func optionalPoints(doc map[string]any) []string {
	var names []string
	points, _ := doc["measurement_points"].([]any)
	for _, raw := range points {
		if point := raw.(map[string]any); point["optional"] == true {
			names = append(names, point["name"].(string))
		}
	}
	return names
}

package component

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestComposingTheFactoryWritesBothPortsOnEveryLink(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.FactoryComposition())
	if err != nil {
		t.Fatal(err)
	}
	var composition struct {
		Connections []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"connections"`
	}
	if err := json.Unmarshal(raw, &composition); err != nil {
		t.Fatal(err)
	}
	composed, err := Compose(builtInLibrary(t), raw, "factory-utilities.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	links := composed.Domain.Spec.Links
	if len(links) == 0 || len(links) != len(composition.Connections) {
		t.Fatalf("%d links for %d connections, want one each", len(links), len(composition.Connections))
	}
	for i, connection := range composition.Connections {
		if links[i].FromPort == "" || links[i].ToPort == "" || !strings.HasSuffix(connection.From, "."+links[i].FromPort) || !strings.HasSuffix(connection.To, "."+links[i].ToPort) {
			t.Errorf("link %d = %+v, want the ports of %s -> %s", i, links[i], connection.From, connection.To)
		}
	}
}

func TestTheAirDryerPassesCompressedAirOnThroughItsOutlet(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.FactoryComposition())
	if err != nil {
		t.Fatal(err)
	}
	var composition map[string]any
	if err := json.Unmarshal(raw, &composition); err != nil {
		t.Fatal(err)
	}
	composition["connections"] = append(withoutAirPath(composition["connections"].([]any)),
		map[string]any{"from": "AC-1.air_out", "to": "D-1.air_in"}, map[string]any{"from": "D-1.air_out", "to": "R-1.air_in"})
	raw, err = json.Marshal(composition)
	if err != nil {
		t.Fatal(err)
	}
	composed, err := Compose(builtInLibrary(t), raw, "dryer-in-line.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, link := range composed.Domain.Spec.Links {
		found = found || link.FromPort == "air_out" && link.ToPort == "air_in" && strings.HasPrefix(link.From, "D-1") && strings.HasPrefix(link.To, "R-1")
	}
	if !found {
		t.Fatalf("no link from the dryer's air outlet to the receiver in %+v", composed.Domain.Spec.Links)
	}
}

func withoutAirPath(connections []any) []any {
	var kept []any
	for _, item := range connections {
		link, _ := item.(map[string]any)
		from, _ := link["from"].(string)
		to, _ := link["to"].(string)
		if to != "R-1.air_in" && from != "R-1.air_out" {
			kept = append(kept, item)
		}
	}
	return kept
}

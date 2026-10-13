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

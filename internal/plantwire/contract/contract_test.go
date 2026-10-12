package contract

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheEmbeddedSchemasAreTheContractsSchemas(t *testing.T) {
	t.Parallel()
	master := filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "schemas")
	entries, err := os.ReadDir(master)
	if err != nil {
		t.Fatal(err)
	}
	embedded, err := fs.ReadDir(Schemas, "schemas")
	if err != nil || len(embedded) != len(entries) {
		t.Fatalf("embedded %d schemas, contract has %d: %v", len(embedded), len(entries), err)
	}
	for _, entry := range entries {
		want, err := os.ReadFile(filepath.Join(master, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		got, err := fs.ReadFile(Schemas, "schemas/"+entry.Name())
		if err != nil || string(got) != string(want) {
			t.Errorf("%s differs from the contract: copy contracts/plant-wire/schemas into internal/plantwire/contract/schemas", entry.Name())
		}
	}
}

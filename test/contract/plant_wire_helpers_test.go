package contract

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func plantWireDir() string {
	return filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire")
}

func readJSON(t *testing.T, path string) any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return value
}

func schemaFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(plantWireDir(), "schemas", "*.schema.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no schemas: %v", err)
	}
	slices.Sort(files)
	return files
}

func compiledSchemas(t *testing.T) map[string]*jsonschema.Schema {
	t.Helper()
	schemas := map[string]*jsonschema.Schema{}
	for _, file := range schemaFiles(t) {
		schema, err := jsonschema.Compile(readJSON(t, file))
		if err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		schemas[strings.TrimSuffix(filepath.Base(file), ".schema.json")] = schema
	}
	return schemas
}

func contractDigest(t *testing.T) string {
	t.Helper()
	var body bytes.Buffer
	for _, file := range schemaFiles(t) {
		canonicalJSON, err := canonical.Marshal(readJSON(t, file))
		if err != nil {
			t.Fatal(err)
		}
		body.WriteString(filepath.Base(file) + "\n")
		body.Write(canonicalJSON)
		body.WriteString("\n")
	}
	return canonical.DigestBytes(body.Bytes())
}

func fixtures(t *testing.T, folder, pattern string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(plantWireDir(), "fixtures", folder, pattern))
	if err != nil || len(files) == 0 {
		t.Fatalf("no %s fixtures: %v", folder, err)
	}
	return files
}

func member(value any, keys ...string) any {
	for _, key := range keys {
		object, _ := value.(map[string]any)
		value = object[key]
	}
	return value
}

package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/contract"
)

var schemas = sync.OnceValue(compileSchemas)

func compileSchemas() map[string]*jsonschema.Schema {
	compiled := map[string]*jsonschema.Schema{}
	for _, name := range schemaNames() {
		raw, _ := fs.ReadFile(contract.Schemas, "schemas/"+name)
		var doc any
		_ = decode(raw, &doc)
		schema, err := jsonschema.Compile(doc)
		if err != nil {
			panic(fmt.Sprintf("plantwire: embedded schema %s: %v", name, err))
		}
		compiled[strings.TrimSuffix(name, ".schema.json")] = schema
	}
	return compiled
}

func Parse(line []byte) (map[string]any, error) {
	var record map[string]any
	if err := decode(line, &record); err != nil {
		return nil, fmt.Errorf("plantwire: not a JSON record: %w", err)
	}
	return record, Check(record)
}

func Check(record map[string]any) error {
	kind, _ := record["type"].(string)
	schema := recordSchema(kind)
	if schema == nil {
		return fmt.Errorf("plantwire: unknown record type %q", kind)
	}
	if errs := schema.Validate(record); len(errs) > 0 {
		return fmt.Errorf("plantwire: %s fails its schema: %w", kind, errs[0])
	}
	return nil
}

func decode(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

const kindFileSchema = "KIND"

func recordSchema(name string) *jsonschema.Schema {
	if name == kindFileSchema {
		return nil
	}
	return schemas()[name]
}

func CheckKind(entry map[string]any) error {
	if errs := schemas()[kindFileSchema].Validate(entry); len(errs) > 0 {
		return fmt.Errorf("plantwire: kind file fails its schema: %w", errs[0])
	}
	return nil
}

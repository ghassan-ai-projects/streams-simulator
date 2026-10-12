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

type Inbound struct {
	Record  map[string]any
	Invalid error
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
	schema := schemas()[kind]
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

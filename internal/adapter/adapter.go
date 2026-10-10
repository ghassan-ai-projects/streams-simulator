// Package adapter implements the declarative output adapter engine. An
// adapter is data — a JSON file projecting native sim-event-v0.1 records
// into a consumer's wire format through a closed set of transforms. The
// binary contains no consumer-specific code, no consumer's field names and
// no consumer's framing rules; adding a consumer must never require a
// release.
package adapter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

// Load reads and validates an adapter file. The validation covers the
// output-adapter schema plus adapter-specific cross-checks (transform
// arity, source names, identity preservation).
func Load(path string) (*model.Adapter, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("adapter: read %s: %w", path, err)
	}
	return decodeAdapter(raw, path, "\n  ")
}

// LoadBytes validates an adapter from an in-memory source document. It is
// used by artifact replay so a run can carry its own adapter definition.
func LoadBytes(raw []byte, src string) (*model.Adapter, error) {
	return decodeAdapter(raw, src, " ")
}

func decodeAdapter(raw []byte, src, separator string) (*model.Adapter, error) {
	if err := validateAdapterDocument(raw, src, separator); err != nil {
		return nil, err
	}
	var a model.Adapter
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("adapter: %s: decode: %w", src, err)
	}
	if err := crossCheck(&a, src); err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	a.Raw = append([]byte(nil), raw...)
	return &a, nil
}

func validateAdapterDocument(raw []byte, src, separator string) error {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return fmt.Errorf("adapter: %s: not valid JSON: %w", src, err)
	}
	return validateAdapterSchema(doc, src, separator)
}

func validateAdapterSchema(doc any, src, separator string) error {
	schema, err := jsonschema.CompileJSON(schemas.OutputAdapter())
	if err != nil {
		return fmt.Errorf("adapter: compile contract schema: %w", err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		return fmt.Errorf("adapter: %s fails output-adapter-v0.1 validation:%s%s", src, separator, jsonschema.FormatErrors(errs))
	}
	return nil
}

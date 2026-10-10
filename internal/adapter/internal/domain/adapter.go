// Package domain holds the declarative output-adapter rules: validation of
// an adapter document against its schema and cross-checks, and the engine
// that projects native sim-event-v0.1 records into a consumer's wire format
// through a closed set of transforms. An adapter is data: this package has no
// consumer-specific code and performs no I/O.
package domain

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

// DecodeFile validates an adapter document read from a file: the same checks
// as LoadBytes, with the multi-line error layout used for files.
func DecodeFile(raw []byte, path string) (*model.Adapter, error) {
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

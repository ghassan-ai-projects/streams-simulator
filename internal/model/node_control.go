package model

import (
	"fmt"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

var interruptRecordSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return jsonschema.CompileJSON(schemas.InterruptLog())
})

func ValidateNodeControl(raw []byte) error {
	return validateAgainst(raw, "node control request", ContractNodeControl, schemas.NodeControl())
}

func ValidateInterruptRecord(raw []byte) error {
	var doc any
	if err := DecodeBytes(raw, &doc); err != nil {
		return NotJSONError(ContractInterruptLog, "", "model: interrupt log record not valid JSON", err)
	}
	schema, err := interruptRecordSchema()
	if err != nil {
		return fmt.Errorf("model: compile interrupt log schema: %w", err)
	}
	if errs := schema.Validate(doc); len(errs) > 0 {
		summary := fmt.Sprintf("model: interrupt log record fails %s: %s", ContractInterruptLog, errs[0].Error())
		return &ValidationError{Contract: ContractInterruptLog, Summary: summary, Issues: SchemaIssues(errs)}
	}
	return nil
}

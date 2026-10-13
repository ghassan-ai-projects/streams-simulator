package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

type contractSchema struct {
	contract model.Contract
	schema   *jsonschema.Schema
}

type contracts struct {
	component, instrument, composition, vocabulary contractSchema
}

func compileContracts() (*contracts, error) {
	component, e1 := compileContract(model.ContractComponent, schemas.Component())
	instrument, e2 := compileContract(model.ContractInstrument, schemas.Instrument())
	composition, e3 := compileContract(model.ContractComposition, schemas.Composition())
	vocabulary, e4 := compileContract(model.ContractVocabulary, schemas.ComponentVocabulary())
	if err := errors.Join(e1, e2, e3, e4); err != nil {
		return nil, err
	}
	return &contracts{component, instrument, composition, vocabulary}, nil
}

func compileContract(contract model.Contract, raw []byte) (contractSchema, error) {
	compiled, err := jsonschema.CompileJSON(raw)
	if err != nil {
		return contractSchema{}, fmt.Errorf("component: compile %s: %w", contract, err)
	}
	return contractSchema{contract: contract, schema: compiled}, nil
}

func (c contractSchema) document(raw []byte, source string) (map[string]any, error) {
	var doc map[string]any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, model.NotJSONError(c.contract, source, fmt.Sprintf("component: %s: not valid JSON", source), err)
	}
	if errs := c.schema.Validate(doc); len(errs) > 0 {
		return nil, violation(c.contract, source, model.SchemaIssues(errs))
	}
	return doc, nil
}

func (c contractSchema) decode(raw []byte, source string, target any) error {
	if err := model.DecodeBytes(raw, target); err != nil {
		return model.NotJSONError(c.contract, source, fmt.Sprintf("component: %s: decode", source), err)
	}
	return nil
}

type parsedFile[T any] struct {
	value    *T
	document map[string]any
	digest   string
}

func parseFile[T any](c contractSchema, raw []byte, source string) (parsedFile[T], error) {
	doc, err := c.document(raw, source)
	if err != nil {
		return parsedFile[T]{}, err
	}
	var value T
	if err := c.decode(raw, source, &value); err != nil {
		return parsedFile[T]{}, err
	}
	digest, err := canonical.Digest(doc)
	if err != nil {
		return parsedFile[T]{}, fmt.Errorf("component: %s: digest: %w", source, err)
	}
	return parsedFile[T]{value: &value, document: doc, digest: digest}, nil
}

func parseEquipment(c contractSchema, raw []byte, source string) (*Equipment, error) {
	parsed, err := parseFile[Equipment](c, raw, source)
	if err != nil {
		return nil, err
	}
	parsed.value.Document, parsed.value.Digest, parsed.value.File = parsed.document, parsed.digest, source
	return parsed.value, nil
}

func parseInstrument(c contractSchema, raw []byte, source string) (*Instrument, error) {
	parsed, err := parseFile[Instrument](c, raw, source)
	if err != nil {
		return nil, err
	}
	parsed.value.Document, parsed.value.Digest, parsed.value.File = parsed.document, parsed.digest, source
	return parsed.value, nil
}

func parseComposition(c contractSchema, raw []byte, source string) (*Composition, error) {
	parsed, err := parseFile[Composition](c, raw, source)
	if err != nil {
		return nil, err
	}
	return parsed.value, nil
}

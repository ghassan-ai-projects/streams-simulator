package domain

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

func validateDocument(raw []byte, src string) (any, error) {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, model.NotJSONError(model.ContractDomainSpec, src, fmt.Sprintf("domain: %s: not valid JSON", src), err)
	}
	sch, err := jsonschema.CompileJSON(schemas.DomainSpec())
	if err != nil {
		return nil, fmt.Errorf("domain: compile contract schema: %w", err)
	}
	if errs := sch.Validate(doc); len(errs) > 0 {
		return nil, schemaViolation(src, errs)
	}
	return doc, nil
}

func compileDocument(raw []byte, doc any, src string) (*Compiled, error) {
	spec, err := decodeSpec(raw, src)
	if err != nil {
		return nil, err
	}
	c, err := compileWithDigest(spec, doc, src)
	if err != nil {
		return nil, err
	}
	c.Raw = append([]byte(nil), raw...)
	return c, nil
}

func compileWithDigest(spec *model.DomainSpec, doc any, src string) (*Compiled, error) {
	digest, err := canonical.Digest(doc)
	if err != nil {
		return nil, fmt.Errorf("domain: %s: digest: %w", src, err)
	}
	c, err := compile(spec, digest, src)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return c, nil
}

func decodeSpec(raw []byte, src string) (*model.DomainSpec, error) {
	var document model.DomainDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, model.NotJSONError(model.ContractDomainSpec, src, fmt.Sprintf("domain: %s: decode", src), err)
	}
	return lower(&document, src)
}

// schemaViolation reports a domain that fails the domain-spec schema.
func schemaViolation(src string, errs []jsonschema.Error) *model.ValidationError {
	summary := fmt.Sprintf("domain: %s fails domain-spec-v0.2 validation:\n  %s", src, jsonschema.FormatErrors(errs))
	return &model.ValidationError{Contract: model.ContractDomainSpec, Source: src, Summary: summary, Issues: model.SchemaIssues(errs)}
}

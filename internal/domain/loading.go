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
		return nil, fmt.Errorf("domain: %s: not valid JSON: %w", src, err)
	}
	sch, err := jsonschema.CompileJSON(schemas.DomainSpec())
	if err != nil {
		return nil, fmt.Errorf("domain: compile contract schema: %w", err)
	}
	if errs := sch.Validate(doc); len(errs) > 0 {
		return nil, fmt.Errorf("domain: %s fails domain-spec-v0.1 validation:\n  %s", src, jsonschema.FormatErrors(errs))
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
	var spec model.DomainSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("domain: %s: decode: %w", src, err)
	}
	return &spec, nil
}

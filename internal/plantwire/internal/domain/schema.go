package domain

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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

const (
	kindFileSchema  = "KIND"
	boardFileSchema = "BOARD"
)

func recordSchema(name string) *jsonschema.Schema {
	if name == kindFileSchema || name == boardFileSchema {
		return nil
	}
	return schemas()[name]
}

func CheckKind(entry map[string]any) error {
	if _, present := entry[kindDigestMember]; present {
		return digestInKindFile()
	}
	errs := schemas()[kindFileSchema].Validate(entry)
	if len(errs) == 0 {
		return checkKindContent(entry)
	}
	return &model.ValidationError{Contract: model.ContractKindFile, Summary: "plantwire: kind file fails its schema: " + errs[0].Error(),
		Issues: sortedByPointer(model.SchemaIssues(errs)), Cause: errs[0]}
}

func checkCanonical(entry map[string]any) error {
	_, err := KindDigest(entry)
	if err == nil {
		return nil
	}
	return &model.ValidationError{Contract: model.ContractKindFile, Summary: "plantwire: kind file has no canonical form: " + err.Error(),
		Issues: []model.Issue{model.UnlocatedRuleIssue("canonical", err.Error())}, Cause: err}
}

func sortedByPointer(issues []model.Issue) []model.Issue {
	slices.SortStableFunc(issues, func(a, b model.Issue) int { return cmp.Compare(a.Pointer, b.Pointer) })
	return issues
}

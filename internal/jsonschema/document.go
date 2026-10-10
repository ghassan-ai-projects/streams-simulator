package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// maxReportedErrors bounds how many validation errors FormatErrors prints.
const maxReportedErrors = 10

// CompileJSON decodes a JSON Schema document, keeping number literals as
// json.Number, and compiles it. It is the one way contract schemas embedded
// in the binary become validators.
func CompileJSON(raw []byte) (*Schema, error) {
	var doc any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("jsonschema: schema document is not valid JSON: %w", err)
	}
	return Compile(doc)
}

// FormatErrors renders validation errors one per line, indented two spaces,
// stopping after ten with a count of the rest.
func FormatErrors(errs []Error) string {
	var b strings.Builder
	for i, e := range errs {
		if i == maxReportedErrors {
			fmt.Fprintf(&b, "  ... and %d more\n", len(errs)-maxReportedErrors)
			break
		}
		fmt.Fprintf(&b, "  %s\n", e.Error())
	}
	return strings.TrimSuffix(b.String(), "\n")
}

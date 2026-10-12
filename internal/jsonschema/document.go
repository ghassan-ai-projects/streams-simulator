package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const maxReportedErrors = 10

func CompileJSON(raw []byte) (*Schema, error) {
	var doc any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("jsonschema: schema document is not valid JSON: %w", err)
	}
	return Compile(doc)
}

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

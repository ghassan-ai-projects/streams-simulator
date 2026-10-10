package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func loadOutputSchema(path string, src VerifyFiles) (*jsonschema.Schema, error) {
	raw, err := src.OutputSchema(path)
	if err != nil {
		return nil, fmt.Errorf("adapter: read output schema %s: %w", path, err)
	}
	return compileOutputSchema(raw)
}

func compileOutputSchema(raw []byte) (*jsonschema.Schema, error) {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	schema, err := jsonschema.Compile(doc)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	return schema, nil
}

func validateOutputRecords(records []string, schema *jsonschema.Schema, name string, res *VerifyResult) bool {
	for i, record := range records {
		if divergence := outputRecordDivergence(record, i, schema, name); divergence != "" {
			res.FirstDivergence = divergence
			return false
		}
	}
	res.SchemaOK = true
	res.RecordCount = len(records)
	return true
}

func outputRecordDivergence(record string, i int, schema *jsonschema.Schema, name string) string {
	var value any
	decoder := json.NewDecoder(strings.NewReader(record))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return fmt.Sprintf("record %d is not valid JSON: %v", i, err)
	}
	if errs := schema.Validate(value); len(errs) > 0 {
		return fmt.Sprintf("record %d fails %s: %s", i, name, errs[0].Msg)
	}
	return ""
}

func compareGolden(path string, out []byte, src VerifyFiles, res *VerifyResult) error {
	golden, err := src.Golden(path)
	if err != nil {
		return fmt.Errorf("adapter: read golden %s: %w", path, err)
	}
	if string(out) != string(golden) {
		res.FirstDivergence = firstDivergence(out, golden)
		res.Detail = fmt.Sprintf("rendered %d bytes, golden %d bytes", len(out), len(golden))
		return nil
	}
	res.GoldenMatch = true
	return nil
}

func firstDivergence(got, want []byte) string {
	line, col := 1, 1
	for i := 0; i < min(len(got), len(want)); i++ {
		if got[i] != want[i] {
			return fmt.Sprintf("first divergent byte at line %d, column %d", line, col)
		}
		advanceTextPosition(&line, &col, got[i])
	}
	return "lengths differ"
}

func advanceTextPosition(line, col *int, value byte) {
	if value == '\n' {
		*line++
		*col = 1
	} else {
		*col++
	}
}

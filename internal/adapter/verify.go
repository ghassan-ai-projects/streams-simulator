package adapter

// VerifyResult is the outcome of `streamsim adapter verify`.
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type VerifyResult struct {
	Adapter         string `json:"adapter"`
	SchemaOK        bool   `json:"schema_ok"`
	GoldenMatch     bool   `json:"golden_match"`
	RecordCount     int    `json:"record_count"`
	FirstDivergence string `json:"first_divergence,omitempty"`
	Detail          string `json:"detail,omitempty"`
}

// Verify proves an adapter correct with the consumer absent: it renders the
// shipped 12-event fixture, validates every rendered record against the
// adapter's declared output schema, and byte-compares the full output to the
// committed golden file. adapterPath is the adapter file; fixturePath is the
// native-event fixture (defaults to the embedded one); base is the directory
// conformance paths resolve against.
func Verify(adapterPath, fixturePath, base string) (*VerifyResult, error) {
	a, err := Load(adapterPath)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	fixture, err := loadFixture(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	meta := map[string]any{
		"run_id": "verify", "sim_version": "0.1.0",
		"domain_id": "fixture", "domain_version": "0.0.0",
		"world_start_time": fixture[0].EventTime, "world_end_time": fixture[len(fixture)-1].EventTime,
		"seed": float64(0),
	}
	e, err := NewEngine(a, meta)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	out, err := e.RenderRun(fixture, mustParse(fixture[len(fixture)-1].ObservedTime))
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	res := &VerifyResult{Adapter: a.ID}

	// Schema conformance: every rendered record must validate against the
	// declared output schema.
	if a.Conformance != nil && a.Conformance.OutputSchema != "" {
		schemaPath := resolvePath(base, a.Conformance.OutputSchema)
		rawSchema, err := os.ReadFile(schemaPath)
		if err != nil {
			return nil, fmt.Errorf("adapter: read output schema %s: %w", schemaPath, err)
		}
		var doc any
		if err := model.DecodeBytes(rawSchema, &doc); err != nil {
			return nil, fmt.Errorf("adapter: %w", err)
		}
		sch, err := jsonschema.Compile(doc)
		if err != nil {
			return nil, fmt.Errorf("adapter: %w", err)
		}
		records, err := splitRecords(out, a.Encoding)
		if err != nil {
			return nil, fmt.Errorf("adapter: %w", err)
		}
		for i, rec := range records {
			var v any
			dec := json.NewDecoder(strings.NewReader(rec))
			dec.UseNumber()
			if err := dec.Decode(&v); err != nil {
				res.FirstDivergence = fmt.Sprintf("record %d is not valid JSON: %v", i, err)
				return res, nil
			}
			if errs := sch.Validate(v); len(errs) > 0 {
				res.FirstDivergence = fmt.Sprintf("record %d fails %s: %s", i, filepath.Base(schemaPath), errs[0].Msg)
				return res, nil
			}
		}
		res.SchemaOK = true
		res.RecordCount = len(records)
	}

	// Golden comparison: byte-exact.
	if a.Conformance != nil && a.Conformance.Golden != "" {
		goldenPath := resolvePath(base, a.Conformance.Golden)
		golden, err := os.ReadFile(goldenPath)
		if err != nil {
			return nil, fmt.Errorf("adapter: read golden %s: %w", goldenPath, err)
		}
		if string(out) != string(golden) {
			res.FirstDivergence = firstDivergence(out, golden)
			res.Detail = fmt.Sprintf("rendered %d bytes, golden %d bytes", len(out), len(golden))
			return res, nil
		}
		res.GoldenMatch = true
	}
	return res, nil
}

func firstDivergence(got, want []byte) string {
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	line, col := 1, 1
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			return fmt.Sprintf("first divergent byte at line %d, column %d", line, col)
		}
		if got[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return "lengths differ"
}

func resolvePath(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if base == "" {
		return p
	}
	return filepath.Join(base, p)
}

func splitRecords(out []byte, encoding string) ([]string, error) {
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	var recs []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		recs = append(recs, l)
	}
	return recs, nil
}

func mustParse(ts string) int64 {
	n, _ := model.ParseTime(ts)
	return n
}

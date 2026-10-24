package domain

import (
	"encoding/json"
	"fmt"
)

func ComparisonJSON(comparison *Comparison) ([]byte, error) {
	raw, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("report: encode comparison: %w", err)
	}
	return append(raw, '\n'), nil
}

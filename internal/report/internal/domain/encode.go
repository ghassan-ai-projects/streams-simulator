package domain

import (
	"encoding/json"
	"fmt"
)

func JSON(report *Report) ([]byte, error) {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("report: encode: %w", err)
	}
	return append(raw, '\n'), nil
}

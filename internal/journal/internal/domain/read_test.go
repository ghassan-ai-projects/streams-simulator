package domain

import "fmt"

func Read(raw []byte) ([]Record, error) {
	var records []Record
	for index, line := range completeLines(raw) {
		record, err := decodeLine(line)
		if err != nil {
			return records, fmt.Errorf("journal: line %d: %w", index+1, err)
		}
		records = append(records, record)
	}
	return records, nil
}

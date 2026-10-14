package journal

import "bytes"

func Read(raw []byte) ([]Record, error) {
	lines := bytes.Split(raw, []byte{'\n'})
	var records []Record
	for index, line := range lines[:len(lines)-1] {
		record, err := DecodeLine(line, index+1)
		if err != nil {
			return records, err
		}
		records = append(records, record)
	}
	return records, nil
}

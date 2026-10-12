package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func completeLines(raw []byte) [][]byte {
	lines := bytes.Split(raw, []byte{'\n'})
	return lines[:len(lines)-1]
}

func Read(raw []byte) ([]Record, error) {
	var records []Record
	for index, line := range completeLines(raw) {
		rec, err := decodeLine(line)
		if err != nil {
			return records, fmt.Errorf("journal: line %d: %w", index+1, err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func decodeLine(line []byte) (Record, error) {
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.UseNumber()
	var fields map[string]any
	if err := decoder.Decode(&fields); err != nil || fields == nil {
		return Record{}, fmt.Errorf("not a JSON object")
	}
	return recordFromFields(fields)
}

func recordFromFields(fields map[string]any) (Record, error) {
	rec := Record{Fields: fields}
	for _, field := range []struct {
		key string
		dst *string
	}{{keyStream, &rec.Stream}, {keyType, &rec.Type}} {
		value, err := takeString(fields, field.key)
		if err != nil {
			return Record{}, err
		}
		*field.dst = value
	}
	return rec, takeNumbers(fields, &rec)
}

func takeNumbers(fields map[string]any, rec *Record) error {
	for _, field := range []struct {
		key string
		dst *int64
	}{{keySeq, &rec.Seq}, {keyAt, &rec.AtNS}} {
		value, err := takeInt(fields, field.key)
		if err != nil {
			return err
		}
		*field.dst = value
	}
	return nil
}

func takeString(fields map[string]any, key string) (string, error) {
	value, ok := fields[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("%q must be a non-empty string", key)
	}
	delete(fields, key)
	return value, nil
}

func takeInt(fields map[string]any, key string) (int64, error) {
	number, ok := fields[key].(json.Number)
	if !ok {
		return 0, fmt.Errorf("%q must be an integer", key)
	}
	value, err := number.Int64()
	if err != nil {
		return 0, fmt.Errorf("%q must be an integer", key)
	}
	delete(fields, key)
	return value, nil
}

func replayableLines(raw []byte) [][]byte {
	var kept [][]byte
	for _, line := range completeLines(raw) {
		var head struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(line, &head) != nil || !IsSession(head.Type) {
			kept = append(kept, line)
		}
	}
	return kept
}

func Digest(raw []byte) string {
	sum := sha256.New()
	for _, line := range replayableLines(raw) {
		sum.Write(line)
		sum.Write([]byte{'\n'})
	}
	return "sha256:" + hex.EncodeToString(sum.Sum(nil))
}

func FirstDifference(a, b []byte) int {
	left, right := replayableLines(a), replayableLines(b)
	for i := range min(len(left), len(right)) {
		if !bytes.Equal(left[i], right[i]) {
			return i
		}
	}
	if len(left) == len(right) {
		return -1
	}
	return min(len(left), len(right))
}

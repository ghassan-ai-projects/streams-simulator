package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

const (
	MaxFieldTraceBytes       = 64 << 20
	MaxFieldTraceRecordBytes = 8 << 20
)

var ErrFieldTrace = errors.New("run: invalid field trace")

func EncodeFieldTraceLine(ordinal uint64, record map[string]any) ([]byte, error) {
	line, err := canonical.Marshal(map[string]any{"ordinal": ordinal, "record": record})
	if err != nil {
		return nil, fmt.Errorf("%w: record %d: %w", ErrFieldTrace, ordinal, err)
	}
	if len(line) > MaxFieldTraceRecordBytes {
		return nil, fmt.Errorf("%w: record %d is %d bytes, over the %d allowed", ErrFieldTrace, ordinal, len(line), MaxFieldTraceRecordBytes)
	}
	return append(line, '\n'), nil
}

func FieldTraceDigest(raw []byte) string {
	return canonical.DigestBytes(raw)
}

func ParseFieldTrace(raw []byte) ([]map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if len(raw) > MaxFieldTraceBytes || raw[len(raw)-1] != '\n' {
		return nil, fmt.Errorf("%w: a trace is at most %d bytes and ends every line with a newline", ErrFieldTrace, MaxFieldTraceBytes)
	}
	return parseFieldTraceLines(bytes.Split(raw[:len(raw)-1], []byte("\n")))
}

func parseFieldTraceLines(lines [][]byte) ([]map[string]any, error) {
	records := make([]map[string]any, 0, len(lines))
	for i, line := range lines {
		record, err := parseFieldTraceLine(uint64(i), line)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func parseFieldTraceLine(ordinal uint64, line []byte) (map[string]any, error) {
	record, err := decodeFieldTraceLine(ordinal, line)
	if err != nil {
		return nil, err
	}
	if again, err := EncodeFieldTraceLine(ordinal, record); err != nil || !bytes.Equal(again, append(bytes.Clone(line), '\n')) {
		return nil, fmt.Errorf("%w: line %d is not the canonical form of its record", ErrFieldTrace, ordinal)
	}
	return record, nil
}

func decodeFieldTraceLine(ordinal uint64, line []byte) (map[string]any, error) {
	var doc map[string]any
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("%w: line %d is no JSON object: %w", ErrFieldTrace, ordinal, err)
	}
	record, isObject := doc["record"].(map[string]any)
	if got, _ := doc["ordinal"].(json.Number); len(doc) != 2 || !isObject || got.String() != strconv.FormatUint(ordinal, 10) {
		return nil, fmt.Errorf("%w: line %d needs exactly an ordinal of %d and a record object", ErrFieldTrace, ordinal, ordinal)
	}
	return record, nil
}

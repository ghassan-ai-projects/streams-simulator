// Package jsonvalue reads decoded JSON values the same way everywhere: numbers of any decoded form, and documents that refuse fields their type does not declare.
package jsonvalue

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func Number(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case interface{ Float64() (float64, error) }:
		parsed, err := number.Float64()
		return parsed, err == nil
	}
	return integer(value)
}

func integer(value any) (float64, bool) {
	switch number := value.(type) {
	case int:
		return float64(number), true
	case int64:
		return float64(number), true
	case uint64:
		return float64(number), true
	}
	return 0, false
}

func StrictDecoder(raw []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder
}

func DecodeStrict(raw []byte, target any) error {
	if err := StrictDecoder(raw).Decode(target); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

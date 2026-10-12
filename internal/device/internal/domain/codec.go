package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func EncodeRecord(document map[string]any) ([]byte, error) {
	if document == nil {
		return nil, fmt.Errorf("device: record is required")
	}
	if err := validateRecord(document); err != nil {
		return nil, err
	}
	return encodeRecordLine(document)
}

func DecodeRecord(frame []byte) (map[string]any, error) {
	if err := admitFrameSize(frame); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(frame))
	document, err := decodeCompleteFrame(decoder)
	if err != nil {
		return nil, err
	}

	if err := validateRecord(document); err != nil {
		return nil, err
	}
	return document, nil
}

func validateRecord(document map[string]any) error {
	messageType, ok := document["message_type"].(string)
	if !ok || messageType == "" {
		return fmt.Errorf("device: message_type is required")
	}
	schema, err := schemaForMessageType(messageType)
	if err != nil {
		return err
	}
	if errs := schema.Validate(document); len(errs) > 0 {
		return fmt.Errorf("device: validate %s record: %w", messageType, errs[0])
	}
	return nil
}

func encodeRecordLine(document map[string]any) ([]byte, error) {
	encoded, err := canonical.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("device: encode record: %w", err)
	}
	encoded = append(encoded, '\n')
	if len(encoded) > MaxFrameBytes {
		return nil, fmt.Errorf("device: record exceeds %d bytes", MaxFrameBytes)
	}
	return encoded, nil
}

func admitFrameSize(frame []byte) error {
	if len(frame) == 0 {
		return fmt.Errorf("device: frame is empty")
	}
	if len(frame) > MaxFrameBytes {
		return fmt.Errorf("device: frame exceeds %d bytes", MaxFrameBytes)
	}
	return nil
}

func decodeFrameObject(decoder *json.Decoder) (map[string]any, error) {
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("device: decode frame: %w", err)
	}
	if document == nil {
		return nil, fmt.Errorf("device: frame must be a JSON object")
	}
	return document, nil
}

func rejectTrailingFrame(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("device: frame contains trailing JSON")
		}
		return fmt.Errorf("device: decode trailing frame data: %w", err)
	}
	return nil
}

func decodeCompleteFrame(decoder *json.Decoder) (map[string]any, error) {
	document, err := decodeFrameObject(decoder)
	if err != nil {
		return nil, err
	}
	if err := rejectTrailingFrame(decoder); err != nil {
		return nil, err
	}
	return document, nil
}

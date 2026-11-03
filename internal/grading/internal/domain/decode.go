package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrNotOneLabel = errors.New("score: a single scorecard needs exactly one label")

func DecodeDocument(raw []byte, path string, validate func([]byte) error, dst any) error {
	if err := validate(raw); err != nil {
		return fmt.Errorf("streamsim: %w", model.WithSource(err, path))
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func DecodeLabels(raw []byte, path string) ([]*model.GroundTruthRecord, error) {
	items, err := labelDocuments(raw, path)
	if err != nil {
		return nil, err
	}
	labels := make([]*model.GroundTruthRecord, len(items))
	for i, item := range items {
		labels[i] = &model.GroundTruthRecord{}
		if err := DecodeDocument(item, path, model.ValidateGroundTruth, labels[i]); err != nil {
			return nil, err
		}
	}
	return labels, nil
}

func labelDocuments(raw []byte, path string) ([]json.RawMessage, error) {
	if !bytes.HasPrefix(bytes.TrimSpace(raw), []byte("[")) {
		return []json.RawMessage{raw}, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return items, nil
}

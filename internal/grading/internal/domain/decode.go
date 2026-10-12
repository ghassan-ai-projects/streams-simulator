package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// DecodeDocument checks one JSON document against its contract with validate
// and decodes it into dst; a violation is a *model.ValidationError naming the
// file.
func DecodeDocument(raw []byte, path string, validate func([]byte) error, dst any) error {
	if err := validate(raw); err != nil {
		return fmt.Errorf("streamsim: %w", model.WithSource(err, path))
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

// strictRecordDecoder decodes a JSON Lines file of one record type, refusing a
// field the record type does not declare.
func strictRecordDecoder(raw []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder
}

// DecodeLedger decodes the newline-delimited delivery ledger written by Run.End.
// A ledger is JSONL, not one JSON array; decoding it as a single JSON value
// would silently discard delivery evidence from the offline score path.
func DecodeLedger(raw []byte, path string) ([]model.LedgerRecord, error) {
	dec := strictRecordDecoder(raw)
	var ledger []model.LedgerRecord
	for {
		var record model.LedgerRecord
		err := dec.Decode(&record)
		if err == io.EOF {
			return ledger, nil
		}
		if err != nil {
			return nil, model.NotJSONError(model.ContractDeliveryLedger, path, "decode ledger "+path, err)
		}
		ledger = append(ledger, record)
	}
}

// DecodeHistory decodes the director-only per-emission state history a run
// publishes beside its artifact.
func DecodeHistory(raw []byte, path string) ([]model.StateSnapshot, error) {
	history, err := decodeHistory(strictRecordDecoder(raw))
	if err != nil {
		return nil, model.NotJSONError(model.ContractStateHistory, path, "decode state history "+path, err)
	}
	return history, nil
}

func decodeHistory(dec *json.Decoder) ([]model.StateSnapshot, error) {
	var history []model.StateSnapshot
	for {
		var snapshot model.StateSnapshot
		err := dec.Decode(&snapshot)
		if errors.Is(err, io.EOF) {
			return history, nil
		}
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		history = append(history, snapshot)
	}
}

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func DecodeDocument(raw []byte, path string, validate func([]byte) error, dst any) error {
	if err := validate(raw); err != nil {
		return fmt.Errorf("streamsim: %w", model.WithSource(err, path))
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func DecodeLedger(raw []byte, path string) ([]model.LedgerRecord, error) {
	dec := jsonvalue.StrictDecoder(raw)
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

func DecodeHistory(raw []byte, path string) ([]model.StateSnapshot, error) {
	history, err := decodeHistory(jsonvalue.StrictDecoder(raw))
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

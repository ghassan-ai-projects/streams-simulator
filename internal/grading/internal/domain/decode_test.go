package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestDecodeLedgerReadsJSONL(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"delivery_id":1,"seq":1,"delivered":true,"delivery_reason":"ok"}
{"delivery_id":2,"seq":2,"delivered":false,"delivery_reason":"dropped_by_perturbation"}
`)
	ledger, err := DecodeLedger(raw, "ledger.jsonl")
	if err != nil {
		t.Fatalf("decode ledger: %v", err)
	}
	if len(ledger) != 2 || ledger[0].DeliveryID != 1 || ledger[1].DeliveryID != 2 {
		t.Fatalf("unexpected ledger rows: %+v", ledger)
	}
}

func TestDecodeLedgerRejectsMalformedAndUndeclaredRecords(t *testing.T) {
	t.Parallel()
	for name, raw := range map[string]string{
		"not JSON":         "{\"delivery_id\":1}\nnot-json\n",
		"undeclared field": `{"delivery_id":1,"invented":true}` + "\n",
	} {
		_, err := DecodeLedger([]byte(raw), "ledger.jsonl")
		var invalid *model.ValidationError
		if !errors.As(err, &invalid) || invalid.Contract != model.ContractDeliveryLedger || invalid.Source != "ledger.jsonl" ||
			!strings.HasPrefix(err.Error(), "decode ledger ledger.jsonl: ") {
			t.Fatalf("%s: err = %v", name, err)
		}
	}
}

func TestDecodeHistoryReadsSnapshotsAndNamesTheFileItCannotDecode(t *testing.T) {
	t.Parallel()
	history, err := DecodeHistory([]byte(`{"seq":1,"time_ns":5,"entity_id":"e-1","states":{"x":1}}`+"\n"), "history.jsonl")
	if err != nil || len(history) != 1 {
		t.Fatalf("history = %+v, %v", history, err)
	}
	_, err = DecodeHistory([]byte("not-json\n"), "history.jsonl")
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Contract != model.ContractStateHistory || invalid.Source != "history.jsonl" {
		t.Fatalf("err = %v, want a state-history violation naming the file", err)
	}
}

func TestDecodeDocumentChecksTheContractThenDecodes(t *testing.T) {
	t.Parallel()
	var verdict model.Verdict
	err := DecodeDocument([]byte(`{}`), "verdict.json", model.ValidateVerdict, &verdict)
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Source != "verdict.json" {
		t.Fatalf("err = %v, want a contract violation naming verdict.json", err)
	}
	accept := func([]byte) error { return nil }
	if err := DecodeDocument([]byte("not json"), "label.json", accept, &verdict); err == nil || !strings.HasPrefix(err.Error(), "decode label.json: ") {
		t.Fatalf("an undecodable document: err = %v", err)
	}
}

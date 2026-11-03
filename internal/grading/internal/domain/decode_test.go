package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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

func TestDecodeLabelsReadsOneLabelOrAnArrayAndNamesTheFileOnFailure(t *testing.T) {
	t.Parallel()
	one := `{"scenario_id":"d/0001","domain":"d","seed":1,"entity_id":"e-1","label":"f","expected_episode":true,"injection_time_ns":0,"first_observable_time_ns":0,"unavoidable_time_ns":0,"observability":{"detector_form":"single_channel_snr","channels":["c"],"effective_sigma":1,"first_observable_snr":1,"unavoidable_snr":1},"trivial_baseline_verdict":"non_trivial"}`
	for name, raw := range map[string]string{"object": one, "array": "[" + one + "," + one + "]", "empty array": "[]"} {
		labels, err := DecodeLabels([]byte(raw), "label.json")
		want := map[string]int{"object": 1, "array": 2, "empty array": 0}[name]
		if err != nil || len(labels) != want {
			t.Fatalf("%s: labels = %d, err = %v, want %d", name, len(labels), err, want)
		}
	}
	if _, err := DecodeLabels([]byte("[1,"), "labels.json"); err == nil || !strings.HasPrefix(err.Error(), "decode labels.json: ") {
		t.Fatalf("a torn array: err = %v", err)
	}
	_, err := DecodeLabels([]byte(`[{}]`), "labels.json")
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Source != "labels.json" {
		t.Fatalf("a label that breaks the contract: err = %v", err)
	}
}

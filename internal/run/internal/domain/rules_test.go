package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestCanonicalHashIsStableForADomainAndSeed(t *testing.T) {
	t.Parallel()
	if a, b := CanonicalHash("aquaculture-pond", 7), CanonicalHash("aquaculture-pond", 7); a != b {
		t.Fatal("the hash must be a pure function of its inputs")
	}
	if CanonicalHash("aquaculture-pond", 7) == CanonicalHash("aquaculture-pond", 8) ||
		CanonicalHash("aquaculture-pond", 7) == CanonicalHash("cold-chain-transit", 7) {
		t.Fatal("domain and seed must both enter the hash")
	}
	const pinned = uint64(16477993674043595849)
	if got := CanonicalHash("aquaculture-pond", 7); got != pinned {
		t.Fatalf("run id hash drifted: %d, pinned %d", got, pinned)
	}
}

func TestAdapterDigestFollowsTheAdapterDocument(t *testing.T) {
	t.Parallel()
	a := &model.Adapter{ID: "a", Version: "1"}
	if AdapterDigest(a) != AdapterDigest(&model.Adapter{ID: "a", Version: "1"}) {
		t.Fatal("equal adapters must digest equally")
	}
	if AdapterDigest(a) == AdapterDigest(&model.Adapter{ID: "a", Version: "2"}) {
		t.Fatal("a changed version must change the digest")
	}
}

func TestDecodeArtifactRejectsTrailingJSONAndUnknownShapes(t *testing.T) {
	t.Parallel()
	if _, err := DecodeArtifact([]byte(`{"run_id":"r"} {}`)); err == nil || !strings.Contains(err.Error(), "trailing JSON") {
		t.Fatalf("trailing document: %v", err)
	}
	if _, err := DecodeArtifact([]byte(`{"run_id":`)); err == nil || !strings.Contains(err.Error(), "decode artifact") {
		t.Fatalf("truncated document: %v", err)
	}
	art, err := DecodeArtifact([]byte(`{"run_id":"r-1"}`))
	if err != nil || art.RunID != "r-1" {
		t.Fatalf("decoded = %+v (%v)", art, err)
	}
}

func TestValidateArtifactDocumentNamesThePathOfInvalidJSON(t *testing.T) {
	t.Parallel()
	if err := ValidateArtifactDocument([]byte(`{`), "run.json"); err == nil || !strings.Contains(err.Error(), "run.json") {
		t.Fatalf("err = %v", err)
	}
	if err := ValidateArtifactDocument([]byte(`{}`), "run.json"); err == nil || !strings.Contains(err.Error(), "streamsim: model: artifact fails run-artifact-v0.1: missing required property \"adapter\"") {
		t.Fatal("an empty object does not satisfy the run-artifact schema")
	}
}

func TestCommandArgumentDecodingToleratesTheShapesAReplayMeets(t *testing.T) {
	t.Parallel()
	args := map[string]any{"s": "x", "f": float64(5), "i": int64(6), "n": json.Number("7"), "bad": json.Number("x"), "m": map[string]any{"k": 1}}
	if CommandString(args, "s") != "x" || CommandString(args, "f") != "" {
		t.Fatal("CommandString returns strings only")
	}
	for key, want := range map[string]int64{"f": 5, "i": 6, "n": 7, "bad": 0, "missing": 0, "s": 0} {
		if got := CommandTime(args, key); got != want {
			t.Errorf("CommandTime(%q) = %d, want %d", key, got, want)
		}
	}
	if AsMap(args["m"])["k"] != 1 || AsMap("x") != nil {
		t.Fatal("AsMap returns maps only")
	}
}

func TestLedgerAndErrorHelpers(t *testing.T) {
	t.Parallel()
	ledger := []model.LedgerRecord{{DeliveryReason: "ok"}, {DeliveryReason: "dropped"}, {DeliveryReason: "ok"}}
	if CountLedger(ledger, "ok") != 2 || CountLedger(ledger, "ok", "dropped") != 3 || CountLedger(nil, "ok") != 0 {
		t.Fatal("CountLedger counts rows by reason")
	}
	if ErrorString(nil) != "" || ErrorString(errors.New("boom")) != "boom" {
		t.Fatal("ErrorString renders nil as empty")
	}
}

func TestCloneVerdictIsADeepCopyAndNilSafe(t *testing.T) {
	t.Parallel()
	if CloneVerdict(nil) != nil {
		t.Fatal("a nil verdict clones to nil")
	}
	original := &model.Verdict{Detections: []model.Detection{{Label: "f"}}}
	clone := CloneVerdict(original)
	clone.Detections[0].Label = "changed"
	if original.Detections[0].Label != "f" {
		t.Fatal("mutating the clone changed the original")
	}
}

func TestValidateSubmittedVerdictAppliesTheContractSchema(t *testing.T) {
	t.Parallel()
	if err := ValidateSubmittedVerdict(&model.Verdict{}); err == nil || !strings.Contains(err.Error(), "SubmitVerdict") {
		t.Fatalf("an empty verdict must fail the contract: %v", err)
	}
}

func TestAdapterDigestIsTheCanonicalDigestOfItsJSONForm(t *testing.T) {
	t.Parallel()
	adapter := &model.Adapter{ID: "a", Version: "1", Encoding: "jsonl", Title: "t"}
	got := AdapterDigest(adapter)
	raw, err := json.Marshal(adapter)
	if err != nil {
		t.Fatal(err)
	}
	var form map[string]any
	if err := json.Unmarshal(raw, &form); err != nil {
		t.Fatal(err)
	}
	want, err := canonical.Digest(form)
	if err != nil || got != want {
		t.Fatalf("AdapterDigest = %s, want the RFC 8785 digest %s (%v)", got, want, err)
	}
	adapter.Raw = []byte(`{  "id":"a"  }`)
	if again := AdapterDigest(adapter); again != got {
		t.Fatalf("the digest must not depend on how the source file was formatted: %s vs %s", again, got)
	}
}

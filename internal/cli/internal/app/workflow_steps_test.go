package app

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func TestRunOptionsPreserveDefaultsEpochZeroAndFileTarget(t *testing.T) {
	t.Parallel()
	options, err := parseRunOptions([]string{"--domain", "rotating-machinery", "--start-time", "0", "--sink", "file", "--out", "results"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if options.startTime != 0 || options.seed != 1 || options.durationS != 6*3600 || options.adapterID != "native-jsonl" || options.sinkTarget != filepath.Join("results", "trace.jsonl") {
		t.Fatalf("options=%+v", options)
	}
	options, err = parseRunOptions([]string{"--domain", "rotating-machinery", "--sink", "file"}, io.Discard)
	if err == nil {
		t.Fatalf("missing file destination accepted: %+v", options)
	}
}

func TestScriptedFaultsPrecedePerturbationsInCommandLog(t *testing.T) {
	options, err := parseRunOptions([]string{"--domain", "rotating-machinery", "--domains-dir", "../../../../domains", "--adapters-dir", "../../../../adapters"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := loadRunConfig(options)
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	options.faults = r.World.EntityIDs()[0] + "=" + cfg.Domain.Spec.Faults[0].ID + "@1"
	options.perts = "drop@2@3"
	if err := applyScriptedFaults(r, options); err != nil {
		t.Fatal(err)
	}
	if err := applyScriptedPerturbations(r, options); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	if len(art.CommandLog) != 2 || art.CommandLog[0].Op != model.OpFaultInject || art.CommandLog[1].Op != model.OpPerturbApply || art.CommandLog[1].Seq != 1 {
		t.Fatalf("commands=%+v", art.CommandLog)
	}
}

func TestManifestSignsCanonicalBodyDigest(t *testing.T) {
	t.Parallel()
	seed := make([]byte, ed25519.SeedSize)
	path := filepath.Join(t.TempDir(), "test-key")
	if err := os.WriteFile(path, []byte(hex.EncodeToString(seed)), 0o600); err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"test":true}`)
	signature, err := signManifest(body, path)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := hex.DecodeString(signature)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	public := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	if !ed25519.Verify(public, digest[:], decoded) {
		t.Fatal("signature must cover the canonical body digest")
	}
	if signature, err := signManifest(body, ""); err != nil || signature != "" {
		t.Fatalf("unsigned manifest: signature=%s error=%v", signature, err)
	}
}

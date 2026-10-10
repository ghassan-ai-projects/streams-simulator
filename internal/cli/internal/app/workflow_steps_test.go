package app

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
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
	if err == nil || err.Error() != "--sink=file requires --sink-target or --out" {
		t.Fatalf("missing file destination: options=%+v err=%v", options, err)
	}
}

func TestScriptedFaultsPrecedePerturbationsInCommandLog(t *testing.T) {
	t.Parallel()
	options, err := parseRunOptions([]string{"--domain", "rotating-machinery", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()}, io.Discard)
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
	options.faults = listFlag{r.World.EntityIDs()[0] + "=" + cfg.Domain.Spec.Faults[0].ID + "@1"}
	options.perts = listFlag{"drop@2@3"}
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

func TestScriptedEffectorsCarryTheirArgumentsAndDeterministicIds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	const pond = "site-a/pond-1"
	args := `{"pond_id":"` + pond + `","level":1}`
	runArgs := []string{"run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(),
		"--domain", pondDomain, "--seed", "3", "--duration", "600",
		"--effector", "start_aerator@" + pond + "@10@" + args,
		"--effector", "start_aerator@" + pond + "@20@" + args}
	logs := make([]string, 2)
	for i := range logs {
		out := filepath.Join(dir, "run"+string(rune('a'+i)))
		invoke(t, append(runArgs, "--out", out)...).mustSucceed(t)
		raw, err := os.ReadFile(filepath.Join(out, "run.json"))
		if err != nil {
			t.Fatal(err)
		}
		var artifact struct {
			CommandLog []map[string]any `json:"command_log"`
		}
		if err := json.Unmarshal(raw, &artifact); err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, command := range artifact.CommandLog {
			if command["op"] == model.OpEffectorInvoke {
				call, _ := command["args"].(map[string]any)
				ids = append(ids, call["command_id"].(string))
				if call["args"] == nil {
					t.Fatalf("the scripted arguments must reach the world: %v", call)
				}
			}
		}
		logs[i] = strings.Join(ids, ",")
	}
	if logs[0] != "cli-0,cli-1" || logs[1] != logs[0] {
		t.Fatalf("scripted command ids = %q and %q, want cli-0,cli-1 both times", logs[0], logs[1])
	}
}

func TestScriptedEffectorArgumentsMustBeAJSONObject(t *testing.T) {
	t.Parallel()
	_, err := parseEffectorSpec(`start_aerator@site-a/pond-1@5@[1,2]`)
	if err == nil || !strings.Contains(err.Error(), "arguments must be a JSON object") {
		t.Fatalf("err = %v", err)
	}
	spec, err := parseEffectorSpec("stop@e-1")
	if err != nil || spec.offsetS != 0 || len(spec.args) != 0 {
		t.Fatalf("spec = %+v, err = %v", spec, err)
	}
}

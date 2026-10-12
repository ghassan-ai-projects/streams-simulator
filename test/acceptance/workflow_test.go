package acceptance

// The run → replay → verify → refconsumer → score workflow through the real
// binary: the commands compose into one reproducible, gradable flow, and a
// tampered artifact is refused at the first command that notices.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const pondDomain = "aquaculture-pond"

// nominalPondLabel is a contract-valid ground-truth record for a nominal pond
// scenario: negative class, no episode expected.
const nominalPondLabel = `{"scenario_id":"aquaculture-pond/0001","domain":"aquaculture-pond","seed":7,"entity_id":"",` +
	`"label":"nominal","is_negative_class":true,"expected_episode":false,"injection_time_ns":0,` +
	`"first_observable_time_ns":0,"unavoidable_time_ns":0,"observability":{"detector_form":"single_channel_snr",` +
	`"channels":["pond.dissolved_oxygen"],"effective_sigma":0.1,"first_observable_snr":3,"unavoidable_snr":10},` +
	`"trivial_baseline_verdict":"non_trivial"}`

// binary is the freshly built streamsim executable of one test.
type binary struct {
	t    *testing.T
	path string
	root string
}

// outcome is what one invocation printed and how it exited.
type outcome struct {
	code           int
	stdout, stderr string
}

func newBinary(t *testing.T) *binary {
	t.Helper()
	root := testsupport.RepositoryRoot()
	return &binary{t: t, path: buildBinary(t, root, ""), root: root}
}

// run invokes streamsim with the shipped domains and adapters selected.
func (b *binary) run(args ...string) outcome {
	b.t.Helper()
	ctx, cancel := context.WithTimeout(b.t.Context(), 5*time.Minute)
	defer cancel()
	// #nosec G204 -- the test runs the freshly built binary with fixed arguments
	cmd := exec.CommandContext(ctx, b.path, args...)
	cmd.Dir = b.root
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			b.t.Fatalf("streamsim %v: %v", args, err)
		}
		code = exit.ExitCode()
	}
	return outcome{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func inputs() []string {
	return []string{"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()}
}

func (o outcome) document(t *testing.T) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(o.stdout), &doc); err != nil {
		t.Fatalf("stdout is not one JSON object: %v\n%s\nstderr: %s", err, o.stdout, o.stderr)
	}
	return doc
}

func (o outcome) mustSucceed(t *testing.T) outcome {
	t.Helper()
	if o.code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", o.code, o.stdout, o.stderr)
	}
	return o
}

// runPond executes a short pond run and returns its output directory.
func (b *binary) runPond(extra ...string) string {
	b.t.Helper()
	dir := b.t.TempDir()
	args := append(append([]string{"run"}, inputs()...), "--domain", pondDomain, "--seed", "7", "--duration", "900",
		"--sink", "file", "--out", dir)
	got := b.run(append(args, extra...)...).mustSucceed(b.t).document(b.t)
	if got["run_id"] == "" || got["reproducible"] != true {
		b.t.Fatalf("run result = %v", got)
	}
	return dir
}

func TestRunPublishesEvidenceThatReplayAndVerifyReproduce(t *testing.T) {
	t.Parallel()
	b := newBinary(t)
	dir := b.runPond("--perturb", "drop@1@2")
	for _, name := range []string{"run.json", "trace.jsonl", "ledger.jsonl"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("run evidence %s: %v", name, err)
		}
	}
	for _, command := range []string{"replay", "verify"} {
		got := b.run(append(append([]string{command}, inputs()...), filepath.Join(dir, "run.json"))...).mustSucceed(t).document(t)
		if got["matches"] != true || got["version_match"] != true {
			t.Fatalf("%s result = %v", command, got)
		}
	}
}

func TestRefconsumerConsumesATraceAndScoreReadsItsVerdict(t *testing.T) {
	t.Parallel()
	b := newBinary(t)
	dir := b.runPond()
	verdict := filepath.Join(dir, "verdict.json")
	consumed := b.run("refconsumer", "--trace", filepath.Join(dir, "trace.jsonl"), "--out", verdict).mustSucceed(t).document(t)
	if consumed["verdict_written"] != verdict {
		t.Fatalf("refconsumer result = %v", consumed)
	}
	label := filepath.Join(dir, "label.json")
	if err := os.WriteFile(label, []byte(nominalPondLabel), 0o600); err != nil {
		t.Fatal(err)
	}
	scored := b.run(append([]string{"score", "--run", filepath.Join(dir, "run.json"), "--label", label}, inputs()...)...).mustSucceed(t).document(t)
	if len(scored) == 0 {
		t.Fatal("score printed an empty scorecard")
	}
}

func TestVerifyFailsTheCommandWhenTheReplayDoesNotReproduceTheArtifact(t *testing.T) {
	t.Parallel()
	b := newBinary(t)
	path := filepath.Join(b.runPond(), "run.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var artifact map[string]any
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	artifact["expected_trace_digest"] = "sha256:" + strings.Repeat("0", 64)
	tampered, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, tampered, 0o600); err != nil {
		t.Fatal(err)
	}

	verified := b.run(append(append([]string{"verify"}, inputs()...), path)...)
	if verified.code != 1 || !strings.HasPrefix(verified.stderr, "streamsim: verify: replay does not reproduce the artifact") {
		t.Fatalf("verify of a tampered artifact: %+v", verified)
	}
	failure, _ := verified.document(t)["error"].(map[string]any)
	if failure["code"] != "replay_mismatch" {
		t.Fatalf("verify failure code = %v, want replay_mismatch", failure["code"])
	}
	if result, _ := failure["result"].(map[string]any); result["matches"] != false {
		t.Fatalf("verify must carry the replay report as its result: %v", failure)
	}
	replayed := b.run(append(append([]string{"replay"}, inputs()...), path)...).mustSucceed(t).document(t)
	if replayed["matches"] != false {
		t.Fatalf("replay reports the mismatch without failing: %v", replayed)
	}
}

func TestScriptedEffectorsCarryTheirArgumentsAndDeterministicIds(t *testing.T) {
	t.Parallel()
	b := newBinary(t)
	dir := t.TempDir()
	const pond = "site-a/pond-1"
	args := `{"pond_id":"` + pond + `","level":1}`
	runArgs := append(append([]string{"run"}, inputs()...), "--domain", pondDomain, "--seed", "3", "--duration", "600",
		"--effector", "start_aerator@"+pond+"@10@"+args,
		"--effector", "start_aerator@"+pond+"@20@"+args)
	logs := make([]string, 2)
	for i := range logs {
		out := filepath.Join(dir, "run"+string(rune('a'+i)))
		b.run(append(runArgs, "--out", out)...).mustSucceed(t)
		logs[i] = scriptedCommandIDs(t, filepath.Join(out, "run.json"))
	}
	if logs[0] != "cli-0,cli-1" || logs[1] != logs[0] {
		t.Fatalf("scripted command ids = %q and %q, want cli-0,cli-1 both times", logs[0], logs[1])
	}
}

// scriptedCommandIDs lists the ids of the effector invocations an artifact
// logged, and requires each to carry its arguments.
func scriptedCommandIDs(t *testing.T, artifactPath string) string {
	t.Helper()
	raw, err := os.ReadFile(artifactPath)
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
		if command["op"] != model.OpEffectorInvoke {
			continue
		}
		call, _ := command["args"].(map[string]any)
		if call["args"] == nil {
			t.Fatalf("the scripted arguments must reach the world: %v", call)
		}
		id, _ := call["command_id"].(string)
		ids = append(ids, id)
	}
	return strings.Join(ids, ",")
}

package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type recordedFolder struct {
	out, domains string
	artifact     *model.RunArtifact
}

func controlScript(c *controlled) func(h *scriptedHost, syncs int) {
	return func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", fanCommand("c-1", sliceEnd(1)))
		case 2:
			c.submit(c.t, `{"op":"approve","command_id":"c-1"}`)
		case 3:
			c.submit(c.t, estopBoard)
			h.reply("DCMD", fanCommand("c-held", sliceEnd(3)))
		case 4:
			c.submit(c.t, releaseBoard)
		}
	}
}

func controlledFolder(t *testing.T) recordedFolder {
	t.Helper()
	out, domains := t.TempDir(), mixedOptions(t, "").Simulated.DomainsDir
	spec, adapterSpec, err := run.LoadInputs(domains, testsupport.AdaptersDir(), "arduino-thermal", "native-jsonl")
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: adapterSpec, Seed: 3, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, JournalDir: out})
	if err != nil {
		t.Fatal(err)
	}
	file, err := openInterruptFile(out)
	if err != nil {
		t.Fatal(err)
	}
	c := &controlled{t: t, port: NewControlPort(), run: r, sink: file, origins: true}
	c.host = &scriptedHost{t: t, afterSync: controlScript(c)}
	c.submit(t, approveFan)
	c.play(t, 5)
	if err := file.finish(out); err != nil {
		t.Fatal(err)
	}
	r.SetInterruptLog(file.digest())
	artifact, err := r.End(out)
	if err != nil {
		t.Fatal(err)
	}
	return recordedFolder{out: out, domains: domains, artifact: artifact}
}

func (f recordedFolder) replay(t *testing.T) *run.ReplayResult {
	t.Helper()
	dirs := run.Directories{Domains: f.domains, Adapters: testsupport.AdaptersDir()}
	res, err := run.ReplayFile(t.Context(), filepath.Join(f.out, model.FileRunArtifact), dirs)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (f recordedFolder) read(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(f.out, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func (f recordedFolder) write(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.out, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestASessionWithAHoldAnApprovalAndAnEmergencyStopReplaysWithTheSameArtifactsAndALogThatReplays(t *testing.T) {
	t.Parallel()
	first, second := controlledFolder(t), controlledFolder(t)
	if res := first.replay(t); !res.Matches {
		t.Fatalf("replay = %+v, want the stream digest, the ledger, the history and the interrupt log to verify", res)
	}
	for _, name := range []string{model.FileInterruptLog, model.FileTrace, model.FileLedger, model.FileHistory} {
		if first.read(t, name) != second.read(t, name) {
			t.Errorf("%s differs between two runs of the same session", name)
		}
	}
	if first.artifact.InterruptLogDigest == "" || first.artifact.InterruptLogDigest != second.artifact.InterruptLogDigest || first.artifact.ExpectedTraceDigest != second.artifact.ExpectedTraceDigest {
		t.Fatalf("digests %q %q, want two equal runs", first.artifact.InterruptLogDigest, second.artifact.InterruptLogDigest)
	}
	replayed, err := VerifyInterruptLog(first.out)
	if err != nil || replayed.Records < 15 || replayed.Digest != first.artifact.InterruptLogDigest {
		t.Fatalf("VerifyInterruptLog = %+v, %v, want the log to replay under the digest run.json records", replayed, err)
	}
	if invoked := countInvokes(first.artifact); invoked != 3 {
		t.Fatalf("the command log holds %d effector calls, want the approved command and the two safe stops", invoked)
	}
}

func countInvokes(art *model.RunArtifact) int {
	invoked := 0
	for _, command := range art.CommandLog {
		if command.Op == model.OpEffectorInvoke {
			invoked++
		}
	}
	return invoked
}

func TestEveryControlAndDecisionIsInTheJournalAndTheLiveProjectionShowsThemToo(t *testing.T) {
	t.Parallel()
	folder := controlledFolder(t)
	log := strings.Count(folder.read(t, model.FileInterruptLog), "\n")
	if journaled := journalOfType(t, folder.out, "interrupt"); len(journaled) != log || log < 15 {
		t.Fatalf("%d interrupt records in the journal for %d in the log, want the same records in both", len(journaled), log)
	}
	var latched, pending, refused bool
	for _, snapshot := range journalOfType(t, folder.out, "live") {
		state, _ := snapshot["interrupts"].(map[string]any)
		holds, _ := state["holds"].([]any)
		waiting, _ := state["pending"].([]any)
		latched = latched || (len(holds) == 1 && holds[0].(map[string]any)["latched"] == true)
		pending = pending || len(waiting) == 1
		for _, entry := range snapshot["commands"].([]any) {
			command := entry.(map[string]any)
			refused = refused || (command["command_id"] == "c-held" && command["receipt"] == "refused" && command["reason"] == "local_control")
		}
	}
	if !latched || !pending || !refused {
		t.Fatalf("latched %v pending %v refused %v, want the live projection to show the latch, the waiting command and the local_control refusal", latched, pending, refused)
	}
}

func TestAnInterruptLogThatWasChangedFailsVerificationWithTheErrorOfWhatBrokeIt(t *testing.T) {
	t.Parallel()
	folder := controlledFolder(t)
	original := folder.read(t, model.FileInterruptLog)
	changed := strings.Replace(original, `"verdict":"queue"`, `"verdict":"pass"`, 1)
	if changed == original {
		t.Fatal("the log holds no queued command to change")
	}
	folder.write(t, model.FileInterruptLog, changed)
	if res := folder.replay(t); res.Matches || res.Reason != "file_differs" || res.File != model.FileInterruptLog {
		t.Fatalf("replay = %+v, want file_differs on %s", res, model.FileInterruptLog)
	}
	if _, err := VerifyInterruptLog(folder.out); !errors.Is(err, run.ErrInterruptLogDigest) {
		t.Fatalf("VerifyInterruptLog = %v, want ErrInterruptLogDigest", err)
	}
}

func TestADecisionChangedWithTheArtifactMadeToAgreeStillFailsTheReplayOfTheRules(t *testing.T) {
	t.Parallel()
	folder := controlledFolder(t)
	changed := strings.Replace(folder.read(t, model.FileInterruptLog), `"verdict":"queue"`, `"verdict":"pass"`, 1)
	artifact := strings.Replace(folder.read(t, model.FileRunArtifact), folder.artifact.InterruptLogDigest, canonical.DigestBytes([]byte(changed)), 1)
	folder.write(t, model.FileInterruptLog, changed)
	folder.write(t, model.FileRunArtifact, artifact)
	if _, err := VerifyInterruptLog(folder.out); !errors.Is(err, domain.ErrInterruptReplay) {
		t.Fatalf("VerifyInterruptLog = %v, want ErrInterruptReplay: the rules do not decide what the log says", err)
	}
}

func TestAnInterruptLogOfRecordsOutsideItsContractIsRefusedBeforeItIsReplayed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	line := `{"ordinal":0,"record":{"clock":"world","event":"explode","node_ns":1}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, model.FileInterruptLog), []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := VerifyInterruptLog(dir)
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractInterruptLog || !strings.Contains(err.Error(), "record 0") {
		t.Fatalf("VerifyInterruptLog = %v, want record 0 refused as a violation of %s", err, model.ContractInterruptLog)
	}
}

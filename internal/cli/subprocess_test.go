package cli

// Slice G: cross-process determinism. Two spawned `streamsim run`
// invocations from the same inputs must produce byte-identical traces,
// ledgers and state histories, and the artifact must verify in a fresh
// process. The in-process replay tests cannot prove process isolation;
// this one can.

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCrossProcessDeterminism(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns the real binary; skipped in -short mode")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "streamsim")
	buildCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// #nosec G204 -- the test builds and runs the repo's own binary
	build := exec.CommandContext(buildCtx, "go", "build", "-o", bin, "./cmd/streamsim")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, out)
	}

	run := func(dir string) {
		cmdCtx, cancel2 := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel2()
		// #nosec G204 -- the test runs the freshly built binary with fixed args
		cmd := exec.CommandContext(cmdCtx, bin, "run",
			"--domain", "aquaculture-pond",
			"--seed", "7",
			"--sink", "file", "--sink-target", filepath.Join(dir, "trace.jsonl"),
			"--out", dir,
			"--duration", "3600",
		)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("run: %v\n%s", err, out)
		}
	}
	dirA, dirB := t.TempDir(), t.TempDir()
	run(dirA)
	run(dirB)

	for _, name := range []string{"trace.jsonl", "ledger.jsonl", "world_state_history.jsonl"} {
		a, err := os.ReadFile(filepath.Join(dirA, name))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(dirB, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Fatalf("%s differs across processes (%d vs %d bytes)", name, len(a), len(b))
		}
	}
	// run.json carries wall-clock created_at, so compare the digest-bearing
	// fields instead of raw bytes.
	artA, err := os.ReadFile(filepath.Join(dirA, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	artB, err := os.ReadFile(filepath.Join(dirB, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	strip := func(b []byte) []byte {
		// Remove the created_at value (the only wall-clock field).
		parts := strings.Split(string(b), `"created_at"`)
		if len(parts) != 2 {
			return b
		}
		val := parts[1]
		if i := strings.Index(val, `"`); i >= 0 {
			val = val[i+1:]
		}
		if i := strings.Index(val, `"`); i >= 0 {
			val = val[i:]
		}
		return []byte(parts[0] + `"created_at"` + `:""` + val)
	}
	if !bytes.Equal(strip(artA), strip(artB)) {
		t.Fatal("run.json diverges across processes beyond created_at")
	}

	// The artifact from process A verifies in a fresh process.
	verifyCtx, cancel3 := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel3()
	// #nosec G204 -- the test verifies the repo's own artifact
	verify := exec.CommandContext(verifyCtx, bin, "verify", filepath.Join(dirA, "run.json"))
	verify.Dir = root
	if out, err := verify.CombinedOutput(); err != nil {
		t.Fatalf("cross-process verify failed: %v\n%s", err, out)
	}
}

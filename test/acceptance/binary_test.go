// Package acceptance holds the product flows that cross module boundaries and
// run the real binary: determinism across processes, the build identity, and
// the closed loop between a director process and an out-of-process consumer.
package acceptance

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// buildBinary compiles cmd/streamsim into the test's temporary directory;
// ldflags may be empty. The Go build cache makes repeated builds cheap.
func buildBinary(t *testing.T, root, ldflags string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "streamsim")
	buildCtx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	args := []string{"build", "-o", bin}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	// #nosec G204 -- the test builds and runs the repo's own binary
	build := exec.CommandContext(buildCtx, "go", append(args, "./cmd/streamsim")...)
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, out)
	}
	return bin
}

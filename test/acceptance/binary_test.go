package acceptance

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// binaries holds the compiled streamsim variants of this test process, one per
// ldflags value, in a directory TestMain removes when the package finishes.
var binaries = struct {
	sync.Mutex
	dir     string
	variant map[string]*variant
}{variant: map[string]*variant{}}

type variant struct {
	once sync.Once
	path string
	err  error
}

// TestMain removes the shared binary directory after the package's tests.
func TestMain(m *testing.M) {
	code := m.Run()
	if binaries.dir != "" {
		_ = os.RemoveAll(binaries.dir)
	}
	os.Exit(code)
}

// buildBinary compiles cmd/streamsim once per ldflags value (which may be
// empty) and returns its path. Tests only run the binary, so sharing one
// build between them cannot couple them; each variant is linked once instead
// of once per test.
func buildBinary(t *testing.T, root, ldflags string) string {
	t.Helper()
	built := variantFor(ldflags)
	built.once.Do(func() { built.path, built.err = compileBinary(root, ldflags) })
	if built.err != nil {
		t.Fatalf("build binary: %v", built.err)
	}
	return built.path
}

func variantFor(ldflags string) *variant {
	binaries.Lock()
	defer binaries.Unlock()
	if binaries.variant[ldflags] == nil {
		binaries.variant[ldflags] = &variant{}
	}
	return binaries.variant[ldflags]
}

func binaryDir() (string, error) {
	binaries.Lock()
	defer binaries.Unlock()
	if binaries.dir != "" {
		return binaries.dir, nil
	}
	dir, err := os.MkdirTemp("", "streamsim-acceptance-")
	if err != nil {
		return "", fmt.Errorf("binary directory: %w", err)
	}
	binaries.dir = dir
	return dir, nil
}

func compileBinary(root, ldflags string) (string, error) {
	dir, err := binaryDir()
	if err != nil {
		return "", err
	}
	bin := filepath.Join(dir, fmt.Sprintf("streamsim-%x", sha256.Sum256([]byte(ldflags))))
	buildCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	args := []string{"build", "-o", bin}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	// #nosec G204 -- the test builds and runs the repo's own binary
	build := exec.CommandContext(buildCtx, "go", append(args, "./cmd/streamsim")...)
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return bin, nil
}

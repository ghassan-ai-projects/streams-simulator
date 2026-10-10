package testsupport

import (
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestShippedFixturesExistWhereTheHelpersSayTheyDo(t *testing.T) {
	t.Parallel()
	for _, path := range []string{Domain("aquaculture-pond"), Example(), Adapter("native-jsonl"), DomainsDir(), AdaptersDir()} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("fixture %s: %v", path, err)
		}
	}
	if filepath.Base(RepositoryRoot()) == "" {
		t.Fatal("empty repository root")
	}
}

func TestSocketPathBindsAndIsRemovedWithTheTest(t *testing.T) {
	t.Parallel()
	var socket string
	// The parent's cleanup runs after the subtest and its own cleanups.
	t.Cleanup(func() {
		if _, err := os.Stat(filepath.Dir(socket)); !os.IsNotExist(err) {
			t.Errorf("socket directory must be removed with its test: %v", err)
		}
	})
	t.Run("bind", func(t *testing.T) {
		t.Parallel()
		socket = SocketPath(t)
		listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", socket)
		if err != nil {
			t.Fatalf("listen on %s: %v", socket, err)
		}
		_ = listener.Close()
	})
}

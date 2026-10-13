package canvasbridge

import (
	"bytes"
	"errors"
	"testing"
)

func TestTheFacadeServesThroughTheModuleAndRefusesANonLoopbackAddress(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	err := Serve(t.Context(), Config{Workspace: t.TempDir(), UIDir: t.TempDir(), Addr: "0.0.0.0:0", AdaptersDir: t.TempDir(), Stderr: &stderr})
	if !errors.Is(err, ErrNonLoopback) || stderr.Len() != 0 {
		t.Fatalf("Serve = %v (stderr %q), want ErrNonLoopback before anything is printed", err, stderr.String())
	}
}

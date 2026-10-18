package app

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestServingRefusesANonLoopbackAddressAfterPreparingTheWorkspace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var stderr bytes.Buffer
	err := Serve(context.Background(), Config{Workspace: dir, UIDir: t.TempDir(), Addr: "0.0.0.0:0", AdaptersDir: "adapters", Stderr: &stderr})
	if !errors.Is(err, domain.ErrNonLoopback) || stderr.Len() != 0 {
		t.Fatalf("Serve on 0.0.0.0 = %v (stderr %q), want ErrNonLoopback before anything is printed", err, stderr.String())
	}
	for _, folder := range workspaceFolders {
		if matches, _ := filepath.Glob(filepath.Join(dir, folder)); len(matches) != 1 {
			t.Errorf("the workspace folder %s was not created", folder)
		}
	}
	if err := Serve(context.Background(), Config{Workspace: filepath.Join(dir, "absent"), Addr: "127.0.0.1:0"}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing workspace was served: %v", err)
	}
}

func TestTheWiredBridgeKnowsItsAbsoluteFoldersAndAFreshCode(t *testing.T) {
	t.Parallel()
	settings, err := settingsOf(Config{Workspace: t.TempDir(), AdaptersDir: "adapters", ComponentDirs: []string{"/lib"}})
	if err != nil || !filepath.IsAbs(settings.WorkspaceDir) || !filepath.IsAbs(settings.AdaptersDir) || len(domain.NormalizeCode(settings.PairingCode)) != 12 {
		t.Fatalf("settings = %+v, %v", settings, err)
	}
	workspace, err := openWorkspace(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	falsy, err := exec.LookPath("false")
	if err != nil {
		t.Fatal(err)
	}
	bridge := wire(workspace, falsy, settings)
	if bridge.edges.RunCommand == nil || bridge.edges.StartDirector == nil || bridge.edges.Tick == nil || bridge.edges.NewSession == nil {
		t.Fatal("every edge is wired")
	}
	if result, err := bridge.edges.RunCommand(context.Background(), nil, nil); err != nil || result.ExitCode != 1 {
		t.Fatalf("the command edge runs the configured executable: %+v, %v", result, err)
	}
}

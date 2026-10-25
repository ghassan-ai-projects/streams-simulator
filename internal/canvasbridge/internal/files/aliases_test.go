package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAliasesFollowARelativeSymlinkFromItsOwnFolder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.Mkdir(real, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real", filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(evalResolved(t, root), "real")
	if got := Aliases(filepath.Join(root, "link")); len(got) != 1 || got[0] != want {
		t.Errorf("Aliases(relative symlink) = %q, want [%q]", got, want)
	}
}

func TestAliasesStopAtTheLinkDepthCapOnASymlinkLoop(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	loop := filepath.Join(root, "loop")
	if err := os.Symlink(loop, loop); err != nil {
		t.Fatal(err)
	}
	if got := Aliases(loop); len(got) > 1 {
		t.Errorf("Aliases(loop) = %q, want at most one alias and no hang", got)
	}
}

func evalResolved(t *testing.T, dir string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

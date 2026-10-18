package transport

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheBuiltUIIsServedWithAnIndexFallback(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	writeUI(t, r.uiDir, "assets/app.js", "console.log('canvas')")
	for path, want := range map[string]string{"/": "<html>canvas</html>", "/index.html": "<html>canvas</html>", "/assets/app.js": "console.log('canvas')",
		"/workspace/plant": "<html>canvas</html>", "/assets/missing.js": "<html>canvas</html>", "/assets/": "<html>canvas</html>"} {
		got := r.do(t, req{Path: path})
		if got.Status != http.StatusOK || !strings.Contains(got.Body, want) {
			t.Errorf("GET %s: %d %q, want %q", path, got.Status, got.Body, want)
		}
	}
	if got := r.do(t, req{Path: "/assets/app.js"}); !strings.Contains(got.Header.Get("Content-Type"), "javascript") {
		t.Errorf("a script is served as %q", got.Header.Get("Content-Type"))
	}
}

func TestAUIWithoutItsIndexAnswersNotFound(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	if err := os.Remove(filepath.Join(r.uiDir, "index.html")); err != nil {
		t.Fatal(err)
	}
	if got := r.do(t, req{Path: "/"}); got.Status != http.StatusNotFound {
		t.Errorf("GET / without an index: %d, want 404", got.Status)
	}
}

func TestTheUIRootCannotBeLeftByTraversalOrSymlink(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	secret := filepath.Join(filepath.Dir(r.uiDir), filepath.Base(r.uiDir)+"-secret.txt")
	if err := os.WriteFile(secret, []byte("TOP-SECRET"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(secret) })
	if err := os.Symlink(secret, filepath.Join(r.uiDir, "leak.txt")); err != nil {
		t.Fatal(err)
	}
	name := filepath.Base(secret)
	for _, path := range []string{"/../" + name, "/%2e%2e/" + name, "/..%2f" + name, "/assets/../../" + name, "/%2e%2e%2f" + name, "/leak.txt", "/" + name} {
		got := r.do(t, req{Path: path})
		if strings.Contains(got.Body, "TOP-SECRET") {
			t.Errorf("GET %s leaked a file outside the UI root", path)
		}
	}
}

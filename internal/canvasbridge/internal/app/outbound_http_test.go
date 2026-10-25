package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var announced = regexp.MustCompile(`canvas: http://(\S+)\ncanvas: pairing code (\S+)\n`)

func TestTheServedListingShowsAFailedRunsReasonWithoutTheWorkspacePath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	reason := "plantwire connect: lstat " + dir + "/sockets/absent2.sock: no such file or directory, then " + resolved + "/sockets/absent2.sock"
	writeRun(t, dir, "c-1", openedLine, `{"seed":1,"error":`+quoted(t, reason)+`}`)
	body := getWorkspace(t, dir)
	if strings.Contains(body, dir) || strings.Contains(body, resolved) {
		t.Fatalf("the served listing names the workspace path: %s", body)
	}
	if !strings.Contains(body, "plantwire connect: lstat sockets/absent2.sock: no such file or directory, then sockets/absent2.sock") {
		t.Fatalf("the served listing lost the reason: %s", body)
	}
}

func quoted(t *testing.T, text string) string {
	t.Helper()
	raw, err := json.Marshal(text)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func writeRun(t *testing.T, dir, id, journal, runJSON string) {
	t.Helper()
	folder := filepath.Join(dir, "runs", id)
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"journal.jsonl": journal, "run.json": runJSON} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func getWorkspace(t *testing.T, dir string) string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stderr := &announcement{seen: make(chan struct{})}
	go func() {
		_ = Serve(ctx, Config{Workspace: dir, UIDir: t.TempDir(), Addr: "127.0.0.1:0", AdaptersDir: "adapters", Stderr: stderr})
	}()
	<-stderr.seen
	found := announced.FindStringSubmatch(stderr.String())
	if found == nil {
		t.Fatalf("announcement = %q", stderr.String())
	}
	client := &http.Client{}
	client.Jar, _ = cookiejar.New(nil)
	base := "http://" + found[1]
	post := doRequest(ctx, t, client, http.MethodPost, base+"/api/session", `{"code":`+quoted(t, found[2])+`}`)
	if !strings.Contains(post, `"paired":true`) {
		t.Fatalf("pairing = %s", post)
	}
	return doRequest(ctx, t, client, http.MethodGet, base+"/api/workspace", "")
}

func doRequest(ctx context.Context, t *testing.T, client *http.Client, method, url, body string) string {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	var out strings.Builder
	buffer := make([]byte, 4096)
	for {
		n, readErr := response.Body.Read(buffer)
		out.Write(buffer[:n])
		if readErr != nil {
			return out.String()
		}
	}
}

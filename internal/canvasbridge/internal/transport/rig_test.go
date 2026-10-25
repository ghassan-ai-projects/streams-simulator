package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const goodSession = "good-session"

type fakeService struct {
	mu        sync.Mutex
	calls     []string
	pairErr   error
	openErr   error
	actionErr error
	fileErr   error
	etag      string
	written   [3]string
	args      map[string]any
	world     string
	after     int64
	records   []domain.Record
	followErr error
	changes   int
}

func (f *fakeService) note(call string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call)
}

func (f *fakeService) called() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calls...)
}

func (f *fakeService) Pair(code string) (string, error) {
	f.note("Pair")
	if f.pairErr != nil {
		return "", f.pairErr
	}
	return goodSession, nil
}

func (f *fakeService) Admits(session string) bool { return session == goodSession }

func (f *fakeService) Workspace(context.Context) (domain.Listing, error) {
	f.note("Workspace")
	return domain.Listing{Name: "ws"}, f.fileErr
}

func (f *fakeService) ReadFile(rel string) (domain.FileDocument, error) {
	f.note("ReadFile")
	return domain.FileDocument{Path: rel, Content: "body", Etag: "e1"}, f.fileErr
}

func (f *fakeService) WriteFile(rel, content, etag string) (domain.FileDocument, error) {
	f.note("WriteFile")
	f.mu.Lock()
	f.written = [3]string{rel, content, etag}
	f.mu.Unlock()
	if f.fileErr != nil {
		return domain.FileDocument{Path: rel, Etag: f.etag}, f.fileErr
	}
	return domain.FileDocument{Path: rel, Content: content, Etag: "e2"}, nil
}

func (f *fakeService) Actions() []domain.Action {
	f.note("Actions")
	return []domain.Action{{Name: "world.create", Kind: domain.KindMCP, Tool: "sim.world.create"}}
}

func (f *fakeService) RunAction(_ context.Context, name string, args map[string]any) (json.RawMessage, error) {
	f.note("RunAction " + name)
	f.mu.Lock()
	f.args = args
	f.mu.Unlock()
	if f.actionErr != nil {
		return nil, f.actionErr
	}
	return json.RawMessage(`{"ok":true}`), f.fileErr
}

func (f *fakeService) OpenRun(worldID string) error {
	f.note("OpenRun " + worldID)
	return f.openErr
}

func (f *fakeService) Follow(_ context.Context, worldID string, after int64, emit func(domain.Record) error) error {
	f.note("Follow")
	f.mu.Lock()
	f.world, f.after = worldID, after
	f.mu.Unlock()
	for _, record := range f.records {
		if err := emit(record); err != nil {
			return err
		}
	}
	return f.followErr
}

func (f *fakeService) Watch(_ context.Context, changed func() error) error {
	f.note("Watch")
	for range f.changes {
		if err := changed(); err != nil {
			return err
		}
	}
	return nil
}

type rig struct {
	service *fakeService
	server  *httptest.Server
	uiDir   string
	bound   string
}

func newRig(t *testing.T) *rig {
	t.Helper()
	uiDir := t.TempDir()
	writeUI(t, uiDir, "index.html", "<html>canvas</html>")
	ui, err := os.OpenRoot(uiDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ui.Close() })
	server := httptest.NewUnstartedServer(nil)
	service := &fakeService{}
	bound := server.Listener.Addr().String()
	server.Config.Handler = (&Server{service: service, bound: bound, ui: ui}).routes()
	server.Start()
	t.Cleanup(server.Close)
	return &rig{service: service, server: server, uiDir: uiDir, bound: bound}
}

func (r *rig) cookieName() string { return domain.SessionCookieName(r.bound) }

func writeUI(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

type req struct {
	Method string
	Path   string
	Body   string
	Host   string
	Paired bool
	Header map[string]string
}

type reply struct {
	Status  int
	Header  http.Header
	Cookies []*http.Cookie
	Body    string
}

func (r *rig) do(t *testing.T, q req) reply {
	t.Helper()
	method := q.Method
	if method == "" {
		method = http.MethodGet
	}
	request, err := http.NewRequestWithContext(t.Context(), method, r.server.URL+q.Path, strings.NewReader(q.Body))
	if err != nil {
		t.Fatal(err)
	}
	if q.Host != "" {
		request.Host = q.Host
	}
	if q.Paired {
		request.AddCookie(&http.Cookie{Name: r.cookieName(), Value: goodSession, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	}
	for key, value := range q.Header {
		request.Header.Set(key, value)
	}
	response, err := r.server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return reply{Status: response.StatusCode, Header: response.Header, Cookies: response.Cookies(), Body: string(body)}
}

func (p reply) errorCode(t *testing.T) string {
	t.Helper()
	var document struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(p.Body), &document); err != nil {
		t.Fatalf("not an error document: %q: %v", p.Body, err)
	}
	return document.Error.Code
}

type route struct{ method, path, body string }

var apiRoutes = []route{
	{http.MethodGet, "/api/workspace", ""},
	{http.MethodGet, "/api/file?path=x", ""},
	{http.MethodPut, "/api/file", `{"path":"x","content":"c","etag":""}`},
	{http.MethodGet, "/api/actions", ""},
	{http.MethodPost, "/api/actions/world.create", `{"args":{}}`},
	{http.MethodGet, "/api/runs/w-1/events", ""},
	{http.MethodGet, "/api/events", ""},
}

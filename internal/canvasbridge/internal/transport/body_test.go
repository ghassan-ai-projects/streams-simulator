package transport

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestOversizedBodiesAreRefusedWith413BeforeTheServiceSeesThem(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	filler := func(n int) string { return strings.Repeat("a", n) }
	cases := []struct {
		name string
		req  req
	}{
		{"a file", req{Method: http.MethodPut, Path: "/api/file", Body: `{"path":"x","content":"` + filler(maxFileBody) + `","etag":""}`}},
		{"an action", req{Method: http.MethodPost, Path: "/api/actions/world.create", Body: `{"args":{"label":"` + filler(maxActionBody) + `"}}`}},
		{"a pairing", req{Method: http.MethodPost, Path: "/api/session", Body: `{"code":"` + filler(maxActionBody) + `"}`}},
	}
	for _, tc := range cases {
		tc.req.Paired = true
		got := r.do(t, tc.req)
		if got.Status != http.StatusRequestEntityTooLarge || got.errorCode(t) != "too_large" {
			t.Errorf("%s over the limit: %d %.80s", tc.name, got.Status, got.Body)
		}
	}
	if calls := r.service.called(); len(calls) != 0 {
		t.Fatalf("an oversized request reached the service: %v", calls)
	}
}

func TestBodiesWithinTheLimitsAreAccepted(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	content := strings.Repeat("a", domain.MaxFileBytes)
	file := r.do(t, req{Method: http.MethodPut, Path: "/api/file", Paired: true, Body: `{"path":"x","content":"` + content + `","etag":""}`})
	action := r.do(t, req{Method: http.MethodPost, Path: "/api/actions/world.create", Paired: true,
		Body: `{"args":{"label":"` + strings.Repeat("a", maxActionBody/2) + `"}}`})
	if file.Status != http.StatusOK || action.Status != http.StatusOK {
		t.Fatalf("a file of the maximum size: %d, a large action: %d", file.Status, action.Status)
	}
}

func TestMalformedJSONIsRefusedWith400(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for _, route := range []route{{http.MethodPut, "/api/file", "{"}, {http.MethodPost, "/api/actions/world.create", "not json"},
		{http.MethodPost, "/api/session", ""}, {http.MethodPut, "/api/file", `{"path":["x"]}`}, {http.MethodPost, "/api/actions/world.create", `{"args":"x"}`}} {
		got := r.do(t, req{Method: route.method, Path: route.path, Body: route.body, Paired: true})
		if got.Status != http.StatusBadRequest || got.errorCode(t) != "bad_arguments" {
			t.Errorf("%s %s %q: %d %s", route.method, route.path, route.body, got.Status, got.Body)
		}
	}
	if calls := r.service.called(); len(calls) != 0 {
		t.Fatalf("malformed requests reached the service: %v", calls)
	}
}

func TestActionArgumentsReachTheServiceWithTheirNumbersExact(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	got := r.do(t, req{Method: http.MethodPost, Path: "/api/actions/clock.advance", Paired: true, Body: `{"args":{"world_id":"w-1","by_ns":9007199254740993}}`})
	r.service.mu.Lock()
	defer r.service.mu.Unlock()
	if got.Status != http.StatusOK || strings.TrimSpace(got.Body) != `{"result":{"ok":true}}` || r.service.args["by_ns"] != json.Number("9007199254740993") {
		t.Fatalf("%d %s, args %#v", got.Status, got.Body, r.service.args)
	}
}

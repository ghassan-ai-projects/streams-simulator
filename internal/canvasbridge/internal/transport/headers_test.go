package transport

import (
	"net/http"
	"testing"
)

const specifiedCSP = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; " +
	"object-src 'none'; base-uri 'none'; frame-ancestors 'none'"

func TestEveryResponseCarriesTheSecurityHeaders(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	paths := []req{{Path: "/"}, {Path: "/missing/route"}, {Path: "/api/session"}, {Path: "/api/workspace"}, {Path: "/api/workspace", Paired: true},
		{Path: "/api/session", Host: "evil.example"}, {Method: http.MethodPost, Path: "/api/session", Body: "{"}}
	want := map[string]string{"Content-Security-Policy": specifiedCSP, "X-Content-Type-Options": "nosniff", "Referrer-Policy": "no-referrer",
		"X-Frame-Options": "DENY", "Cache-Control": "no-store"}
	for _, q := range paths {
		got := r.do(t, q)
		for header, value := range want {
			if got.Header.Get(header) != value {
				t.Errorf("%s %s (%d): %s = %q, want %q", q.Method, q.Path, got.Status, header, got.Header.Get(header), value)
			}
		}
	}
}

func TestJSONRoutesAnswerWithJSON(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for _, route := range apiRoutes[:5] {
		got := r.do(t, req{Method: route.method, Path: route.path, Body: route.body, Paired: true})
		if got.Status != http.StatusOK || got.Header.Get("Content-Type") != "application/json" {
			t.Errorf("%s %s: %d %s", route.method, route.path, got.Status, got.Header.Get("Content-Type"))
		}
	}
}

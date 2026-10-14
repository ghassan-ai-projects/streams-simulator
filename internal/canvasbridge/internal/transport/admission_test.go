package transport

import (
	"net/http"
	"strings"
	"testing"
)

func TestOnlyTheBoundHostIsServedAgainstDNSRebinding(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	port := r.bound[strings.LastIndexByte(r.bound, ':')+1:]
	for host, status := range map[string]int{r.bound: http.StatusOK, "localhost:" + port: http.StatusOK, "127.0.0.1:" + port: http.StatusOK,
		"evil.example:" + port: http.StatusForbidden, "evil.example": http.StatusForbidden, "127.0.0.1:1": http.StatusForbidden,
		"localhost.evil.example:" + port: http.StatusForbidden} {
		got := r.do(t, req{Path: "/api/workspace", Host: host, Paired: true})
		if got.Status != status || (status == http.StatusForbidden && got.errorCode(t) != "not_admitted") {
			t.Errorf("Host %q: %d %s, want %d", host, got.Status, got.Body, status)
		}
	}
	if got := r.do(t, req{Path: "/api/session", Host: "evil.example"}); got.Status != http.StatusForbidden {
		t.Errorf("the session route accepted a foreign host: %d", got.Status)
	}
}

func TestCrossOriginRequestsAreRefusedOnEveryAPIRoute(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	routes := append([]route{{http.MethodGet, "/api/session", ""}, {http.MethodPost, "/api/session", `{"code":"X"}`}}, apiRoutes...)
	attacks := map[string]map[string]string{"foreign origin": {"Origin": "http://evil.example"}, "cross-site fetch": {"Sec-Fetch-Site": "cross-site"},
		"same-site fetch": {"Sec-Fetch-Site": "same-site"}}
	for _, route := range routes {
		for name, header := range attacks {
			got := r.do(t, req{Method: route.method, Path: route.path, Body: route.body, Paired: true, Header: header})
			if got.Status != http.StatusForbidden || got.errorCode(t) != "not_admitted" {
				t.Errorf("%s %s with %s: %d %s", route.method, route.path, name, got.Status, got.Body)
			}
		}
	}
	if calls := r.service.called(); len(calls) != 0 {
		t.Fatalf("a refused request reached the service: %v", calls)
	}
}

func TestSameOriginAndNavigationFetchesAreAdmitted(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for _, header := range []map[string]string{{"Origin": "http://" + r.bound, "Sec-Fetch-Site": "same-origin"}, {"Sec-Fetch-Site": "none"}, nil} {
		if got := r.do(t, req{Path: "/api/workspace", Paired: true, Header: header}); got.Status != http.StatusOK {
			t.Errorf("headers %v: %d %s", header, got.Status, got.Body)
		}
	}
}

func TestEveryAPIRouteButSessionNeedsThePairedSession(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for _, route := range apiRoutes {
		got := r.do(t, req{Method: route.method, Path: route.path, Body: route.body})
		if got.Status != http.StatusUnauthorized || got.errorCode(t) != "unpaired" {
			t.Errorf("%s %s unpaired: %d %s", route.method, route.path, got.Status, got.Body)
		}
		wrong := r.do(t, req{Method: route.method, Path: route.path, Body: route.body, Header: map[string]string{"Cookie": r.cookieName() + "=guess"}})
		if wrong.Status != http.StatusUnauthorized {
			t.Errorf("%s %s with a guessed session: %d", route.method, route.path, wrong.Status)
		}
	}
	if calls := r.service.called(); len(calls) != 0 {
		t.Fatalf("an unpaired request reached the service: %v", calls)
	}
	if got := r.do(t, req{Path: "/api/session"}); got.Status != http.StatusOK || !strings.Contains(got.Body, `"paired":false`) {
		t.Fatalf("session status unpaired: %d %s", got.Status, got.Body)
	}
}

func TestAnUnknownAPIRouteIsAdmittedThenNotFoundNeverTheUI(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for _, route := range []route{{http.MethodGet, "/api/nothing", ""}, {http.MethodGet, "/api/actions/world.create", ""}} {
		got := r.do(t, req{Method: route.method, Path: route.path, Paired: true})
		if got.Status != http.StatusNotFound || got.errorCode(t) != "not_found" || strings.Contains(got.Body, "<html>") {
			t.Errorf("%s %s: %d %s", route.method, route.path, got.Status, got.Body)
		}
	}
	if got := r.do(t, req{Path: "/api/nothing", Host: "evil.example"}); got.Status != http.StatusForbidden {
		t.Errorf("an unknown route skipped admission: %d", got.Status)
	}
	for _, route := range []route{{http.MethodDelete, "/api/file", ""}, {http.MethodPost, "/api/workspace", ""}} {
		if got := r.do(t, req{Method: route.method, Path: route.path, Paired: true}); got.Status != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: %d", route.method, route.path, got.Status)
		}
	}
}

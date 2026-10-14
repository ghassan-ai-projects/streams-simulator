package transport

import (
	"net/http"
	"testing"
)

func TestTwoBridgesOnOneHostKeepTheirOwnSessionCookies(t *testing.T) {
	t.Parallel()
	first, second := newRig(t), newRig(t)
	if first.cookieName() == second.cookieName() {
		t.Fatalf("both bridges name their cookie %q", first.cookieName())
	}
	paired := first.do(t, req{Method: http.MethodPost, Path: "/api/session", Body: `{"code":"ABCD-EFGH-JKLM"}`})
	if len(paired.Cookies) != 1 || paired.Cookies[0].Name != first.cookieName() {
		t.Fatalf("cookies = %v", paired.Cookies)
	}
	header := map[string]string{"Cookie": first.cookieName() + "=" + goodSession}
	if got := second.do(t, req{Path: "/api/workspace", Header: header}); got.Status != http.StatusUnauthorized {
		t.Fatalf("the other bridge accepted the first bridge's session: %d", got.Status)
	}
	if got := first.do(t, req{Path: "/api/workspace", Header: header}); got.Status != http.StatusOK {
		t.Fatalf("the first bridge refused its own session: %d", got.Status)
	}
}

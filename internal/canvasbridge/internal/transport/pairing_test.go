package transport

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestPairingSetsAnHttpOnlyStrictSessionCookie(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	got := r.do(t, req{Method: http.MethodPost, Path: "/api/session", Body: `{"code":"ABCD-EFGH-JKLM"}`})
	if got.Status != http.StatusOK || strings.TrimSpace(got.Body) != `{"paired":true}` {
		t.Fatalf("pairing: %d %s", got.Status, got.Body)
	}
	if len(got.Cookies) != 1 {
		t.Fatalf("cookies = %v", got.Cookies)
	}
	cookie := got.Cookies[0]
	if cookie.Name != cookieName || cookie.Value != goodSession || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode ||
		cookie.Path != "/" || cookie.Domain != "" || cookie.MaxAge != 12*60*60 {
		t.Fatalf("session cookie = %+v", cookie)
	}
}

func TestTheCodeAndTheSessionNeverAppearInAResponseBody(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	code := "SECRETCODE23"
	r.service.pairErr = domain.ErrWrongCode
	wrong := r.do(t, req{Method: http.MethodPost, Path: "/api/session", Body: `{"code":"` + code + `"}`})
	if wrong.Status != http.StatusUnauthorized || wrong.errorCode(t) != "wrong_code" || len(wrong.Cookies) != 0 {
		t.Fatalf("a wrong code: %d %s %v", wrong.Status, wrong.Body, wrong.Cookies)
	}
	r.service.pairErr = nil
	paired := r.do(t, req{Method: http.MethodPost, Path: "/api/session", Body: `{"code":"` + code + `"}`})
	status := r.do(t, req{Path: "/api/session", Paired: true})
	for name, body := range map[string]string{"wrong code": wrong.Body, "pairing": paired.Body, "status": status.Body} {
		if strings.Contains(body, code) || strings.Contains(body, goodSession) {
			t.Errorf("%s response leaks a secret: %s", name, body)
		}
	}
	if !strings.Contains(status.Body, `"paired":true`) {
		t.Fatalf("a paired session reports %s", status.Body)
	}
}

func TestAPIPairingFailuresMapToUnauthorized(t *testing.T) {
	t.Parallel()
	for err, code := range map[error]string{domain.ErrPairingLocked: "pairing_locked", domain.ErrPairingUsed: "pairing_used"} {
		r := newRig(t)
		r.service.pairErr = err
		got := r.do(t, req{Method: http.MethodPost, Path: "/api/session", Body: `{"code":"X"}`})
		if got.Status != http.StatusUnauthorized || got.errorCode(t) != code {
			t.Errorf("%v: %d %s", err, got.Status, got.Body)
		}
	}
	r := newRig(t)
	r.service.pairErr = errors.New("entropy exhausted")
	if got := r.do(t, req{Method: http.MethodPost, Path: "/api/session", Body: `{"code":"X"}`}); got.Status != http.StatusInternalServerError || len(got.Cookies) != 0 {
		t.Fatalf("a bridge failure while pairing: %d %v", got.Status, got.Cookies)
	}
}

func TestTheSessionIsReadFromAmongOtherCookies(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	header := map[string]string{"Cookie": "theme=dark; " + cookieName + "=" + goodSession + "; other=1"}
	if got := r.do(t, req{Path: "/api/workspace", Header: header}); got.Status != http.StatusOK {
		t.Fatalf("%d %s", got.Status, got.Body)
	}
}

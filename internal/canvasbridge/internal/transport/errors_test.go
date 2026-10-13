package transport

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestEachDomainErrorMapsToItsStatusAndCode(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		err    error
		status int
	}{
		"not_admitted":   {domain.ErrNotAdmitted, http.StatusForbidden},
		"unpaired":       {domain.ErrUnpaired, http.StatusUnauthorized},
		"wrong_code":     {domain.ErrWrongCode, http.StatusUnauthorized},
		"pairing_locked": {domain.ErrPairingLocked, http.StatusUnauthorized},
		"pairing_used":   {domain.ErrPairingUsed, http.StatusUnauthorized},
		"bad_path":       {domain.ErrBadPath, http.StatusBadRequest},
		"bad_arguments":  {domain.ErrBadArguments, http.StatusBadRequest},
		"unknown_action": {domain.ErrUnknownAction, http.StatusNotFound},
		"not_found":      {domain.ErrAbsent, http.StatusNotFound},
		"read_only":      {domain.ErrReadOnly, http.StatusForbidden},
		"sealed":         {domain.ErrSealed, http.StatusForbidden},
		"conflict":       {domain.ErrConflict, http.StatusConflict},
		"too_large":      {domain.ErrTooLarge, http.StatusRequestEntityTooLarge},
	}
	for code, tc := range cases {
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			r := newRig(t)
			r.service.fileErr = fmt.Errorf("wrapped by the module: %w", tc.err)
			got := r.do(t, req{Path: "/api/file?path=x", Paired: true})
			if got.Status != tc.status || got.errorCode(t) != code || !strings.Contains(got.Body, "wrapped by the module") {
				t.Fatalf("%v: %d %s, want %d %s", tc.err, got.Status, got.Body, tc.status, code)
			}
		})
	}
}

func TestAToolOrCommandFailureKeepsItsCodeMessageAndIssues(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.fileErr = &domain.Failure{Code: "domain_invalid", Message: "bad", Issues: []byte(`[{"pointer":"/x"}]`)}
	got := r.do(t, req{Method: http.MethodPost, Path: "/api/actions/domain.validate", Body: `{"args":{}}`, Paired: true})
	want := `{"error":{"code":"domain_invalid","message":"bad","issues":[{"pointer":"/x"}]}}`
	if got.Status != http.StatusUnprocessableEntity || strings.TrimSpace(got.Body) != want {
		t.Fatalf("%d %s", got.Status, got.Body)
	}
}

func TestAnUnexpectedErrorIsABridgeFailure(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.fileErr = fmt.Errorf("disk on fire")
	got := r.do(t, req{Path: "/api/workspace", Paired: true})
	if got.Status != http.StatusInternalServerError || got.errorCode(t) != "bridge_failed" {
		t.Fatalf("%d %s", got.Status, got.Body)
	}
}

func TestAConflictedWriteCarriesTheCurrentEtag(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.fileErr, r.service.etag = fmt.Errorf("%w: x", domain.ErrConflict), "current-etag"
	got := r.do(t, req{Method: http.MethodPut, Path: "/api/file", Paired: true, Body: `{"path":"x","content":"c","etag":"stale"}`})
	if got.Status != http.StatusConflict || got.errorCode(t) != "conflict" || !strings.Contains(got.Body, `"etag":"current-etag"`) {
		t.Fatalf("%d %s", got.Status, got.Body)
	}
	r.service.mu.Lock()
	written := r.service.written
	r.service.mu.Unlock()
	if written != [3]string{"x", "c", "stale"} {
		t.Fatalf("the service was asked to write %v", written)
	}
}

func TestASuccessfulWriteAndReadReturnTheirDocuments(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	write := r.do(t, req{Method: http.MethodPut, Path: "/api/file", Paired: true, Body: `{"path":"x","content":"c","etag":""}`})
	read := r.do(t, req{Path: "/api/file?path=x", Paired: true})
	listing := r.do(t, req{Path: "/api/workspace", Paired: true})
	actions := r.do(t, req{Path: "/api/actions", Paired: true})
	if strings.TrimSpace(write.Body) != `{"etag":"e2","path":"x"}` || !strings.Contains(read.Body, `"etag":"e1"`) ||
		!strings.Contains(listing.Body, `"name":"ws"`) || !strings.Contains(actions.Body, `"tool":"sim.world.create"`) {
		t.Fatalf("write %s, read %s, listing %s, actions %s", write.Body, read.Body, listing.Body, actions.Body)
	}
}

func TestAnActionRefusedForAStaleVersionCarriesTheCurrentEtag(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.actionErr = fmt.Errorf("wrapped: %w", &domain.ConflictError{Path: "x", Etag: "current-etag"})
	got := r.do(t, req{Method: http.MethodPost, Path: "/api/actions/domain.compose", Paired: true, Body: `{"args":{}}`})
	if got.Status != http.StatusConflict || got.errorCode(t) != "conflict" || !strings.Contains(got.Body, `"etag":"current-etag"`) {
		t.Fatalf("%d %s", got.Status, got.Body)
	}
}

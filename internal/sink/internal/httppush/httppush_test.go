package httppush

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestHTTPPushEquivalence(t *testing.T) {
	var mu sync.Mutex
	var received []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		mu.Lock()
		received = append(received, string(buf))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	lines := [][]byte{[]byte(`{"s":0}`), []byte(`{"s":1}`), []byte(`{"s":2}`)}
	h := New(t.Context(), srv.URL)
	for _, l := range lines {
		if err := h.Write(l); err != nil {
			t.Fatal(err)
		}
	}
	got, err := h.Close()
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("{\"s\":0}\n{\"s\":1}\n{\"s\":2}\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("http-push stream wrong:\n%q\n%q", got, want)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Fatalf("server received %d posts, want 3", len(received))
	}
	if received[0] != "{\"s\":0}" || received[2] != "{\"s\":2}" {
		t.Fatalf("delivery order wrong: %v", received)
	}
}

func TestHTTPPushUnreachableFails(t *testing.T) {
	// A configured endpoint that is down must fail the write, never silently
	// drop: a dropped record is indistinguishable from a modeled dropout.
	h := New(t.Context(), "http://127.0.0.1:1/unreachable")
	if err := h.Write([]byte(`{"x":1}`)); err == nil {
		t.Fatal("unreachable endpoint must fail the write")
	}
}

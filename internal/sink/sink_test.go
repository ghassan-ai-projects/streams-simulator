package sink_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
)

func TestInprocAndFileEqual(t *testing.T) {
	lines := [][]byte{[]byte(`{"a":1}`), []byte(`{"a":2}`), []byte(`{"a":3}`)}
	in := &sink.Inproc{}
	for _, l := range lines {
		if err := in.Write(l); err != nil {
			t.Fatal(err)
		}
	}
	inBytes, err := in.Close()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	f, err := sink.NewFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		if err := f.Write(l); err != nil {
			t.Fatal(err)
		}
	}
	fileBytes, err := f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(inBytes, fileBytes) {
		t.Fatalf("inproc/file divergence:\n%q\n%q", inBytes, fileBytes)
	}
	// File exists on disk with the same content.
	disk, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(disk, fileBytes) {
		t.Fatal("file sink did not persist its output")
	}
}

func TestFileSinkCreatesMissingDirectoriesAndFlushesBeforeClose(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "nested", "deeper", "trace.jsonl")
	f, err := sink.NewFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Write([]byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := f.Flush(); err != nil {
		t.Fatal(err)
	}
	if disk, err := os.ReadFile(path); err != nil || string(disk) != "{\"a\":1}\n" {
		t.Fatalf("flushed file = %q (%v)", disk, err)
	}
	if _, err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestInprocSinkIsReadyAtItsZeroValueAndSatisfiesTheContract(t *testing.T) {
	t.Parallel()
	var contract sink.Sink = &sink.Inproc{}
	if err := contract.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if got, err := contract.Close(); err != nil || string(got) != "x\n" {
		t.Fatalf("close = %q (%v)", got, err)
	}
}

func TestHTTPPushDeliversEachLineInOrderAndKeepsTheStream(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var posted []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buffer)
		mu.Lock()
		posted = append(posted, string(buffer))
		mu.Unlock()
	}))
	defer server.Close()
	push := sink.NewHTTPPush(t.Context(), server.URL)
	for _, line := range []string{"one", "two"} {
		if err := push.Write([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}
	got, err := push.Close()
	mu.Lock()
	defer mu.Unlock()
	if err != nil || string(got) != "one\ntwo\n" || len(posted) != 2 || posted[0] != "one" {
		t.Fatalf("stream = %q posted = %v (%v)", got, posted, err)
	}
}

func TestZeroValueFileAndHTTPPushRefuseInsteadOfPanicking(t *testing.T) {
	t.Parallel()
	var file sink.File
	if err := file.Write([]byte("x")); !errors.Is(err, sink.ErrNotConstructed) {
		t.Fatalf("file: %v", err)
	}
	var push sink.HTTPPush
	if err := push.Write([]byte("x")); !errors.Is(err, sink.ErrNotConstructed) {
		t.Fatalf("http push: %v", err)
	}
}

package acceptance

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const bridgeTimeout = 60 * time.Second

var pairingLine = regexp.MustCompile(`pairing code ([A-Z2-7-]+)`)

// canvasBridge is a running `streamsim canvas` process: where it listens and
// the pairing code it printed on stderr.
type canvasBridge struct {
	url, code string
	stderr    *lockedBuffer
}

type lockedBuffer struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, _ := b.data.Write(p)
	return n, nil
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data.String()
}

// startCanvasBridge runs the real binary on a loopback port it picks itself
// and returns once the address and the pairing code were printed.
func startCanvasBridge(t *testing.T, workspace, uiDir string) *canvasBridge {
	t.Helper()
	// #nosec G204 -- the test runs the freshly built binary with fixed arguments
	cmd := exec.CommandContext(context.WithoutCancel(t.Context()), buildBinary(t, testsupport.RepositoryRoot(), ""), "canvas", "--workspace", workspace,
		"--ui-dir", uiDir, "--addr", "127.0.0.1:0", "--adapters-dir", testsupport.AdaptersDir())
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = writer
	if err := cmd.Start(); err != nil {
		t.Fatalf("start the canvas bridge: %v", err)
	}
	_ = writer.Close()
	t.Cleanup(func() { stopProcess(cmd) })
	bridge := &canvasBridge{stderr: &lockedBuffer{}}
	announced := make(chan struct{})
	go bridge.scan(reader, announced)
	select {
	case <-announced:
	case <-time.After(bridgeTimeout):
		t.Fatalf("the bridge did not announce itself:\n%s", bridge.stderr.String())
	}
	return bridge
}

func (b *canvasBridge) scan(reader *os.File, announced chan<- struct{}) {
	defer func() { _ = reader.Close() }()
	scanner := bufio.NewScanner(reader)
	once := sync.Once{}
	for scanner.Scan() {
		line := scanner.Text()
		_, _ = b.stderr.Write([]byte(line + "\n"))
		if _, address, found := strings.Cut(line, "canvas: http://"); found {
			b.url = "http://" + address
		}
		if match := pairingLine.FindStringSubmatch(line); match != nil {
			b.code = match[1]
			once.Do(func() { close(announced) })
		}
	}
}

func stopProcess(cmd *exec.Cmd) {
	_ = cmd.Process.Signal(os.Interrupt)
	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(exited)
	}()
	select {
	case <-exited:
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		<-exited
	}
}

// bridgeClient is the canvas's side of the conversation: one cookie jar and a
// transcript of every byte the bridge answered with.
type bridgeClient struct {
	t          *testing.T
	base       string
	http       *http.Client
	mu         sync.Mutex
	transcript strings.Builder
}

func newBridgeClient(t *testing.T, base string) *bridgeClient {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &bridgeClient{t: t, base: base, http: &http.Client{Jar: jar}}
}

func (c *bridgeClient) record(raw []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transcript.Write(raw)
	c.transcript.WriteByte('\n')
}

func (c *bridgeClient) heard() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.transcript.String()
}

func (c *bridgeClient) do(ctx context.Context, method, path string, body any, header map[string]string) (*http.Response, error) {
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		payload = bytes.NewReader(raw)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.base+path, payload)
	if err != nil {
		c.t.Fatal(err)
	}
	for key, value := range header {
		request.Header.Set(key, value)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	return response, nil
}

// call sends one request and returns the status and the decoded JSON body.
func (c *bridgeClient) call(method, path string, body any) (int, map[string]any, http.Header) {
	c.t.Helper()
	ctx, cancel := context.WithTimeout(c.t.Context(), bridgeTimeout)
	defer cancel()
	response, err := c.do(ctx, method, path, body, nil)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	c.record(raw)
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil && len(bytes.TrimSpace(raw)) > 0 {
		c.t.Fatalf("%s %s answered %d with a body that is not a JSON object: %q", method, path, response.StatusCode, raw)
	}
	return response.StatusCode, document, response.Header
}

// action runs a registered canvas action that must succeed.
func (c *bridgeClient) action(name string, args map[string]any) map[string]any {
	c.t.Helper()
	status, document, _ := c.call(http.MethodPost, "/api/actions/"+name, map[string]any{"args": args})
	result, _ := document["result"].(map[string]any)
	if status != http.StatusOK || result == nil {
		c.t.Fatalf("action %s %v = %d %v", name, args, status, document)
	}
	return result
}

func errorCode(document map[string]any) string {
	failure, _ := document["error"].(map[string]any)
	code, _ := failure["code"].(string)
	return code
}

func (c *bridgeClient) page(path string) string {
	c.t.Helper()
	response, err := c.do(c.t.Context(), http.MethodGet, path, nil, nil)
	if err != nil {
		c.t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	return string(raw)
}

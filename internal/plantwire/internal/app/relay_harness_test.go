package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

type relayRun struct {
	stats domain.RelayStats
	err   error
}

type startedRelay struct {
	gateway string
	done    chan relayRun
	cancel  context.CancelFunc
}

func (s startedRelay) stop() relayRun {
	s.cancel()
	return <-s.done
}

func startRelay(t *testing.T, hostPath, out string, ticks <-chan time.Time) startedRelay {
	t.Helper()
	gatewayPath := filepath.Join(filepath.Dir(hostPath), "g.sock")
	ctx, cancel := context.WithCancel(t.Context())
	ready, done := make(chan struct{}), make(chan relayRun, 1)
	go func() {
		stats, err := serveRelay(ctx, RelayOptions{GatewaySocket: gatewayPath, HostSocket: hostPath, Out: out, Ready: func() { close(ready) }}, ticks)
		done <- relayRun{stats: stats, err: err}
	}()
	select {
	case <-ready:
	case ended := <-done:
		cancel()
		t.Fatalf("the relay did not start: %v", ended.err)
	}
	return startedRelay{gateway: gatewayPath, done: done, cancel: cancel}
}

type startedHost struct {
	done   chan host.Report
	cancel context.CancelFunc
}

func (s startedHost) stop() host.Report {
	s.cancel()
	return <-s.done
}

func startHost(t *testing.T, path string, options host.Options) startedHost {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	ready, done := make(chan struct{}), make(chan host.Report, 1)
	options.Ready = func() { close(ready) }
	go func() {
		report, _ := ServeHost(ctx, path, 0, options)
		done <- report
	}()
	<-ready
	return startedHost{done: done, cancel: cancel}
}

func tick(t *testing.T, ticks chan<- time.Time) {
	t.Helper()
	select {
	case ticks <- time.Time{}:
	case <-t.Context().Done():
		t.Fatal("the relay took no tick")
	}
}

type gatewayDriver struct {
	t    *testing.T
	node *standin.Node
	link *uds.Link
}

func dialGateway(t *testing.T, path string, node *standin.Node) *gatewayDriver {
	t.Helper()
	link, err := uds.Dial(t.Context(), path, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = link.Close() })
	return &gatewayDriver{t: t, node: node, link: link}
}

func (d *gatewayDriver) send(records ...map[string]any) {
	d.t.Helper()
	if err := d.link.Send(records); err != nil {
		d.t.Fatal(err)
	}
}

func ofKind(kind string) func(domain.Message) bool {
	return func(m domain.Message) bool { return m.Type == kind }
}

func (d *gatewayDriver) await(until func(domain.Message) bool) []domain.Inbound {
	d.t.Helper()
	inbound, met, err := d.link.Await(d.t.Context(), until)
	if err != nil || !met {
		d.t.Fatalf("the relay never sent what the gateway waited for (met %v): %v", met, err)
	}
	return inbound
}

func (d *gatewayDriver) handshake() domain.Inbound {
	d.t.Helper()
	birth, err := d.node.Birth()
	if err != nil {
		d.t.Fatal(err)
	}
	d.send(birth)
	inbound, met, err := d.link.AwaitHandshake(d.t.Context(), ofKind("STATE"))
	if err != nil || !met {
		d.t.Fatalf("no STATE answered the birth (met %v): %v", met, err)
	}
	state := inbound[len(inbound)-1]
	if state.Online {
		d.accept(state)
	}
	return state
}

func (d *gatewayDriver) accept(state domain.Inbound) {
	d.t.Helper()
	births, err := d.node.Accept(state.Record)
	if err != nil {
		d.t.Fatal(err)
	}
	d.send(births...)
}

func (d *gatewayDriver) sync() []domain.Inbound {
	d.t.Helper()
	d.send(d.node.Emit(domain.Sync(0, 1767225660000000000))...)
	inbound := d.await(ofKind("HSYNC"))
	for _, in := range inbound {
		if in.Type == "DCMD" {
			d.send(d.node.Handle(in.Record)...)
		}
	}
	return inbound
}

func lastState(t *testing.T, inbound []domain.Inbound) domain.Inbound {
	t.Helper()
	for i := len(inbound) - 1; i >= 0; i-- {
		if inbound[i].Type == "STATE" && inbound[i].Online {
			return inbound[i]
		}
	}
	t.Fatalf("no online STATE among %d records", len(inbound))
	return domain.Inbound{}
}

func wireLines(t *testing.T, out string) []map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, "journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var lines []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
		var parsed map[string]any
		if err := json.Unmarshal(line, &parsed); err != nil {
			t.Fatalf("journal line %q: %v", line, err)
		}
		if parsed["type"] == "wire" {
			lines = append(lines, parsed)
		}
	}
	return lines
}

func recordOf(line map[string]any) map[string]any {
	record, _ := line["record"].(map[string]any)
	return record
}

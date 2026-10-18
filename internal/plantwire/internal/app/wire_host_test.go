package app

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"slices"
	"sync"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

type wireHost struct {
	t        *testing.T
	listener net.Listener
	mu       sync.Mutex
	conn     net.Conn
	tx       *domain.Transmitter
	received []map[string]any
	held     bool
	pending  []uint64
	changed  chan struct{}
}

func serveWireHost(t *testing.T, path string, hold bool) *wireHost {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	h := &wireHost{t: t, listener: listener, held: hold, changed: make(chan struct{}, 1)}
	t.Cleanup(func() { _ = listener.Close() })
	go h.serve()
	return h
}

func (h *wireHost) serve() {
	conn, err := h.listener.Accept()
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()
	h.mu.Lock()
	h.conn = conn
	h.mu.Unlock()
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		if in, ok := domain.Decode(scanner.Bytes()); ok {
			h.handle(in)
		}
	}
}

func (h *wireHost) handle(in domain.Inbound) {
	h.mu.Lock()
	h.received = append(h.received, in.Record)
	first := h.tx == nil && in.Type == "NBIRTH"
	if first {
		h.tx = domain.NewTransmitter(in.SessionID)
	}
	hold := in.Type == "NSYNC" && h.held
	if hold {
		h.pending = append(h.pending, in.Seq)
	}
	h.mu.Unlock()
	select {
	case h.changed <- struct{}{}:
	default:
	}
	switch {
	case first:
		h.write(domain.Record{Type: "STATE", Lane: domain.LaneControl, Fields: map[string]any{"host_id": "wire-host", "online": true, "heartbeat": 0,
			"heartbeat_interval_ms": 1000, "authority_epoch": 7, "protocol_min": 1, "protocol_max": 1}})
	case in.Type == "NSYNC" && !hold:
		h.write(hsyncRecord(in.Seq))
	}
}

func hsyncRecord(seq uint64) domain.Record {
	return domain.Record{Type: "HSYNC", Lane: domain.LaneControl, Fields: map[string]any{"through_seq": seq}}
}

func (h *wireHost) write(records ...domain.Record) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, record := range records {
		h.tx.Queue(record)
	}
	for _, sent := range h.tx.Drain() {
		line, err := json.Marshal(sent)
		if err == nil {
			_, err = h.conn.Write(append(line, '\n'))
		}
		if err != nil {
			return
		}
	}
}

func (h *wireHost) release() {
	h.mu.Lock()
	h.held = false
	pending := h.pending
	h.pending = nil
	h.mu.Unlock()
	for _, seq := range pending {
		h.write(hsyncRecord(seq))
	}
}

func (h *wireHost) snapshot() []map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.received)
}

func (h *wireHost) waitFor(until func(received []map[string]any) bool) {
	h.t.Helper()
	for !until(h.snapshot()) {
		select {
		case <-h.changed:
		case <-h.t.Context().Done():
			h.t.Fatal("the host never received what the test waited for")
		}
	}
}

func hasRecord(received []map[string]any, kind, member string, want any) bool {
	return slices.ContainsFunc(received, func(record map[string]any) bool {
		return record["type"] == kind && (member == "" || record[member] == want)
	})
}

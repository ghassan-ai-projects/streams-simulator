package acceptance

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// serverEvent is one Server-Sent Event of the bridge.
type serverEvent struct {
	id, name, data string
}

func (e serverEvent) recordType() string {
	var head struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal([]byte(e.data), &head)
	return head.Type
}

// eventStream follows one Server-Sent Events route until it is closed.
type eventStream struct {
	t      *testing.T
	events chan serverEvent
	cancel context.CancelFunc
}

func (c *bridgeClient) follow(path, lastEventID string) *eventStream {
	c.t.Helper()
	ctx, cancel := context.WithTimeout(c.t.Context(), bridgeTimeout)
	header := map[string]string{}
	if lastEventID != "" {
		header["Last-Event-ID"] = lastEventID
	}
	response, err := c.do(ctx, http.MethodGet, path, nil, header)
	if err != nil || response.StatusCode != http.StatusOK {
		cancel()
		c.t.Fatalf("follow %s: %v (%v)", path, err, response)
	}
	stream := &eventStream{t: c.t, events: make(chan serverEvent, 4096), cancel: cancel}
	c.t.Cleanup(cancel)
	go stream.read(response, c)
	return stream
}

func (s *eventStream) read(response *http.Response, c *bridgeClient) {
	defer close(s.events)
	defer func() { _ = response.Body.Close() }()
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 0, 64<<10), 8<<20)
	var current serverEvent
	for scanner.Scan() {
		line := scanner.Text()
		c.record([]byte(line))
		switch {
		case line == "":
			s.events <- current
			current = serverEvent{}
		case strings.HasPrefix(line, "id: "):
			current.id = strings.TrimPrefix(line, "id: ")
		case strings.HasPrefix(line, "event: "):
			current.name = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			current.data = strings.TrimPrefix(line, "data: ")
		}
	}
}

// next waits for the following event; ok is false once the stream ended.
func (s *eventStream) next() (serverEvent, bool) {
	s.t.Helper()
	select {
	case event, ok := <-s.events:
		return event, ok
	case <-s.t.Context().Done():
		s.t.Fatal("the test ended while waiting for an event")
	}
	return serverEvent{}, false
}

// until reads events until one satisfies stop, returning everything read.
func (s *eventStream) until(stop func(serverEvent) bool) []serverEvent {
	s.t.Helper()
	var read []serverEvent
	for {
		event, ok := s.next()
		if !ok {
			s.t.Fatalf("the stream ended before the expected event; read %d events", len(read))
		}
		read = append(read, event)
		if stop(event) {
			return read
		}
	}
}

// drain reads to the end of the stream.
func (s *eventStream) drain() []serverEvent {
	s.t.Helper()
	var read []serverEvent
	for event, ok := s.next(); ok; event, ok = s.next() {
		read = append(read, event)
	}
	return read
}

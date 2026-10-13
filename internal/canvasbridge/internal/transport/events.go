package transport

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

type eventStream struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func openStream(w http.ResponseWriter) (*eventStream, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	return &eventStream{w: w, flusher: flusher}, true
}

func (e *eventStream) send(id, event string, data []byte) error {
	if id != "" {
		if _, err := fmt.Fprintf(e.w, "id: %s\n", id); err != nil {
			return fmt.Errorf("canvasbridge: stream: %w", err)
		}
	}
	if _, err := fmt.Fprintf(e.w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return fmt.Errorf("canvasbridge: stream: %w", err)
	}
	e.flusher.Flush()
	return nil
}

func (s *Server) followRun(w http.ResponseWriter, r *http.Request) {
	if err := s.service.OpenRun(r.PathValue("world")); err != nil {
		writeError(w, err)
		return
	}
	stream, ok := openStream(w)
	if !ok {
		return
	}
	if err := s.service.Follow(r.Context(), r.PathValue("world"), lastEventID(r), stream.record); err != nil {
		_ = stream.send("", "failure", []byte(strconv.Quote(err.Error())))
	}
}

func (e *eventStream) record(record domain.Record) error {
	if !record.Seen {
		if err := e.send(strconv.FormatInt(record.Index, 10), "record", record.Line); err != nil {
			return err
		}
	}
	if !record.Closes {
		return nil
	}
	return e.send("", "closed", []byte("{}"))
}

func (s *Server) watch(w http.ResponseWriter, r *http.Request) {
	stream, ok := openStream(w)
	if !ok {
		return
	}
	_ = s.service.Watch(r.Context(), func() error {
		return stream.send("", "workspace", []byte("{}"))
	})
}

func lastEventID(r *http.Request) int64 {
	id, err := strconv.ParseInt(r.Header.Get("Last-Event-ID"), 10, 64)
	if err != nil {
		return -1
	}
	return id
}

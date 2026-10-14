// Package transport is the HTTP edge of the canvas bridge: the loopback
// listener, the JSON routes, Server-Sent Events and the built UI, each
// response with its security headers. Every /api route but pairing is
// admitted (host, origin, fetch site) and needs the paired session.
package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const (
	maxFileBody     = domain.MaxFileBytes + 1<<20
	maxActionBody   = 1 << 20
	sessionMaxAge   = 12 * 60 * 60
	contentSecurity = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; " +
		"connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'"
)

type Service interface {
	Pair(code string) (string, error)
	Admits(session string) bool
	Workspace() (domain.Listing, error)
	ReadFile(rel string) (domain.FileDocument, error)
	WriteFile(rel, content, etag string) (domain.FileDocument, error)
	Actions() []domain.Action
	RunAction(ctx context.Context, name string, args map[string]any) (json.RawMessage, error)
	OpenRun(worldID string) error
	Follow(ctx context.Context, worldID string, after int64, emit func(domain.Record) error) error
	Watch(ctx context.Context, changed func() error) error
}

type Server struct {
	service Service
	bound   string
	ui      *os.Root
}

func Listen(ctx context.Context, addr string) (net.Listener, error) {
	if err := domain.AdmitListenAddress(addr); err != nil {
		return nil, err
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, errors.Join(domain.ErrNonLoopback, err)
	}
	return listener, nil
}

func Serve(ctx context.Context, listener net.Listener, service Service, uiDir string) error {
	ui, err := os.OpenRoot(uiDir)
	if err != nil {
		return errors.Join(errors.New("canvasbridge: open the UI directory"), err)
	}
	defer func() { _ = ui.Close() }()
	server := &Server{service: service, bound: listener.Addr().String(), ui: ui}
	return server.run(ctx, listener)
}

func (s *Server) run(ctx context.Context, listener net.Listener) error {
	httpServer := &http.Server{Handler: s.routes(), ReadHeaderTimeout: 10 * time.Second, BaseContext: func(net.Listener) context.Context { return ctx }}
	go func() {
		<-ctx.Done()
		_ = httpServer.Close()
	}()
	if err := httpServer.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return errors.Join(errors.New("canvasbridge: serve"), err)
	}
	return nil
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/session", s.admitted(s.sessionStatus, false))
	mux.HandleFunc("POST /api/session", s.admitted(s.pair, false))
	mux.HandleFunc("GET /api/workspace", s.admitted(s.workspace, true))
	mux.HandleFunc("GET /api/file", s.admitted(s.readFile, true))
	mux.HandleFunc("PUT /api/file", s.admitted(s.writeFile, true))
	mux.HandleFunc("GET /api/actions", s.admitted(s.actions, true))
	mux.HandleFunc("POST /api/actions/{name}", s.admitted(s.runAction, true))
	mux.HandleFunc("GET /api/runs/{world}/events", s.admitted(s.followRun, true))
	mux.HandleFunc("GET /api/events", s.admitted(s.watch, true))
	mux.HandleFunc("GET /api/", s.admitted(s.unknownRoute, false))
	mux.HandleFunc("GET /", s.static)
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Content-Security-Policy", contentSecurity)
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

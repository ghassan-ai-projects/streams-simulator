package transport

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (s *Server) workspace(w http.ResponseWriter, _ *http.Request) {
	listing, err := s.service.Workspace()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listing)
}

func (s *Server) readFile(w http.ResponseWriter, r *http.Request) {
	document, err := s.service.ReadFile(r.URL.Query().Get("path"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, document)
}

type fileWrite struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Etag    string `json:"etag"`
}

func (s *Server) writeFile(w http.ResponseWriter, r *http.Request) {
	var body fileWrite
	if !readBody(w, r, maxFileBody, &body) {
		return
	}
	document, err := s.service.WriteFile(body.Path, body.Content, body.Etag)
	if err != nil {
		writeConflict(w, err, document.Etag)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": document.Path, "etag": document.Etag})
}

func writeConflict(w http.ResponseWriter, err error, current string) {
	status, failure := classify(err)
	writeJSON(w, status, map[string]any{"error": failure, "etag": current})
}

func (s *Server) actions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"actions": s.service.Actions()})
}

func (s *Server) runAction(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Args map[string]any `json:"args"`
	}
	if !readBody(w, r, maxActionBody, &body) {
		return
	}
	result, err := s.service.RunAction(r.Context(), r.PathValue("name"), body.Args)
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}

func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if name == "" || !s.exists(name) {
		name = "index.html"
	}
	//nolint:gosec // G703: the name is cleaned and the file system is an os.Root of the UI folder
	http.ServeFileFS(w, r, s.ui.FS(), name)
}

func (s *Server) exists(name string) bool {
	info, err := s.ui.Stat(name)
	return err == nil && !info.IsDir()
}

func (s *Server) unknownRoute(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]any{"error": &domain.Failure{Code: "not_found", Message: "no such API route"}})
}

func writeActionError(w http.ResponseWriter, err error) {
	var conflict *domain.ConflictError
	if errors.As(err, &conflict) {
		writeConflict(w, err, conflict.Etag)
		return
	}
	writeError(w, err)
}

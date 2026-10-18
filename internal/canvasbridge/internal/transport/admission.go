package transport

import (
	"net/http"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (s *Server) admitted(handler http.HandlerFunc, needsSession bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := domain.Request{Host: r.Host, Origin: r.Header.Get("Origin"), SecFetchSite: r.Header.Get("Sec-Fetch-Site")}
		if err := domain.Admit(request, s.bound); err != nil {
			writeError(w, err)
			return
		}
		if needsSession && !s.service.Admits(domain.CookieValue(r.Header.Get("Cookie"), s.cookieName())) {
			writeError(w, domain.ErrUnpaired)
			return
		}
		handler(w, r)
	}
}

func (s *Server) cookieName() string { return domain.SessionCookieName(s.bound) }

func (s *Server) sessionStatus(w http.ResponseWriter, r *http.Request) {
	paired := s.service.Admits(domain.CookieValue(r.Header.Get("Cookie"), s.cookieName()))
	writeJSON(w, http.StatusOK, map[string]bool{"paired": paired})
}

func (s *Server) pair(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !readBody(w, r, maxActionBody, &body) {
		return
	}
	session, err := s.service.Pair(body.Code)
	if err != nil {
		writeError(w, err)
		return
	}
	http.SetCookie(w, sessionCookie(s.cookieName(), session))
	writeJSON(w, http.StatusOK, map[string]bool{"paired": true})
}

func sessionCookie(name, session string) *http.Cookie {
	//nolint:gosec // G124: the bridge serves plain HTTP on loopback only, where a Secure cookie is dropped by Safari and by Go's cookie jar (governance/DEFERRED.md D-63)
	return &http.Cookie{Name: name, Value: session, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: sessionMaxAge}
}

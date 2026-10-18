package app

import (
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

type liveSession struct {
	mu       sync.Mutex
	view     domain.SessionView
	lasting  Lasting
	stopping bool
	failed   *domain.SessionStatus
}

func (s *liveSession) status() domain.SessionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch {
	case s.failed != nil:
		return *s.failed
	case s.lasting == nil:
		return s.view.Status(domain.SessionStarting)
	}
	return s.attached()
}

func (s *liveSession) attached() domain.SessionStatus {
	if exit, exited := s.lasting.Exited(); exited {
		return s.view.Outcome(exit.Stdout, exit.ExitCode, exit.Stderr)
	}
	if s.stopping {
		return s.view.Status(domain.SessionStopping)
	}
	return s.view.Status(domain.SessionRunning)
}

func (s *liveSession) live() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failed != nil {
		return false
	}
	if s.lasting == nil {
		return true
	}
	_, exited := s.lasting.Exited()
	return !exited
}

func (s *liveSession) attach(lasting Lasting) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lasting = lasting
	return s.stopping
}

func (s *liveSession) fail(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	failed := s.view.Unlaunched(err)
	s.failed = &failed
}

func (s *liveSession) requestStop() Lasting {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopping || s.failed != nil {
		return nil
	}
	s.stopping = true
	return s.lasting
}

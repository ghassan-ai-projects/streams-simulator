// Package app holds the use cases of the canvas bridge: pair a browser, list
// the workspace, read and write its files, run a registered action through the
// director or a streamsim command, own the child of a connected session until
// it is stopped, follow a run's observer journal and watch the workspace for
// changes. It decides nothing itself: the rules are in the
// domain layer and every effect goes through an edge.
package app

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

type Director interface {
	Call(ctx context.Context, tool string, args map[string]any) (process.ToolResult, error)
	Close() error
}

type Edges struct {
	Workspace        *files.Workspace
	Tick             func(every time.Duration) (<-chan time.Time, func())
	StartDirector    func(ctx context.Context) (Director, error)
	RunCommand       func(ctx context.Context, argv []string, stdin []byte) (process.CommandResult, error)
	StartLasting     func(argv []string, stdin []byte) (Lasting, error)
	NewControlSocket func() (path string, remove func(), err error)
	ControlReady     func(path string) bool
	SendControl      func(ctx context.Context, path string, line []byte) ([]byte, error)
	NewSession       func() string
	NewRunID         func() string
}

type Settings struct {
	WorkspaceDir  string
	AdaptersDir   string
	ComponentDirs []string
	PairingCode   string
}

type Bridge struct {
	edges      Edges
	settings   Settings
	stateMu    sync.Mutex
	pairing    *domain.Pairing
	running    map[string]bool
	directorMu sync.Mutex
	director   Director
	catalog    string
	writeMu    sync.Mutex
	session    *liveSession
	sessions   map[string]json.RawMessage
}

func New(edges Edges, settings Settings) *Bridge {
	return &Bridge{edges: edges, settings: settings, pairing: domain.NewPairing(settings.PairingCode),
		running: map[string]bool{}, sessions: map[string]json.RawMessage{}}
}

func (b *Bridge) Pair(code string) (string, error) {
	session := b.edges.NewSession()
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	if err := b.pairing.Pair(code, session); err != nil {
		return "", err
	}
	return session, nil
}

func (b *Bridge) Admits(session string) bool {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	return b.pairing.Admits(session)
}

func (b *Bridge) Actions() []domain.Action {
	return domain.Actions()
}

func (b *Bridge) Close() error {
	b.stopLiving()
	b.directorMu.Lock()
	defer b.directorMu.Unlock()
	if b.director == nil {
		return nil
	}
	return b.director.Close()
}

func decode(raw []byte) (json.RawMessage, error) {
	if !json.Valid(raw) {
		return nil, &domain.Failure{Code: "command_failed", Message: "streamsim printed no JSON document"}
	}
	return json.RawMessage(raw), nil
}

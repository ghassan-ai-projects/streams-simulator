package process

import (
	"strings"
	"sync"
)

const tailLimit = 4096

type tail struct {
	mu   sync.Mutex
	text string
}

func (t *tail) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.text += string(p)
	if len(t.text) > tailLimit {
		t.text = t.text[len(t.text)-tailLimit:]
	}
	return len(p), nil
}

func (t *tail) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return strings.TrimSpace(t.text)
}

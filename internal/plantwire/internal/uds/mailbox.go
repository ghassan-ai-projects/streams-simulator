package uds

import "sync"

const mailboxLimit = 1 << 20

type Mailbox[T any] struct {
	mu     sync.Mutex
	items  []T
	shut   bool
	signal chan struct{}
}

func NewMailbox[T any]() *Mailbox[T] {
	return &Mailbox[T]{signal: make(chan struct{}, 1)}
}

func (m *Mailbox[T]) Put(items ...T) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || len(m.items)+len(items) > mailboxLimit {
		return false
	}
	m.items = append(m.items, items...)
	m.Wake()
	return true
}

func (m *Mailbox[T]) Pop() (T, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var none T
	if len(m.items) == 0 {
		return none, false
	}
	item := m.items[0]
	m.items[0], m.items = none, m.items[1:]
	return item, true
}

func (m *Mailbox[T]) Take() []T {
	m.mu.Lock()
	defer m.mu.Unlock()
	taken := m.items
	m.items = nil
	return taken
}

func (m *Mailbox[T]) Ready() <-chan struct{} { return m.signal }

func (m *Mailbox[T]) Shut() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shut = true
	m.Wake()
}

func (m *Mailbox[T]) IsShut() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.shut
}

func (m *Mailbox[T]) Wake() {
	select {
	case m.signal <- struct{}{}:
	default:
	}
}

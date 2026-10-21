package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/canonical"

const MaxDeadlineAheadNS = int64(24 * 3600 * 1e9)

type Command struct {
	ID, Key, Device, Name, Boot, Policy string
	Args                                map[string]any
	DeadlineNS, NotBeforeNS, Epoch      int64
}

func CommandOf(m Message) Command {
	return Command{ID: m.CommandID, Key: m.IdempotencyKey, Device: m.Device, Name: m.Command, Boot: m.ExpectedBootID,
		Policy: m.PolicyDigest, Args: m.Args, DeadlineNS: m.DeadlineNS, NotBeforeNS: m.NotBeforeNS, Epoch: m.AuthorityEpoch}
}

func (c Command) Content() string {
	digest, _ := canonical.Digest(map[string]any{"device": c.Device, "command": c.Name, "args": c.Args, "expected_boot_id": c.Boot})
	return digest
}

func (c Command) IntentKey(commandID string) string {
	digest, _ := canonical.Digest(map[string]any{"command_id": commandID, "content": c.Content()})
	return digest
}

func Freshness(c Command, nowNS int64) string {
	switch {
	case c.DeadlineNS < nowNS:
		return "expired"
	case c.DeadlineNS > nowNS+MaxDeadlineAheadNS:
		return "out_of_range"
	}
	return ""
}

type Answer struct {
	Status, Reason, Outcome string
	Silent                  bool
}

type Ledger map[string]*Entry

type Entry struct {
	content string
	Answer  Answer
	Pending bool
	Waiting []string
}

func (l Ledger) Recall(c Command) (*Entry, bool) {
	entry := l[c.Key]
	if entry != nil && entry.content != c.Content() {
		return nil, true
	}
	return entry, false
}

func (l Ledger) Hold(c Command) {
	l[c.Key] = &Entry{content: c.Content(), Pending: true, Waiting: []string{c.ID}}
}

func (l Ledger) Settle(c Command, answer Answer) []string {
	entry := l[c.Key]
	if entry == nil {
		entry = &Entry{content: c.Content(), Waiting: []string{c.ID}}
		l[c.Key] = entry
	}
	entry.Answer, entry.Pending = answer, false
	return entry.Waiting
}

func (l Ledger) Release(c Command) []string {
	waiting := l[c.Key].Waiting
	delete(l, c.Key)
	return waiting
}

func Result(commandID, outcome, bootID string) Record {
	return newRecord("DRESULT", LaneControl, map[string]any{"command_id": commandID, "reported_outcome": outcome, "boot_id": bootID})
}

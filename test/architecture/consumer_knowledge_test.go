package architecture

import (
	"regexp"
	"strings"
	"testing"
)

// consumerNames matches the names of the consumers this simulator is
// integrated with. The binary has no knowledge of its consumers (AGENTS.md
// Forbidden Changes): a consumer is data (an adapter file), never code.
var consumerNames = regexp.MustCompile(`(?i)agentic[ _-]?stream|tamoz|arduino`)

// consumerKnowledgeExceptions are the production paths allowed to name a
// consumer. The device emulator vendors the Agentic Stream device wire
// contract; the exception is time-boxed and ends when Agentic Stream moves to
// the plant wire (governance/CANVAS_QUALITY_BAR.md §1.7).
var consumerKnowledgeExceptions = []string{
	"internal/device/",
}

func TestNoProductionSourceNamesAConsumer(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	for _, file := range productionFiles(t) {
		if consumerKnowledgeAllowed(file.path) {
			continue
		}
		raw, err := root.ReadFile(file.path)
		if err != nil {
			t.Fatal(err)
		}
		if name := consumerNames.FindString(string(raw)); name != "" {
			t.Errorf("%s names the consumer %q; consumers are data, not code", file.path, name)
		}
	}
}

func TestEveryConsumerKnowledgeExceptionIsStillNeeded(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	for _, allowed := range consumerKnowledgeExceptions {
		needed := false
		for _, file := range productionFiles(t) {
			if !consumerKnowledgeAllowed(file.path) || !strings.HasPrefix(file.path, allowed) {
				continue
			}
			raw, err := root.ReadFile(file.path)
			if err != nil {
				t.Fatal(err)
			}
			needed = needed || consumerNames.Match(raw)
		}
		if !needed {
			t.Errorf("consumer-knowledge exception %q names no consumer any more (or no longer exists); remove it", allowed)
		}
	}
}

func TestConsumerNamePatternIsSpecific(t *testing.T) {
	t.Parallel()
	for text, want := range map[string]bool{
		"the Agentic Stream serial effector": true,
		"agentic-stream.adapter.json":        true,
		"a Tamoz episode":                    true,
		"an Arduino bench":                   true,
		"a stream of agentic records":        false,
		"streams simulator":                  false,
	} {
		if got := consumerNames.MatchString(text); got != want {
			t.Errorf("%q: matches=%v, want %v", text, got, want)
		}
	}
}

func consumerKnowledgeAllowed(path string) bool {
	for _, allowed := range consumerKnowledgeExceptions {
		if path == allowed || (strings.HasSuffix(allowed, "/") && strings.HasPrefix(path, allowed)) {
			return true
		}
	}
	return false
}

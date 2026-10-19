package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/canonical"

func (l *Live) SetInterrupts(view map[string]any) bool {
	digest, err := canonical.Digest(view)
	if err != nil || (l.interrupts != nil && l.interruptDigest == digest) {
		return false
	}
	l.interrupts, l.interruptDigest = view, digest
	return true
}

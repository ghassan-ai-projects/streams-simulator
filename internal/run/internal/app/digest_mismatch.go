package app

import "fmt"

type digestMismatch struct{ subject, recorded, current string }

func (e *digestMismatch) Error() string {
	return fmt.Sprintf("run: %s %s: artifact=%s current=%s", e.subject, ErrDigestMismatch, e.recorded, e.current)
}

func (e *digestMismatch) Unwrap() error { return ErrDigestMismatch }

func (e *digestMismatch) verdict() string {
	if e.subject == "world" {
		return fmt.Sprintf("the recorded specs replay to a different world than the recorded world digest (recorded %s, replayed %s)", e.recorded, e.current)
	}
	return fmt.Sprintf("the recorded %s spec does not match its digest (recorded %s, embedded spec is %s)", e.subject, e.recorded, e.current)
}

package domain

import "errors"

// ErrInvalidInvocation marks an effector invocation refused at admission:
// a missing command id, an entity that does not exist or is retired, an
// undeclared effector, arguments its args_schema refuses. Such a refusal
// happens before anything is recorded or drawn, so the world is exactly as
// it was without the call.
var ErrInvalidInvocation = errors.New("world: invalid invocation")

// invalidInvocation keeps the admission error's own message while marking it
// as ErrInvalidInvocation.
type invalidInvocation struct{ err error }

func (e invalidInvocation) Error() string { return e.err.Error() }

// Unwrap exposes both the mark and the admission error.
func (e invalidInvocation) Unwrap() []error { return []error{ErrInvalidInvocation, e.err} }

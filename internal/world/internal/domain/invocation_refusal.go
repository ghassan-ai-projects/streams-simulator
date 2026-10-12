package domain

import "errors"

var ErrInvalidInvocation = errors.New("world: invalid invocation")

type invalidInvocation struct{ err error }

func (e invalidInvocation) Error() string { return e.err.Error() }

func (e invalidInvocation) Unwrap() []error { return []error{ErrInvalidInvocation, e.err} }

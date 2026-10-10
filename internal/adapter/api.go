package adapter

import "errors"

// ErrNoAdapter is returned by NewEngine when no adapter is given.
var ErrNoAdapter = errors.New("adapter: an adapter is required")

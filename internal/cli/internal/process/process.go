// Package process is the operating-system edge of the cli module: the
// standard streams and the wall clock of the process a command runs in.
package process

import (
	"io"
	"os"
	"time"
)

// Env is what a command may reach of its process: where it writes results and
// diagnostics, and what time it is.
type Env struct {
	Stdout io.Writer
	Stderr io.Writer
	Now    func() time.Time
}

// Standard is the environment of the running process.
func Standard() Env {
	return Env{Stdout: os.Stdout, Stderr: os.Stderr, Now: time.Now}
}

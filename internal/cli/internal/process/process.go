// Package process is the operating-system edge of the cli module: the
// standard streams and the wall clock of the process a command runs in.
package process

import (
	"io"
	"os"
	"time"
)

type Env struct {
	Stdout io.Writer
	Stderr io.Writer
	Now    func() time.Time
}

func Standard() Env {
	return Env{Stdout: os.Stdout, Stderr: os.Stderr, Now: time.Now}
}

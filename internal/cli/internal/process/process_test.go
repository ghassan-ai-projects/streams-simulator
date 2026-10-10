package process

import (
	"os"
	"testing"
	"time"
)

func TestStandardIsTheRunningProcess(t *testing.T) {
	t.Parallel()
	env := Standard()
	if env.Stdout != os.Stdout || env.Stderr != os.Stderr {
		t.Fatal("standard streams must be the process streams")
	}
	if delta := time.Since(env.Now()); delta < 0 || delta > time.Minute {
		t.Fatalf("Now is %v away from the wall clock", delta)
	}
}

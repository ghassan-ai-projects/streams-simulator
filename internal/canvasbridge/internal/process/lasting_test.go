package process

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"testing"
	"time"
)

func lingerUntilInterrupted(args []string) int {
	interrupted := make(chan os.Signal, 1)
	if args[0] == "stubborn" {
		signal.Ignore(os.Interrupt)
	} else {
		signal.Notify(interrupted, os.Interrupt)
	}
	plan, _ := io.ReadAll(os.Stdin)
	writeOwnPID(args[1])
	select {
	case <-interrupted:
	case <-time.After(time.Hour):
	}
	_, _ = os.Stderr.WriteString("stopping\n")
	_, _ = os.Stdout.WriteString(`{"stopped":true,"plan":` + string(plan) + `}`)
	return 0
}

func startedLasting(t *testing.T, mode string) *Lasting {
	t.Helper()
	ready := filepath.Join(t.TempDir(), "ready")
	lasting, err := StartLasting(self(t), t.TempDir(), []string{"--helper", mode, ready}, []byte(`{"seed":1}`))
	if err != nil {
		t.Fatal(err)
	}
	ticks := time.NewTicker(10 * time.Millisecond)
	defer ticks.Stop()
	for attempt := 0; attempt < 1000; attempt++ {
		if _, err := os.Stat(ready); err == nil {
			return lasting
		}
		<-ticks.C
	}
	t.Fatal("the lasting child never became ready")
	return nil
}

func TestALastingChildRunsUntilInterruptedAndLeavesWhatItPrinted(t *testing.T) {
	t.Parallel()
	lasting := startedLasting(t, "linger")
	if _, exited := lasting.Exited(); exited {
		t.Fatal("the child exited before it was stopped")
	}
	exit := lasting.Stop(10 * time.Second)
	if exit.ExitCode != 0 || string(exit.Stdout) != `{"stopped":true,"plan":{"seed":1}}` || exit.Stderr != "stopping" {
		t.Fatalf("exit = %+v, want a clean exit with the document it printed, given the plan on stdin", exit)
	}
	if again, exited := lasting.Exited(); !exited || again.ExitCode != 0 {
		t.Fatalf("Exited after Stop = %+v, %v, want the same exit", again, exited)
	}
	if twice := lasting.Stop(time.Second); twice.ExitCode != 0 {
		t.Fatalf("a second Stop = %+v, want the first exit again", twice)
	}
}

func TestALastingChildThatIgnoresTheInterruptIsKilledAfterTheWait(t *testing.T) {
	t.Parallel()
	lasting := startedLasting(t, "stubborn")
	if exit := lasting.Stop(200 * time.Millisecond); exit.ExitCode != -1 || len(exit.Stdout) != 0 {
		t.Fatalf("exit = %+v, want the child killed", exit)
	}
}

func TestALastingChildThatCannotStartSaysSo(t *testing.T) {
	t.Parallel()
	if _, err := StartLasting(filepath.Join(t.TempDir(), "absent"), t.TempDir(), nil, nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("StartLasting of an absent executable = %v, want ErrNotExist", err)
	}
}

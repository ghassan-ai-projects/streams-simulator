package process

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const drainWait = 2 * time.Second

type lastWords struct {
	text    tail
	drained chan struct{}
}

func captureStderr(command *exec.Cmd) (*lastWords, func(), error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("canvasbridge: capture the director's stderr: %w", err)
	}
	words := &lastWords{drained: make(chan struct{})}
	command.Stderr = writer
	go words.collect(reader)
	return words, func() { _ = writer.Close() }, nil
}

func (w *lastWords) collect(reader *os.File) {
	defer close(w.drained)
	defer func() { _ = reader.Close() }()
	_, _ = io.Copy(&w.text, reader)
}

func connectionLost(err error) bool {
	for _, lost := range []error{mcpsdk.ErrConnectionClosed, io.EOF, io.ErrClosedPipe, os.ErrClosed} {
		if errors.Is(err, lost) {
			return true
		}
	}
	return lostByText(err.Error())
}

func lostByText(text string) bool {
	for _, sign := range []string{"broken pipe", "file already closed", "use of closed"} {
		if strings.Contains(text, sign) {
			return true
		}
	}
	return false
}

func (w *lastWords) said() string {
	timer := time.NewTimer(drainWait)
	defer timer.Stop()
	select {
	case <-w.drained:
	case <-timer.C:
	}
	if text := w.text.String(); text != "" {
		return "; the director said: " + text
	}
	return ""
}

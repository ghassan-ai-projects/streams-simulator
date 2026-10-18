package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const (
	followInterval = 200 * time.Millisecond
	followChunk    = 4 << 20
)

type follower struct {
	bridge   *Bridge
	rel      string
	offset   int64
	line     int64
	after    int64
	emit     func(domain.Record) error
	caughtUp bool
}

func (b *Bridge) OpenRun(worldID string) error {
	if _, err := domain.RunJournal(worldID); err != nil {
		return err
	}
	if !b.edges.Workspace.Exists("runs/" + worldID) {
		return fmt.Errorf("%w: run %s", domain.ErrAbsent, worldID)
	}
	return nil
}

func (b *Bridge) Follow(ctx context.Context, worldID string, after int64, emit func(domain.Record) error) error {
	path, err := domain.RunJournal(worldID)
	if err != nil {
		return err
	}
	f := &follower{bridge: b, rel: path.Rel, after: after, emit: emit}
	return b.everyTick(ctx, followInterval, f.drain)
}

func (b *Bridge) everyTick(ctx context.Context, every time.Duration, step func() (bool, error)) error {
	ticks, stop := b.edges.Tick(every)
	defer stop()
	for {
		if done, err := step(); err != nil || done {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticks:
		}
	}
}

func (f *follower) drain() (bool, error) {
	chunk, read := f.readChunk()
	if !read {
		return false, nil
	}
	if end := bytes.LastIndexByte(chunk, '\n'); end >= 0 {
		f.offset += int64(end + 1)
		if done, failed := f.emitLines(chunk[:end]); failed != nil || done {
			return done, failed
		}
	}
	return false, f.signalCaughtUp(len(chunk) < followChunk)
}

func (f *follower) readChunk() ([]byte, bool) {
	chunk, err := f.bridge.edges.Workspace.ReadFrom(f.rel, f.offset, followChunk)
	return chunk, err == nil
}

func (f *follower) signalCaughtUp(atEnd bool) error {
	if !atEnd || f.caughtUp {
		return nil
	}
	f.caughtUp = true
	return f.emit(domain.Record{CaughtUp: true})
}

func (f *follower) emitLines(lines []byte) (bool, error) {
	for line := range bytes.SplitSeq(lines, []byte("\n")) {
		if done, err := f.handle(line); err != nil || done {
			return done, err
		}
	}
	return false, nil
}

func (f *follower) handle(line []byte) (bool, error) {
	record, err := parseRecord(line, f.line)
	if err != nil {
		return false, err
	}
	f.line++
	if record.Index <= f.after {
		record.Seen = true
		return record.Closes, f.emitSeenClose(record)
	}
	if err := f.emit(record); err != nil || record.Closes {
		return true, err
	}
	return false, nil
}

func parseRecord(line []byte, index int64) (domain.Record, error) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &head); err != nil {
		return domain.Record{}, fmt.Errorf("canvasbridge: journal line %d: %w", index, err)
	}
	return domain.Record{Index: index, Type: head.Type, Line: bytes.Clone(line), Closes: head.Type == "run.closed"}, nil
}

func (f *follower) emitSeenClose(record domain.Record) error {
	if !record.Closes {
		return nil
	}
	return f.emit(record)
}

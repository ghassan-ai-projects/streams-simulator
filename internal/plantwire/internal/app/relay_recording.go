package app

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type wireSink interface {
	Wire(direction string, record map[string]any)
	Flush() error
}

type runWire struct{ run *run.Run }

func (w runWire) Wire(direction string, record map[string]any) { w.run.JournalWire(direction, record) }

func (w runWire) Flush() error {
	if err := w.run.FlushJournal(); err != nil {
		return fmt.Errorf("flush the run journal: %w", err)
	}
	return nil
}

type noteSink interface {
	JournalOrigin(fields map[string]any)
	JournalBinding(fields map[string]any)
	JournalLive(fields map[string]any)
}

type fieldSink interface {
	Add(record map[string]any) error
	Flush() error
	Close() error
	Digest() string
}

type recording struct {
	wire        wireSink
	field       fieldSink
	notes       noteSink
	closeWire   func() error
	incomplete  bool
	live        *domain.Live
	liveChanged bool
	clock       func() int64
	ints        *interruptFile
	interrupt   func(map[string]any)
	dir         string
}

func openRecording(dir string, sim *simNode, withControl bool) (recording, error) {
	rec, err := openTrace(dir, sim)
	if err != nil || !withControl || dir == "" {
		return rec, err
	}
	return rec.withInterrupts()
}

func (c recording) withInterrupts() (recording, error) {
	file, err := openInterruptFile(c.dir)
	if err != nil {
		return c, errors.Join(err, c.close())
	}
	c.ints = file
	return c, nil
}

func openTrace(dir string, sim *simNode) (recording, error) {
	if dir == "" {
		return recording{}, nil
	}
	field, err := run.NewFieldTrace(dir)
	if err != nil {
		return recording{}, fmt.Errorf("plantwire relay: %w", err)
	}
	if sim != nil {
		return recording{wire: runWire{sim.run}, field: field, notes: sim.run, dir: dir, interrupt: sim.run.JournalInterrupt,
			live: domain.NewLive(true, sim.instances, sim.bound), clock: func() int64 { return sim.run.Status().ClockNS }}, nil
	}
	return hardwareRecording(dir, field)
}

func hardwareRecording(dir string, field *run.FieldTrace) (recording, error) {
	trace, err := run.NewWireTrace(dir)
	if err != nil {
		return recording{}, errors.Join(fmt.Errorf("plantwire relay: %w", err), field.Close())
	}
	return recording{wire: trace, field: field, closeWire: trace.Close, dir: dir, interrupt: trace.Interrupt}, nil
}

func (c *recording) empty() bool { return c.wire == nil && c.field == nil }

func (c *recording) note(line domain.Traced) error {
	if c.wire != nil {
		c.wire.Wire(line.Direction, line.Record)
	}
	c.liveRecord(line)
	if c.field != nil && line.Field && !c.incomplete {
		return c.addField(line.Record)
	}
	return nil
}

func (c *recording) addField(record map[string]any) error {
	err := c.field.Add(record)
	if errors.Is(err, run.ErrFieldTraceFull) || errors.Is(err, run.ErrFieldTrace) {
		slog.Warn("plantwire relay: the field trace is incomplete, the gateway is disconnected", "error", err)
		c.incomplete = true
		return nil
	}
	if err != nil {
		return fmt.Errorf("plantwire relay: field trace: %w", err)
	}
	return nil
}

func (c *recording) noted(notes []domain.Note) {
	if c.notes == nil {
		return
	}
	for _, note := range notes {
		c.recordNote(note)
	}
}

func (c *recording) recordNote(note domain.Note) {
	switch note.Type {
	case domain.NoteConnection:
		c.liveConnection(note.Fields)
	case domain.NoteOrigin:
		c.notes.JournalOrigin(note.Fields)
	default:
		c.notes.JournalBinding(note.Fields)
	}
}

func (c *recording) flush() error {
	var wired, fielded error
	if c.wire != nil {
		wired = c.wire.Flush()
	}
	if c.field != nil {
		fielded = c.field.Flush()
	}
	return errors.Join(wired, fielded)
}

func (c *recording) close() error {
	c.endLive()
	var wired, fielded error
	if c.closeWire != nil {
		wired = c.closeWire()
	}
	if c.field != nil {
		fielded = c.field.Close()
	}
	if err := errors.Join(wired, fielded, c.ints.finish(c.dir)); err != nil {
		return fmt.Errorf("plantwire relay: close trace: %w", err)
	}
	return nil
}

func (c *recording) interruptDigest() string {
	return c.ints.digest()
}

func (c *recording) interrupts(entries []domain.InterruptEntry) {
	recordInterrupts(entries, c.ints, c.interrupt)
}

func (c *recording) digest() string {
	if c.field == nil || c.incomplete {
		return ""
	}
	return c.field.Digest()
}

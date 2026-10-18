package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"

func (s *Session) journalLive(direction string, record map[string]any) {
	if s.live == nil {
		return
	}
	in, err := domain.InboundOf(record)
	in.Within(s.options.SessionID, domain.Version)
	if err == nil && in.Invalid == nil && s.live.Observe(direction, record, s.run.Status().ClockNS) {
		s.liveChanged = true
	}
}

func (s *Session) flushLive() {
	if s.live != nil && s.liveChanged {
		s.liveChanged = false
		s.run.JournalLive(s.live.Fields())
	}
}

func (s *Session) endLive() {
	if s.live != nil {
		s.live.End(s.run.Status().ClockNS)
		s.liveChanged = true
		s.flushLive()
	}
}

func (c *recording) liveRecord(line domain.Traced) {
	if c.live == nil || (line.Direction != domain.TraceNode && !line.Admitted) {
		return
	}
	if c.live.Observe(line.Direction, line.Record, c.clock()) {
		c.liveChanged = true
	}
}

func (c *recording) liveConnection(fields map[string]any) {
	if c.live == nil {
		return
	}
	side, _ := fields["side"].(string)
	state, _ := fields["state"].(string)
	if c.live.Connection(side, state, c.clock()) {
		c.liveChanged = true
	}
}

func (c *recording) endLive() {
	if c.live != nil && c.live.End(c.clock()) {
		c.liveChanged = true
	}
	c.flushLive()
}

func (c *recording) tickLive(milliseconds int64) {
	if c.live != nil && c.live.Tick(milliseconds) {
		c.liveChanged = true
	}
	c.flushLive()
}

func (c *recording) flushLive() {
	if c.live != nil && c.liveChanged {
		c.liveChanged = false
		c.notes.JournalLive(c.live.Fields())
	}
}

package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"

func (s *Session) journalLive(direction string, record map[string]any) {
	if s.live == nil {
		return
	}
	in, err := domain.InboundOf(record)
	in.Within(s.options.SessionID, domain.Version)
	if err == nil && in.Invalid == nil && s.live.Observe(direction,record,s.run.Status().ClockNS) {
		s.run.JournalLive(s.live.Fields())
	}
}

func (s *Session) endLive() {
	if s.live != nil {
		s.live.Connection("host","offline",s.run.Status().ClockNS)
		s.run.JournalLive(s.live.Fields())
	}
}

func (c *recording) liveRecord(line domain.Traced) {
	if c.live == nil || (line.Direction != domain.TraceNode && !line.Admitted) {
		return
	}
	if c.live.Observe(line.Direction,line.Record,c.clock()) {
		c.notes.JournalLive(c.live.Fields())
	}
}

func (c *recording) liveConnection(fields map[string]any) {
	if c.live == nil {
		return
	}
	side, _ := fields["side"].(string)
	state, _ := fields["state"].(string)
	c.live.Connection(side,state,c.clock())
	c.notes.JournalLive(c.live.Fields())
}

func (c *recording) tickLive(milliseconds int64) {
	if c.live != nil && c.live.Tick(milliseconds) {
		c.notes.JournalLive(c.live.Fields())
	}
}

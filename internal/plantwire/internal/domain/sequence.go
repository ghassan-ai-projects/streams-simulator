package domain

import "maps"

type Transmitter struct {
	session string
	version int
	next    uint64
	control []Record
	data    []Record
}

func NewTransmitter(session string) *Transmitter {
	return &Transmitter{session: session, version: Version}
}

func (t *Transmitter) UseVersion(version int) { t.version = version }

func (t *Transmitter) Queue(record Record) {
	if record.Lane == LaneControl {
		t.control = append(t.control, record)
		return
	}
	t.data = append(t.data, record)
}

func (t *Transmitter) Drain() []map[string]any {
	queued := append(t.control, t.data...)
	t.control, t.data = nil, nil
	out := make([]map[string]any, len(queued))
	for i, record := range queued {
		out[i] = t.stamp(record)
	}
	return out
}

func (t *Transmitter) LastSeq() uint64 { return t.next - 1 }

func (t *Transmitter) stamp(record Record) map[string]any {
	stamped := maps.Clone(record.Fields)
	stamped["type"], stamped["lane"], stamped["session_id"] = record.Type, record.Lane, t.session
	stamped["protocol_version"], stamped["seq"] = t.version, t.next
	t.next++
	return stamped
}

type Receiver struct{ next uint64 }

type Arrival int

const (
	InOrder Arrival = iota
	Duplicate
	Gap
)

func (r *Receiver) Accept(seq uint64) Arrival {
	switch {
	case seq == r.next:
		r.next++
		return InOrder
	case seq < r.next:
		return Duplicate
	}
	r.next = seq + 1
	return Gap
}

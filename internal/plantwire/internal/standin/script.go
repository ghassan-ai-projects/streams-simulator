package standin

import (
	"cmp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const (
	LinkDown = "down"
	LinkUp   = "up"
)

type Script struct {
	Steps   []Step
	Replies []Reply
}

type Step struct {
	AtNS    int64
	Device  string
	Samples []Sample
	Link    string
	BootID  string
}

type Sample struct {
	Metric  string
	Value   any
	Quality string
}

func (n *Node) Play(untilNS int64) []map[string]any {
	var out []map[string]any
	for len(n.steps) > 0 && n.steps[0].AtNS <= untilNS {
		n.queueStep(n.steps[0])
		n.steps = n.steps[1:]
		out = append(out, n.out.Drain()...)
	}
	return out
}

func (n *Node) queueStep(step Step) {
	switch step.Link {
	case LinkDown:
		n.out.Queue(n.death(step))
	case LinkUp:
		n.boots[step.Device] = cmp.Or(step.BootID, n.boots[step.Device])
		n.out.Queue(n.deviceBirth(n.devices[step.Device], step.AtNS))
	default:
		n.out.Queue(n.data(step))
	}
}

func (n *Node) death(step Step) domain.Record {
	return domain.Record{Type: "DDEATH", Lane: domain.LaneControl, Fields: map[string]any{"device": step.Device,
		"boot_id": n.boots[step.Device], "reason": "link_down", "last_seq": n.out.LastSeq()}}
}

func (n *Node) data(step Step) domain.Record {
	samples := make([]map[string]any, len(step.Samples))
	for i, sample := range step.Samples {
		samples[i] = domain.Sample(n.alias(step.Device, sample.Metric), sample.Value, sample.Quality, step.AtNS)
	}
	return domain.DeviceData(step.Device, n.boots[step.Device], false, samples)
}

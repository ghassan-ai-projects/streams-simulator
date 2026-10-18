package standin

import (
	"cmp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

type Reply struct {
	CommandID string
	Command   string
	Receipt   string
	Reason    string
	Result    string
}

func (n *Node) Handle(record map[string]any) []map[string]any {
	command, err := messageOf(record)
	if err != nil || command.Type != "DCMD" {
		return nil
	}
	reply := n.replyTo(command)
	boot := cmp.Or(n.boots[command.Device], command.ExpectedBootID)
	records := []domain.Record{domain.Receipt(command.CommandID, reply.Receipt, reply.Reason, boot)}
	if reply.Result != "" {
		records = append(records, domain.Result(command.CommandID, reply.Result, boot))
	}
	return n.Emit(records...)
}

func (n *Node) replyTo(command domain.Message) Reply {
	device, onBoard := n.devices[command.Device]
	switch {
	case !onBoard:
		return refusal("unknown_device")
	case command.ExpectedBootID != n.boots[device.Name]:
		return refusal("wrong_boot")
	case !n.kinds[device.Kind].commands[command.Command]:
		return refusal("unknown_command")
	}
	return n.scripted(command)
}

func (n *Node) scripted(command domain.Message) Reply {
	replies := n.config.Script.Replies
	if at := slices.IndexFunc(replies, func(r Reply) bool { return r.CommandID == command.CommandID }); at >= 0 {
		return replies[at]
	}
	if at := slices.IndexFunc(replies, func(r Reply) bool { return r.CommandID == "" && r.Command == command.Command }); at >= 0 {
		return replies[at]
	}
	return refusal("not_ready")
}

func refusal(reason string) Reply {
	return Reply{Receipt: "refused", Reason: reason}
}

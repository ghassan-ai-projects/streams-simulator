package standin

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

var (
	receipts = []string{"accepted", "refused", "queued"}
	results  = []string{"", "succeeded", "failed", "unknown"}
)

func (n *Node) admit() error {
	for _, check := range []func() error{n.admitNames, n.admitDevices, n.admitScript} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) admitNames() error {
	for _, id := range []string{n.config.Session, n.config.Board} {
		if !wireSafe(id) {
			return fmt.Errorf("%w: %q is no wire id", ErrConfig, id)
		}
	}
	if !slices.Contains([]string{"", "hardware", "simulated"}, n.config.Environment) {
		return fmt.Errorf("%w: environment %q is neither hardware nor simulated", ErrConfig, n.config.Environment)
	}
	return nil
}

func wireSafe(id string) bool {
	return domain.WireID(id) == id
}

func (n *Node) admitDevices() error {
	for _, device := range n.config.Devices {
		if err := n.admitDevice(device); err != nil {
			return err
		}
		n.devices[device.Name], n.boots[device.Name] = device, device.BootID
	}
	return nil
}

func (n *Node) admitDevice(device Device) error {
	_, known := n.kinds[device.Kind]
	_, twice := n.devices[device.Name]
	switch {
	case !wireSafe(device.Name) || !wireSafe(device.BootID):
		return fmt.Errorf("%w: device %q with boot %q needs wire ids", ErrConfig, device.Name, device.BootID)
	case !known:
		return fmt.Errorf("%w: device %q is of kind %q, which no kind file defines", ErrConfig, device.Name, device.Kind)
	case twice:
		return fmt.Errorf("%w: device %q is given twice", ErrConfig, device.Name)
	}
	return nil
}

func (n *Node) admitScript() error {
	for i, step := range n.config.Script.Steps {
		if err := n.admitStep(step); err != nil {
			return fmt.Errorf("step %d: %w", i, err)
		}
	}
	for i, reply := range n.config.Script.Replies {
		if err := admitReply(reply); err != nil {
			return fmt.Errorf("reply %d: %w", i, err)
		}
	}
	return nil
}

func (n *Node) admitStep(step Step) error {
	device, onBoard := n.devices[step.Device]
	switch {
	case !onBoard:
		return fmt.Errorf("%w: device %q is not on the board", ErrConfig, step.Device)
	case (len(step.Samples) > 0) == (step.Link != ""):
		return fmt.Errorf("%w: a step carries samples or a link change, not both and not neither", ErrConfig)
	case step.Link != "" && step.Link != LinkDown && step.Link != LinkUp, step.BootID != "" && !wireSafe(step.BootID):
		return fmt.Errorf("%w: link %q with boot %q is not down or up with a wire id", ErrConfig, step.Link, step.BootID)
	}
	return admitSamples(step, n.kinds[device.Kind])
}

func admitSamples(step Step, kind kindShape) error {
	for _, sample := range step.Samples {
		if !slices.Contains(kind.metrics, sample.Metric) {
			return fmt.Errorf("%w: device %q has no metric %q", ErrConfig, step.Device, sample.Metric)
		}
	}
	return nil
}

func admitReply(reply Reply) error {
	named := reply.Command != "" || reply.CommandID != ""
	if !named || !slices.Contains(receipts, reply.Receipt) || !slices.Contains(results, reply.Result) {
		return fmt.Errorf("%w: a reply needs a command or command id, a receipt of accepted, refused or queued, and a result of succeeded, failed, unknown or none", ErrConfig)
	}
	return nil
}

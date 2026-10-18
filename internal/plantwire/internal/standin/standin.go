// Package standin plays a scripted plant wire node for tests, standing in for a board behind a gateway: no I/O, no clock, script time only.
package standin

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

var ErrConfig = errors.New("standin: invalid configuration")

type Device struct {
	Name, Kind, BootID string
	Initial            map[string]any
}

type Config struct {
	Session     string
	Board       string
	Environment string
	Kinds       []map[string]any
	Devices     []Device
	Script      Script
}

type Node struct {
	config  Config
	out     *domain.Transmitter
	kinds   map[string]kindShape
	devices map[string]Device
	boots   map[string]string
	steps   []Step
}

type kindShape struct {
	metrics  []string
	commands map[string]bool
}

func New(config Config) (*Node, error) {
	kinds, err := shapes(config.Kinds)
	if err != nil {
		return nil, err
	}
	node := &Node{config: config, out: domain.NewTransmitter(config.Session), kinds: kinds, devices: map[string]Device{}, boots: map[string]string{}}
	if err := node.admit(); err != nil {
		return nil, err
	}
	node.steps = slices.SortedStableFunc(slices.Values(config.Script.Steps), func(a, b Step) int { return cmp.Compare(a.AtNS, b.AtNS) })
	return node, nil
}

func shapes(kinds []map[string]any) (map[string]kindShape, error) {
	found := map[string]kindShape{}
	for i, entry := range kinds {
		if err := domain.CheckKind(entry); err != nil {
			return nil, fmt.Errorf("%w: kind file %d: %w", ErrConfig, i, err)
		}
		name, _ := entry["kind"].(string)
		if _, twice := found[name]; twice {
			return nil, fmt.Errorf("%w: kind %q is given twice", ErrConfig, name)
		}
		found[name] = shapeOf(entry)
	}
	return found, nil
}

func shapeOf(entry map[string]any) kindShape {
	commands := map[string]bool{}
	for _, name := range namesOf(entry["commands"]) {
		commands[name] = true
	}
	return kindShape{metrics: namesOf(entry["metrics"]), commands: commands}
}

func namesOf(list any) []string {
	items, _ := list.([]any)
	names := make([]string, 0, len(items))
	for _, item := range items {
		fields, _ := item.(map[string]any)
		name, _ := fields["name"].(string)
		names = append(names, name)
	}
	return names
}

func (n *Node) alias(device, metric string) int {
	return slices.Index(n.kinds[n.devices[device].Kind].metrics, metric) + 1
}

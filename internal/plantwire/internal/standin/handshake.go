package standin

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

var ErrRecord = errors.New("standin: not a valid plant wire record")

func (n *Node) Birth() (map[string]any, error) {
	identity := domain.NodeIdentity{Environment: cmp.Or(n.config.Environment, "hardware"), ProtocolMax: domain.Version, PriorLastSeq: -1, ClockMode: "wall"}
	record, err := domain.NodeBirthOf(identity, n.templates(), []model.Board{{Name: n.config.Board, Members: n.members()}})
	if err != nil {
		return nil, fmt.Errorf("standin: %w", err)
	}
	n.out.Queue(record)
	return n.out.Drain()[0], nil
}

func (n *Node) templates() []any {
	templates := make([]any, len(n.config.Kinds))
	for i, kind := range n.config.Kinds {
		templates[i] = maps.Clone(kind)
	}
	return templates
}

func (n *Node) members() []string {
	members := make([]string, len(n.config.Devices))
	for i, device := range n.config.Devices {
		members[i] = device.Name
	}
	return members
}

func (n *Node) Accept(state map[string]any) ([]map[string]any, error) {
	message, err := messageOf(state)
	if err != nil {
		return nil, err
	}
	version, err := domain.Negotiate(message, domain.Version)
	if err != nil {
		return nil, fmt.Errorf("standin: handshake: %w", err)
	}
	n.out.UseVersion(version)
	for _, device := range n.config.Devices {
		n.out.Queue(n.deviceBirth(device, 0))
	}
	return n.out.Drain(), nil
}

func (n *Node) deviceBirth(device Device, atNS int64) domain.Record {
	return domain.Record{Type: "DBIRTH", Lane: domain.LaneData, Fields: map[string]any{"device": device.Name, "kind": device.Kind,
		"board": n.config.Board, "boot_id": n.boots[device.Name], "metadata_version": 1, "birth_time_ns": atNS, "metrics": n.birthMetrics(device)}}
}

func (n *Node) birthMetrics(device Device) []any {
	metrics := []any{}
	for i, name := range n.kinds[device.Kind].metrics {
		value, quality := device.Initial[name], "good"
		if value == nil {
			quality = "warming_up"
		}
		metrics = append(metrics, map[string]any{"name": name, "alias": i + 1, "value": value, "quality": quality})
	}
	return metrics
}

func (n *Node) Shutdown() map[string]any {
	n.out.Queue(domain.NodeDeath(0, "shutdown"))
	return n.out.Drain()[0]
}

func (n *Node) Emit(records ...domain.Record) []map[string]any {
	for _, record := range records {
		n.out.Queue(record)
	}
	return n.out.Drain()
}

func messageOf(record map[string]any) (domain.Message, error) {
	raw, err := json.Marshal(record)
	if err != nil {
		return domain.Message{}, fmt.Errorf("%w: %w", ErrRecord, err)
	}
	inbound, ok := domain.Decode(raw)
	if !ok || inbound.Invalid != nil {
		return domain.Message{}, errors.Join(ErrRecord, inbound.Invalid)
	}
	return inbound.Message, nil
}

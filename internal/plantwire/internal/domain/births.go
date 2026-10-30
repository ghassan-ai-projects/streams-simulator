package domain

import (
	"cmp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const ContractVersion = "1.5.0"

type NodeIdentity struct {
	Environment  string
	ProtocolMax  int
	BDSeq        uint64
	PriorLastSeq int64
	ClockMode    string
}

func NodeBirth(identity NodeIdentity, kinds []*model.Kind, boards []model.Board) (Record, error) {
	return NodeBirthOf(identity, KindTemplates(kinds), boards)
}

func NodeBirthOf(identity NodeIdentity, templates []any, boards []model.Board) (Record, error) {
	templates, err := WithKindDigests(templates)
	if err != nil {
		return Record{}, err
	}
	return newRecord("NBIRTH", LaneControl, map[string]any{
		"protocol_min": Version, "protocol_max": identity.ProtocolMax, "contract_version": ContractVersion,
		"contract_digest": ContractDigest(), "environment": identity.Environment, "bd_seq": identity.BDSeq,
		"prior_last_seq": identity.PriorLastSeq, "capability_digest": CapabilityDigest(templates), "kinds": templates,
		"boards": boardList(boards), "node_metrics": map[string]any{"clock_mode": cmp.Or(identity.ClockMode, "virtual"), "run_state": "running"}}), nil
}

type DeviceIdentity struct {
	Device, Kind, Board, BootID string
	BirthTimeNS                 int64
}

func DeviceBirth(identity DeviceIdentity, base, effective *model.Kind, current map[string]any) Record {
	fields := map[string]any{"device": identity.Device, "kind": identity.Kind, "board": identity.Board, "boot_id": identity.BootID,
		"metadata_version": 1, "birth_time_ns": identity.BirthTimeNS, "metrics": currentValues(effective, current)}
	if overrides := overridesOf(base, effective); len(overrides) > 0 {
		fields["overrides"] = overrides
	}
	return newRecord("DBIRTH", LaneData, fields)
}

func currentValues(kind *model.Kind, current map[string]any) []any {
	values := []any{}
	for i := range kind.Channels {
		if dataType(&kind.Channels[i]) != "" {
			values = append(values, currentValue(kind, kind.Channels[i].Name, current))
		}
	}
	return values
}

func currentValue(kind *model.Kind, channel string, current map[string]any) map[string]any {
	value, seen := current[channel]
	quality := "good"
	if !seen || value == nil {
		quality = "warming_up"
	}
	return map[string]any{"name": channel, "alias": Alias(kind, channel), "value": value, "quality": quality}
}

func Alias(kind *model.Kind, channel string) int {
	for i := range kind.Channels {
		if kind.Channels[i].Name == channel {
			return i + 1
		}
	}
	return 0
}

func boardList(boards []model.Board) []any {
	out := make([]any, 0, len(boards))
	for _, board := range boards {
		out = append(out, map[string]any{"board": board.Name, "environment": cmp.Or(board.Environment, EnvironmentSimulated), "members": toAny(board.Members)})
	}
	return out
}

func Carried(ch *model.Channel) bool { return ch != nil && dataType(ch) != "" }

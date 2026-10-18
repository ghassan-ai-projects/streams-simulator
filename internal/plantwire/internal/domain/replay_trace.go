package domain

import (
	"encoding/json"
	"errors"
	"maps"
)

var ErrNoReplayBirth = errors.New("plantwire: the field trace holds no NBIRTH to replay")

type ReplayChunk struct {
	Records     []Record
	WorldTimeNS int64
}

type ReplayPlan struct {
	Chunks  []ReplayChunk
	Skipped int
	Births  uint64
	Boots   map[string]string
}

type replayPlanner struct {
	plan    ReplayPlan
	gateway Receiver
	births  uint64
	worldNS int64
	begun   bool
	current ReplayChunk
}

func PlanReplay(trace []map[string]any) (ReplayPlan, error) {
	planner := replayPlanner{plan: ReplayPlan{Boots: map[string]string{}}}
	for _, record := range trace {
		planner.take(record)
	}
	planner.cut()
	if len(planner.plan.Chunks) == 0 {
		return ReplayPlan{}, ErrNoReplayBirth
	}
	planner.plan.Births = planner.births
	return planner.plan, nil
}

func (p *replayPlanner) take(record map[string]any) {
	in, err := InboundOf(record)
	switch {
	case err != nil || in.Invalid != nil || p.stale(&in):
		p.plan.Skipped++
	case in.Type == "NBIRTH":
		p.add(replayedBirth(in, p.births))
		p.births, p.begun = p.births+1, true
	case !p.begun:
		p.plan.Skipped++
	default:
		p.carry(in)
	}
}

func (p *replayPlanner) stale(in *Inbound) bool {
	if in.Type == "NBIRTH" && in.Seq == 0 {
		p.gateway = Receiver{}
	}
	return arrivalOf(&p.gateway, *in) == Duplicate
}

func (p *replayPlanner) carry(in Inbound) {
	switch in.Type {
	case "DBIRTH", "DDATA", "DDEATH", "DSTATE":
		p.note(in)
		p.add(carried(in))
	case "NSYNC":
		p.cut()
	default:
		p.plan.Skipped++
	}
}

func (p *replayPlanner) note(in Inbound) {
	switch in.Type {
	case "DBIRTH":
		p.plan.Boots[in.Device] = in.BootID
	case "DDATA":
		p.worldNS = max(p.worldNS, latestEventTime(in.Record))
	}
}

func (p *replayPlanner) add(record Record) {
	p.current.Records = append(p.current.Records, record)
}

func (p *replayPlanner) cut() {
	if len(p.current.Records) == 0 {
		return
	}
	p.current.WorldTimeNS = p.worldNS
	p.plan.Chunks = append(p.plan.Chunks, p.current)
	p.current = ReplayChunk{}
}

func carried(in Inbound) Record {
	return Record{Type: in.Type, Lane: text(in.Record, "lane"), Fields: content(in.Record)}
}

func replayedBirth(in Inbound, births uint64) Record {
	fields := content(in.Record)
	fields["environment"] = EnvironmentSimulated
	fields["boards"] = simulatedBoards(listOf(fields["boards"]))
	fields["node_metrics"] = map[string]any{"clock_mode": "virtual", "run_state": "running"}
	fields["bd_seq"], fields["prior_last_seq"] = births, -1
	fields["protocol_max"] = min(in.ProtocolMax, relayProtocol)
	return Record{Type: "NBIRTH", Lane: LaneControl, Fields: fields}
}

func simulatedBoards(boards []any) []any {
	replayed := make([]any, len(boards))
	for i, board := range boards {
		fields, _ := board.(map[string]any)
		copied := maps.Clone(fields)
		if copied == nil {
			copied = map[string]any{}
		}
		copied["environment"] = EnvironmentSimulated
		replayed[i] = copied
	}
	return replayed
}

func latestEventTime(record map[string]any) int64 {
	var latest int64
	for _, entry := range listOf(record["metrics"]) {
		sample, _ := entry.(map[string]any)
		if at, ok := sample["event_time_ns"].(json.Number); ok {
			ns, _ := at.Int64()
			latest = max(latest, ns)
		}
	}
	return latest
}

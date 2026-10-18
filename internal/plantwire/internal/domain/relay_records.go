package domain

import "maps"

var (
	envelopeMembers = []string{"type", "lane", "session_id", "protocol_version", "seq"}
	nodeRecords     = map[string]bool{"NBIRTH": true, "DBIRTH": true, "DDATA": true, "DDEATH": true, "NDEATH": true,
		"DRECEIPT": true, "DRESULT": true, "DSTATE": true, "NSYNC": true}
	hostRecords = map[string]bool{"STATE": true, "NCMD": true, "DCMD": true, "DQUERY": true, "HSYNC": true}
)

func content(record map[string]any) map[string]any {
	fields := maps.Clone(record)
	for _, member := range envelopeMembers {
		delete(fields, member)
	}
	return fields
}

func relayState(online bool, refusal string) map[string]any {
	state := map[string]any{"host_id": relayName, "online": online, "heartbeat": 0, "heartbeat_interval_ms": 1000,
		"authority_epoch": 0, "protocol_min": Version, "protocol_max": relayProtocol}
	if refusal != "" {
		state["refusal"] = refusal
	}
	return state
}

func stateRecord(fields map[string]any) Record {
	return Record{Type: "STATE", Lane: LaneControl, Fields: maps.Clone(fields)}
}

func deviceDeath(device, boot string, lastSeq uint64) Record {
	return Record{Type: "DDEATH", Lane: LaneControl,
		Fields: map[string]any{"device": device, "boot_id": boot, "reason": "link_down", "last_seq": lastSeq}}
}

package domain

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

const (
	maxLiveCommands = 128
	maxLiveMarks    = 256
	maxLiveReadings = 512
	maxEvidence     = 128
	maxSeenCommands = 4096
	noBoot          = "none"
)

var linkWords = map[string]string{"host": "Host link", "gateway": "Gateway link"}

type LiveReading struct {
	Name    string `json:"name"`
	Unit    string `json:"unit"`
	Value   any    `json:"value"`
	Quality string `json:"quality"`
	State   string `json:"state"`
}

type LiveDevice struct {
	Device      string        `json:"device"`
	Board       string        `json:"board"`
	Instance    string        `json:"instance"`
	Environment string        `json:"environment"`
	BootID      string        `json:"boot_id"`
	Health      string        `json:"health"`
	Reboots     int           `json:"reboots"`
	Readings    []LiveReading `json:"readings"`
	kind        string
	aliases     map[uint64]string
	updated     map[string]int64
	unannounced bool
}

type LiveCommand struct {
	CommandID      string `json:"command_id"`
	Device         string `json:"device"`
	Environment    string `json:"environment"`
	Command        string `json:"command"`
	BootID         string `json:"boot_id"`
	Sent           bool   `json:"sent"`
	Receipt        string `json:"receipt"`
	Reason         string `json:"reason"`
	Report         string `json:"report"`
	Effect         string `json:"effect"`
	EffectEvidence string `json:"effect_evidence"`
	SentAtNS       int64  `json:"sent_at_ns"`
	UpdatedAtNS    int64  `json:"updated_at_ns"`
}

type LiveMark struct {
	AtNS    int64  `json:"at_ns"`
	Kind    string `json:"kind"`
	Subject string `json:"subject"`
	Label   string `json:"label"`
	Entity  string `json:"entity"`
}

type Live struct {
	host      string
	gateway   string
	kinds     map[string]map[string]map[string]any
	boards    map[string]string
	devices   map[string]*LiveDevice
	commands  []*LiveCommand
	marks     []LiveMark
	instances map[string]string
	bound     []Bound
	elapsedMS int64
	nowNS     int64
	truncated bool
	seen      map[string]bool
	seenOrder []string

	interrupts      map[string]any
	interruptDigest string
}

func NewLive(gateway bool, instances map[string]string, bound []Bound) *Live {
	state := "not_applicable"
	if gateway {
		state = "unknown"
	}
	return &Live{host: "unknown", gateway: state, kinds: map[string]map[string]map[string]any{}, boards: map[string]string{},
		devices: map[string]*LiveDevice{}, commands: []*LiveCommand{}, instances: instances, bound: bound, seen: map[string]bool{}}
}

func (l *Live) Fields() map[string]any {
	marks := append([]LiveMark{}, l.marks[max(0, len(l.marks)-maxLiveMarks):]...)
	l.marks = nil
	fields := map[string]any{"host": l.host, "gateway": l.gateway, "devices": l.sortedDevices(), "commands": l.commandCopies(), "marks": marks, "commands_truncated": l.truncated}
	if l.interrupts != nil {
		fields["interrupts"] = l.interrupts
	}
	return fields
}

func (l *Live) orderedDevices() []*LiveDevice {
	devices := make([]*LiveDevice, 0, len(l.devices))
	for _, id := range slices.Sorted(maps.Keys(l.devices)) {
		devices = append(devices, l.devices[id])
	}
	return devices
}

func (l *Live) sortedDevices() []*LiveDevice {
	devices := make([]*LiveDevice, 0, len(l.devices))
	for _, d := range l.orderedDevices() {
		copied := *d
		copied.Readings = slices.Clone(copied.Readings)
		devices = append(devices, &copied)
	}
	return devices
}

func (l *Live) commandCopies() []*LiveCommand {
	copies := make([]*LiveCommand, 0, len(l.commands))
	for _, c := range l.commands {
		copied := *c
		copies = append(copies, &copied)
	}
	return copies
}

func (l *Live) Observe(direction string, record map[string]any, at int64) bool {
	l.nowNS = at
	switch direction {
	case TraceNode:
		return l.nodeRecord(record, at)
	case TraceHost:
		return l.hostRecord(record, at)
	}
	return false
}

func (l *Live) mark(at int64, kind, subject, label, entity string) {
	l.marks = append(l.marks, LiveMark{AtNS: at, Kind: kind, Subject: subject, Label: label, Entity: entity})
}

func (l *Live) Connection(side, state string, at int64) bool {
	link := &l.gateway
	if side == "host" {
		link = &l.host
	}
	if *link == state || *link == "not_applicable" || (*link == "unknown" && state == "offline") {
		return false
	}
	*link = state
	l.mark(at, "connection", side, linkWords[side]+" "+state, side)
	if state == "offline" {
		l.disconnect()
	}
	return true
}

func (l *Live) End(at int64) bool {
	changed := l.Connection("host", "offline", at)
	changed = l.Connection("gateway", "offline", at) || changed
	return l.disconnect() || changed
}

func (l *Live) disconnect() bool {
	changed := false
	for _, d := range l.orderedDevices() {
		if d.Environment != EnvironmentSimulated {
			changed = l.disconnectDevice(d) || changed
		}
	}
	return changed
}

func (l *Live) disconnectDevice(d *LiveDevice) bool {
	changed := d.Health != "offline"
	d.Health = "offline"
	for i := range d.Readings {
		changed = changed || d.Readings[i].State != "disconnected"
		d.Readings[i].State = "disconnected"
	}
	return changed
}

func (l *Live) nodeRecord(r map[string]any, at int64) bool {
	switch text(r, "type") {
	case "DDATA", "DSTATE":
		return l.readings(r)
	case "DDEATH":
		return l.death(r, at)
	case "DRECEIPT", "DRESULT":
		return l.answer(r, at)
	default:
		return l.lifecycleRecord(r, at)
	}
}

func (l *Live) lifecycleRecord(r map[string]any, at int64) bool {
	switch text(r, "type") {
	case "NBIRTH":
		l.birth(r)
	case "DBIRTH":
		l.deviceBirth(r, at)
	case "NDEATH":
		l.Connection("host", "offline", at)
	default:
		return false
	}
	return true
}

func (l *Live) hostRecord(r map[string]any, at int64) bool {
	switch text(r, "type") {
	case "STATE":
		return l.hostState(r, at)
	case "DCMD":
		return l.sent(r, at)
	case "HSYNC":
		return l.evidence(r, at)
	default:
		return false
	}
}

func (l *Live) hostState(r map[string]any, at int64) bool {
	state := "offline"
	if _, refused := Negotiate(stateMessage(r), Version); refused == nil {
		state = "online"
	}
	return l.Connection("host", state, at)
}

func stateMessage(r map[string]any) Message {
	version, _ := jsonvalue.Number(r["protocol_version"])
	low, _ := jsonvalue.Number(r["protocol_min"])
	high, _ := jsonvalue.Number(r["protocol_max"])
	return Message{Online: r["online"] == true, ProtocolVersion: int64(version), ProtocolMin: int64(low), ProtocolMax: int64(high)}
}

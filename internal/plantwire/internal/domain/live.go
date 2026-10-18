package domain

import (
	"maps"
	"slices"
)

const maxLiveCommands = 128

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
}

type LiveCommand struct {
	CommandID      string `json:"command_id"`
	Device         string `json:"device"`
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
	AtNS   int64  `json:"at_ns"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Entity string `json:"entity"`
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
	truncated bool
}

func NewLive(gateway bool, instances map[string]string, bound []Bound) *Live {
	state := "not_applicable"
	if gateway {
		state = "unknown"
	}
	return &Live{host: "unknown", gateway: state, kinds: map[string]map[string]map[string]any{}, boards: map[string]string{},
		devices: map[string]*LiveDevice{}, commands: []*LiveCommand{}, instances: instances, bound: bound}
}

func (l *Live) Fields() map[string]any {
	marks := append([]LiveMark{}, l.marks...)
	l.marks = nil
	return map[string]any{"host": l.host, "gateway": l.gateway, "devices": l.sortedDevices(), "commands": l.commands, "marks": marks, "commands_truncated": l.truncated}
}

func (l *Live) sortedDevices() []*LiveDevice {
	devices := make([]*LiveDevice, 0, len(l.devices))
	for _, id := range slices.Sorted(maps.Keys(l.devices)) {
		devices = append(devices, l.devices[id])
	}
	return devices
}

func (l *Live) Observe(direction string, record map[string]any, at int64) bool {
	switch direction {
	case TraceNode:
		return l.nodeRecord(record, at)
	case TraceHost:
		return l.hostRecord(record, at)
	}
	return false
}

func (l *Live) mark(at int64, kind, label, entity string) {
	l.marks = append(l.marks, LiveMark{AtNS: at, Kind: kind, Label: label, Entity: entity})
}

func (l *Live) Connection(side, state string, at int64) {
	link := &l.gateway
	if side == "host" {
		link = &l.host
	}
	if *link == state {
		return
	}
	*link = state
	l.mark(at, "connection", linkWords[side]+" "+state, side)
	if state == "offline" {
		l.disconnect(side)
	}
}

func (l *Live) disconnect(side string) {
	for _, d := range l.sortedDevices() {
		if side == "host" || d.Environment == EnvironmentHardware {
			d.Health = "offline"
			l.disconnectReadings(d)
		}
	}
}

func (l *Live) disconnectReadings(d *LiveDevice) {
	for i := range d.Readings {
		d.Readings[i].State = "disconnected"
	}
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
	if r["online"] == true {
		state = "online"
	}
	if state == l.host {
		return false
	}
	l.Connection("host", state, at)
	return true
}

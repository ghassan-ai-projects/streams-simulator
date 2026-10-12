package domain

import (
	"sync"
)

type Config struct {
	DeviceID         string
	BootID           string
	FirmwareDigest   string
	CapabilityDigest string

	Plant Plant

	Clock func() int64

	Capabilities *Capabilities

	FaultSchedule []FaultInjection
}

type Faults struct {
	AckLost bool

	Stuck bool
}

type Outcome struct {
	Receipt map[string]any
	Result  map[string]any

	AckLost bool

	Duplicate bool

	Disconnect bool

	Fault           string
	AcceptedCommand int
}

type Device struct {
	mu               sync.Mutex
	deviceID         string
	bootID           string
	firmwareDigest   string
	capabilityDigest string
	plant            Plant
	clock            func() int64
	capabilities     *Capabilities
	faults           Faults
	faultSchedule    map[int][]string
	acceptedCommands int

	dedup            map[string]dedupEntry
	curTarget        string
	curOp            string
	curValue         float64
	energized        bool
	safeState        bool
	leaseUntilMicros int64
}

type dedupEntry struct {
	digest  string
	outcome Outcome
}

func New(cfg Config) *Device {
	d := configuredDevice(cfg)
	if cfg.Clock != nil {
		d.clock = cfg.Clock
	} else {

		d.clock = func() int64 { return 0 }
	}
	return d
}

func configuredDevice(cfg Config) *Device {
	digest := configuredCapabilityDigest(cfg)
	return &Device{deviceID: orDefault(cfg.DeviceID, "dev-01"), bootID: orDefault(cfg.BootID, "boot-A"),
		firmwareDigest: orDefault(cfg.FirmwareDigest, "sha256:"+repeat('c', 64)), capabilityDigest: digest,
		plant: cfg.Plant, capabilities: cfg.Capabilities, dedup: map[string]dedupEntry{}, faultSchedule: faultNames(cfg.FaultSchedule), safeState: true}
}

func configuredCapabilityDigest(cfg Config) string {
	digest := orDefault(cfg.CapabilityDigest, "sha256:"+repeat('d', 64))
	if cfg.Capabilities != nil {
		if catalogDigest := cfg.Capabilities.Digest(); catalogDigest != "" {
			digest = catalogDigest
		}
	}
	return digest
}

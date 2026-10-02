package device

import (
	"sync"
)

// Config configures a device instance. Identity and firmware digest default to
// fixed deterministic values matching the vendored contract's golden state
// frame. When Capabilities is supplied, its canonical digest is authoritative.
type Config struct {
	DeviceID         string
	BootID           string
	FirmwareDigest   string
	CapabilityDigest string
	// Plant is the physical process the device actuates. When nil, the device
	// derives the output value/energized state from the capability catalog's
	// declared energize_field (no external plant); bind a world-backed Plant to
	// make the world the effector oracle.
	Plant Plant
	// Clock returns the device monotonic time in microseconds. It must be
	// monotonic non-decreasing. Defaults to a manual clock starting at 0.
	Clock func() int64
	// Capabilities is the device's data-defined capability catalog (loaded via
	// LoadCapabilities). When nil the device declares no targets and rejects
	// every command wrong_target — capabilities are never hard-coded.
	Capabilities *Capabilities
	// FaultSchedule is an optional deterministic schedule of protocol faults.
	// Invalid schedules should be rejected with ValidateFaultSchedule before
	// constructing a device.
	FaultSchedule []FaultInjection
}

// Faults is the injectable protocol fault state. Zero value is a healthy
// device. Faults are applied to the next accepted command until cleared.
type Faults struct {
	// AckLost withholds the receipt from the wire even though the effect is
	// applied — the upstream sees an unknown outcome, not a failure.
	AckLost bool
	// Stuck accepts and "executes" the command but the plant does not move:
	// energized stays false. This is the desired≠observed case.
	Stuck bool
}

// Outcome is the device's response to one command.
type Outcome struct {
	Receipt map[string]any // the receipt record (always populated)
	Result  map[string]any // the terminal result record (always populated)
	// AckLost is true when an ack_lost fault means the receipt must NOT be
	// written to the wire, even though Result records what actually happened.
	AckLost bool
	// Duplicate asks the gateway loop to deliver this command to the device a
	// second time. The device's idempotency ledger still applies it once.
	Duplicate bool
	// Disconnect asks the gateway loop to close the current connection after
	// handling this command.
	Disconnect bool
	// Fault and AcceptedCommand are local evidence for deterministic fault logs;
	// neither field is serialized onto the device wire.
	Fault           string
	AcceptedCommand int
}

// Device is a wire-faithful serial device emulator. It is safe for use by one
// gateway connection at a time; methods are mutex-guarded.
type Device struct {
	mu               sync.Mutex
	deviceID         string
	bootID           string
	firmwareDigest   string
	capabilityDigest string
	plant            Plant
	clock            func() int64
	manualMono       int64
	capabilities     *Capabilities
	faults           Faults
	faultSchedule    map[int][]string
	acceptedCommands int

	dedup            map[string]dedupEntry // idempotency_key -> prior semantic command/outcome
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

// New builds a device from cfg, applying deterministic defaults.
func New(cfg Config) *Device {
	capabilityDigest := orDefault(cfg.CapabilityDigest, "sha256:"+repeat('d', 64))
	if cfg.Capabilities != nil {
		if digest := cfg.Capabilities.Digest(); digest != "" {
			capabilityDigest = digest
		}
	}
	d := &Device{
		deviceID:         orDefault(cfg.DeviceID, "dev-01"),
		bootID:           orDefault(cfg.BootID, "boot-A"),
		firmwareDigest:   orDefault(cfg.FirmwareDigest, "sha256:"+repeat('c', 64)),
		capabilityDigest: capabilityDigest,
		plant:            cfg.Plant,
		capabilities:     cfg.Capabilities,
		dedup:            map[string]dedupEntry{},
		faultSchedule:    faultNames(cfg.FaultSchedule),
		safeState:        true,
	}
	if cfg.Clock != nil {
		d.clock = cfg.Clock
	} else {
		d.clock = func() int64 { return d.manualMono }
	}
	return d
}

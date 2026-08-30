package device

import (
	"errors"
	"fmt"
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

	dedup            map[string]Outcome // idempotency_key -> prior outcome
	curTarget        string
	curOp            string
	curValue         float64
	energized        bool
	safeState        bool
	leaseUntilMicros int64
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
		dedup:            map[string]Outcome{},
		safeState:        true,
	}
	if cfg.Clock != nil {
		d.clock = cfg.Clock
	} else {
		d.clock = func() int64 { return d.manualMono }
	}
	return d
}

// Advance moves the default manual clock forward by deltaMicros. It is a no-op
// when a custom Clock was supplied.
func (d *Device) Advance(deltaMicros int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.manualMono += deltaMicros
	d.expireLease(d.clock())
}

// SetFaults installs the fault state applied to subsequent accepted commands.
func (d *Device) SetFaults(f Faults) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.faults = f
}

// Reboot simulates a power cycle: a new boot identity, safe outputs, and a
// cleared volatile dedup ledger. Commands bound to the old boot are then
// rejected wrong_boot, which is exactly what forces upstream reconciliation.
func (d *Device) Reboot(newBootID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.bootID = newBootID
	d.energized = false
	d.curTarget, d.curOp, d.curValue = "", "", 0
	d.safeState = true
	d.dedup = map[string]Outcome{}
	d.leaseUntilMicros = 0
}

// BootID returns the current device boot identity.
func (d *Device) BootID() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.bootID
}

// State returns a fresh device.state record.
func (d *Device) State() map[string]any {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.expireLease(d.clock())
	state := map[string]any{
		"message_type":      "state",
		"protocol_version":  float64(ProtocolVersion),
		"device_id":         d.deviceID,
		"boot_id":           d.bootID,
		"firmware_digest":   d.firmwareDigest,
		"capability_digest": d.capabilityDigest,
		"safe_state":        d.safeState,
		"dedup_ledger":      map[string]any{"persistent": false, "size": float64(len(d.dedup))},
	}
	if d.curTarget != "" {
		state["current_output"] = map[string]any{
			"target": d.curTarget, "operation": d.curOp,
			"value": d.curValue, "energized": d.energized,
		}
	}
	return state
}

// HandleCommand decodes, validates, and applies one command frame, returning
// the encoded receipt and result frames. When ackLost is true the receipt frame
// must not be written to the wire (the effect still happened).
func (d *Device) HandleCommand(frame []byte) (receipt []byte, result []byte, ackLost bool, err error) {
	command, decodeErr := DecodeRecord(frame)
	if decodeErr != nil {
		return nil, nil, false, decodeErr
	}
	if command["message_type"] != "command" {
		return nil, nil, false, fmt.Errorf("device: expected a command frame, got %v", command["message_type"])
	}
	outcome := d.ApplyCommand(command)
	receiptFrame, err := EncodeRecord(outcome.Receipt)
	if err != nil {
		return nil, nil, false, fmt.Errorf("device: encode receipt: %w", err)
	}
	resultFrame, err := EncodeRecord(outcome.Result)
	if err != nil {
		return nil, nil, false, fmt.Errorf("device: encode result: %w", err)
	}
	return receiptFrame, resultFrame, outcome.AckLost, nil
}

// ApplyCommand runs the device's admission logic on an already-decoded command
// and returns the receipt and result records. It is the deterministic core of
// the emulator; HandleCommand is the wire wrapper.
func (d *Device) ApplyCommand(command map[string]any) Outcome {
	d.mu.Lock()
	defer d.mu.Unlock()

	commandID, _ := command["command_id"].(string)
	idempotencyKey, _ := command["idempotency_key"].(string)
	now := d.clock()
	d.expireLease(now)

	// Idempotency: a repeated key replays the prior outcome and applies no new
	// effect, regardless of faults.
	if prior, ok := d.dedup[idempotencyKey]; ok {
		return prior
	}

	if reject := d.admit(command, now); reject != "" {
		return d.rejection(commandID, now, reject)
	}

	outcome, accepted := d.applyAcceptedCommand(command, commandID, now)
	if accepted {
		d.dedup[idempotencyKey] = outcome
	}
	return outcome
}

func (d *Device) applyAcceptedCommand(command map[string]any, commandID string, now int64) (Outcome, bool) {
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	params := numericParams(command["parameters"])
	plantCommand := PlantCommand{
		Target: target, Operation: operation, Params: params,
		CommandID: commandID, AtMicros: now,
	}
	value, energized, err := d.applyPlant(plantCommand)
	if err != nil {
		return d.rejection(commandID, now, plantRejectCode(err)), false
	}
	d.curTarget, d.curOp, d.curValue, d.energized = target, operation, value, energized
	d.safeState = !energized
	d.leaseUntilMicros = leaseDeadline(now, params)

	return Outcome{
		Receipt: map[string]any{
			"message_type":     "receipt",
			"protocol_version": float64(ProtocolVersion),
			"command_id":       commandID,
			"boot_id":          d.bootID,
			"accepted":         true,
			"received_mono_us": float64(now),
		},
		Result: map[string]any{
			"message_type":      "result",
			"protocol_version":  float64(ProtocolVersion),
			"command_id":        commandID,
			"boot_id":           d.bootID,
			"status":            "executed",
			"detail":            resultDetail(energized),
			"completed_mono_us": float64(now),
		},
		AckLost: d.faults.AckLost,
	}, true
}

// admit returns "" when the command may execute, or a device reject_code. All
// target/operation/bound knowledge comes from the data-defined capability
// catalog, never a code branch.
func (d *Device) admit(command map[string]any, now int64) string {
	if boot, _ := command["expected_boot_id"].(string); boot != d.bootID {
		return "wrong_boot"
	}
	notBefore, _ := command["not_before_mono_us"].(float64)
	expiresAfter, _ := command["expires_after_ms"].(float64)
	if float64(now) < notBefore {
		return "not_ready"
	}
	if float64(now) > notBefore+expiresAfter*1000 {
		return "expired"
	}
	target, _ := command["target"].(string)
	capa, ok := d.capabilities.target(target)
	if !ok {
		return "wrong_target"
	}
	if operation, _ := command["operation"].(string); operation != capa.Operation {
		return "unknown_operation"
	}
	params := numericParams(command["parameters"])
	if len(params) != len(capa.Bounds) {
		return "out_of_range"
	}
	for name, bounds := range capa.Bounds {
		v, ok := params[name]
		if !ok || v < bounds[0] || v > bounds[1] {
			return "out_of_range"
		}
	}
	return ""
}

// applyPlant resolves the output value and energized state for an accepted command.
// A bound Plant (e.g. the world oracle) is authoritative; otherwise the value is
// the command's energize_field and energized means that field is positive — both
// read from the capability catalog, never switched on an operation name.
func (d *Device) applyPlant(cmd PlantCommand) (value float64, energized bool, err error) {
	// A stuck device accepts the command but does not hand it to the plant; this
	// keeps a world-backed oracle aligned with the observed no-effect.
	if d.faults.Stuck {
		return 0, false, nil
	}
	if d.plant != nil {
		e, err := d.plant.Apply(cmd)
		if err != nil {
			return 0, false, fmt.Errorf("device: apply plant command: %w", err)
		}
		return e.Value, e.Energized, nil
	}
	capa, ok := d.capabilities.target(cmd.Target)
	if !ok {
		return 0, false, nil
	}
	value = cmd.Params[capa.EnergizeField]
	return value, value > 0, nil
}

func (d *Device) expireLease(now int64) {
	if d.leaseUntilMicros == 0 || now < d.leaseUntilMicros {
		return
	}
	d.energized = false
	d.safeState = true
	d.leaseUntilMicros = 0
}

func leaseDeadline(now int64, params map[string]float64) int64 {
	leaseMS, ok := params["lease_ms"]
	if !ok || leaseMS <= 0 {
		return 0
	}
	return now + int64(leaseMS*1000)
}

func plantRejectCode(err error) string {
	if errors.Is(err, ErrPlantInterlocked) {
		return "interlocked"
	}
	return "not_ready"
}

func (d *Device) rejection(commandID string, now int64, code string) Outcome {
	return Outcome{
		Receipt: map[string]any{
			"message_type":     "receipt",
			"protocol_version": float64(ProtocolVersion),
			"command_id":       commandID,
			"boot_id":          d.bootID,
			"accepted":         false,
			"reject_code":      code,
			"received_mono_us": float64(now),
		},
		Result: map[string]any{
			"message_type":     "result",
			"protocol_version": float64(ProtocolVersion),
			"command_id":       commandID,
			"boot_id":          d.bootID,
			"status":           "rejected",
			"error_code":       code,
		},
	}
}

func resultDetail(energized bool) string {
	if energized {
		return "output energized"
	}
	return "accepted, output not energized"
}

func numericParams(raw any) map[string]float64 {
	out := map[string]float64{}
	m, ok := raw.(map[string]any)
	if !ok {
		return out
	}
	for name, v := range m {
		if f, ok := v.(float64); ok {
			out[name] = f
		}
	}
	return out
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func repeat(b byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return string(out)
}

package device

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
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

// SetFaultSchedule installs a deterministic one-shot fault schedule. Entries
// are keyed by the one-based admission ordinal and are consumed when that
// ordinal is reached.
func (d *Device) SetFaultSchedule(schedule []FaultInjection) error {
	if err := ValidateFaultSchedule(schedule); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.faultSchedule = faultNames(schedule)
	d.acceptedCommands = 0
	return nil
}

// AcceptedCommandCount returns the number of commands that reached the
// schedule's admission ordinal. It is useful for deterministic test evidence.
func (d *Device) AcceptedCommandCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.acceptedCommands
}

// Reboot simulates a power cycle: a new boot identity, safe outputs, and a
// cleared volatile dedup ledger. Commands bound to the old boot are then
// rejected wrong_boot, which is exactly what forces upstream reconciliation.
func (d *Device) Reboot(newBootID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.rebootLocked(newBootID)
}

func (d *Device) rebootLocked(newBootID string) {
	safeStopSucceeded := d.invokeSafeStopLocked()
	d.bootID = newBootID
	if safeStopSucceeded {
		d.energized = false
		d.curTarget, d.curOp, d.curValue = "", "", 0
		d.safeState = true
	} else {
		// Preserve the last output as conservative evidence when the plant
		// could not confirm the safe transition.
		d.energized = true
		d.safeState = false
	}
	d.dedup = map[string]dedupEntry{}
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
	outcome, err := d.handleCommand(frame)
	if err != nil {
		return nil, nil, false, err
	}
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

func (d *Device) handleCommand(frame []byte) (Outcome, error) {
	command, decodeErr := DecodeRecord(frame)
	if decodeErr != nil {
		return Outcome{}, decodeErr
	}
	if command["message_type"] != "command" {
		return Outcome{}, fmt.Errorf("device: expected a command frame, got %v", command["message_type"])
	}
	return d.ApplyCommand(command), nil
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
	digest, digestErr := semanticCommandDigest(command)
	if digestErr != nil {
		return d.rejection(commandID, now, "malformed")
	}

	// Idempotency: a repeated key replays the prior outcome and applies no new
	// effect, regardless of faults.
	if prior, ok := d.dedup[idempotencyKey]; ok {
		if prior.digest != digest {
			return d.rejection(commandID, now, "duplicate")
		}
		return prior.outcome
	}

	if reject := d.admit(command, now); reject != "" {
		return d.rejection(commandID, now, reject)
	}
	d.acceptedCommands++
	ordinal := d.acceptedCommands
	injections := d.faultSchedule[ordinal]
	delete(d.faultSchedule, ordinal)
	fault := strings.Join(injections, ",")
	originalFaults := d.faults
	for _, injection := range injections {
		switch injection {
		case FaultAckLost:
			d.faults.AckLost = true
		case FaultStuck:
			d.faults.Stuck = true
		}
	}
	for _, injection := range injections {
		if injection != FaultStale && injection != FaultExpired {
			continue
		}
		code := "not_ready"
		if injection == FaultExpired {
			code = "expired"
		}
		outcome := d.rejection(commandID, now, code)
		outcome.Fault = fault
		outcome.AcceptedCommand = ordinal
		d.faults = originalFaults
		return outcome
	}

	outcome, accepted := d.applyAcceptedCommand(command, commandID, now)
	d.faults = originalFaults
	outcome.Fault = fault
	outcome.AcceptedCommand = ordinal
	if accepted {
		for _, injection := range injections {
			switch injection {
			case FaultDuplicate:
				outcome.Duplicate = true
			case FaultDisconnect:
				outcome.Disconnect = true
			case FaultReboot:
				d.rebootLocked("boot-reboot-" + strconv.Itoa(ordinal))
				// The response was created under the old boot identity. Do not
				// emit it after reboot; require the caller to observe new state.
				outcome.AckLost = true
			}
		}
		if !containsFault(injections, FaultReboot) {
			replay := outcome
			// Ack loss is a property of this wire delivery, not of the
			// idempotent execution. A later retry must be able to recover the
			// receipt without applying the plant a second time.
			replay.AckLost = false
			d.dedup[idempotencyKey] = dedupEntry{digest: digest, outcome: replay}
		}
	}
	return outcome
}

func (d *Device) applyAcceptedCommand(command map[string]any, commandID string, now int64) (Outcome, bool) {
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	params, _ := numericParams(command["parameters"])
	if operation == "safe_stop" {
		if !d.applySafeStop(target, now) {
			return d.rejection(commandID, now, "not_ready"), false
		}
		d.curTarget, d.curOp, d.curValue, d.energized = target, operation, 0, false
		d.safeState = true
		d.leaseUntilMicros = 0
		return Outcome{
			Receipt: map[string]any{
				"message_type": "receipt", "protocol_version": float64(ProtocolVersion),
				"command_id": commandID, "boot_id": d.bootID, "accepted": true,
				"received_mono_us": float64(now),
			},
			Result: map[string]any{
				"message_type": "result", "protocol_version": float64(ProtocolVersion),
				"command_id": commandID, "boot_id": d.bootID, "status": "safe_state",
				"completed_mono_us": float64(now),
			},
			AckLost: d.faults.AckLost,
		}, true
	}
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
	// Zero is the immediate-dispatch sentinel emitted by Agentic Stream.
	// Nonzero values are boot-relative freshness anchors.
	if notBefore > 0 {
		if float64(now) < notBefore {
			return "not_ready"
		}
		if float64(now) > notBefore+expiresAfter*1000 {
			return "expired"
		}
	}
	rawParams, validParams := strictParams(command["parameters"])
	if !validParams {
		return "out_of_range"
	}
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	if operation == "safe_stop" {
		if d.capabilities == nil || !d.capabilities.hasSafeStop(target) {
			return "wrong_target"
		}
		if len(rawParams) != 0 {
			return "out_of_range"
		}
		stop, _ := d.capabilities.safeStops[target]
		if stop.expiresAfterMS > 0 && expiresAfter > float64(stop.expiresAfterMS) {
			return "expired"
		}
		return ""
	}
	capa, ok := d.capabilities.target(target)
	if !ok {
		return "wrong_target"
	}
	if capa.ExpiresAfterMS > 0 && expiresAfter > float64(capa.ExpiresAfterMS) {
		return "expired"
	}
	if operation != capa.Operation {
		return "unknown_operation"
	}
	if len(rawParams) != len(capa.Bounds)+len(capa.StringValues) {
		return "out_of_range"
	}
	for name, bounds := range capa.Bounds {
		v, ok := rawParams[name].(float64)
		if !ok || v < bounds[0] || v > bounds[1] {
			return "out_of_range"
		}
	}
	for name, allowed := range capa.StringValues {
		value, ok := rawParams[name].(string)
		if !ok {
			return "out_of_range"
		}
		if _, ok := allowed[value]; !ok {
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
	safeStopSucceeded := d.invokeSafeStopLocked()
	if safeStopSucceeded {
		d.energized = false
		d.safeState = true
	} else {
		// A failed stop leaves the physical output unknown; do not claim the
		// device is safe or erase the last energized observation.
		d.energized = true
		d.safeState = false
	}
	d.leaseUntilMicros = 0
}

func (d *Device) invokeSafeStopLocked() bool {
	if d.curTarget == "" || d.capabilities == nil || !d.capabilities.hasSafeStop(d.curTarget) {
		return true
	}
	return d.applySafeStop(d.curTarget, d.clock())
}

func (d *Device) applySafeStop(target string, atMicros int64) bool {
	stopper, ok := d.plant.(SafeStopper)
	if !ok {
		return true
	}
	effect, err := stopper.SafeStop(target, atMicros)
	if err != nil {
		slog.Error("device safe stop failed", "target", target, "error", err)
		return false
	}
	if effect.Energized {
		slog.Error("device safe stop did not de-energize target", "target", target)
		return false
	}
	return true
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

func containsFault(faults []string, want string) bool {
	for _, fault := range faults {
		if fault == want {
			return true
		}
	}
	return false
}

func resultDetail(energized bool) string {
	if energized {
		return "output energized"
	}
	return "accepted, output not energized"
}

func strictParams(raw any) (map[string]any, bool) {
	out := map[string]any{}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}
	for name, v := range m {
		if name == "" {
			return nil, false
		}
		switch value := v.(type) {
		case float64:
			if !finite(value) {
				return nil, false
			}
		case string:
			if value == "" {
				return nil, false
			}
		default:
			return nil, false
		}
		out[name] = v
	}
	return out, true
}

func numericParams(raw any) (map[string]float64, bool) {
	params, ok := strictParams(raw)
	if !ok {
		return nil, false
	}
	out := map[string]float64{}
	for name, value := range params {
		number, ok := value.(float64)
		if ok {
			out[name] = number
		}
	}
	return out, true
}

func semanticCommandDigest(command map[string]any) (string, error) {
	identity := make(map[string]any, len(command))
	for key, value := range command {
		if key != "command_id" {
			identity[key] = value
		}
	}
	digest, err := canonical.DigestDomain("situation-runtime/device-command/v1\n", identity)
	if err != nil {
		return "", fmt.Errorf("canonicalize semantic command identity: %w", err)
	}
	return digest, nil
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

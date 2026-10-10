package device

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/uds"
)

// Config configures a device instance. Identity and firmware digest default to
// fixed deterministic values matching the vendored contract's golden state
// frame. When Capabilities is supplied, its canonical digest is authoritative.
type Config = layer.Config

// Faults is the injectable protocol fault state. The zero value is a healthy
// device; faults apply to the next accepted command until cleared.
type Faults = layer.Faults

// Outcome is the device's response to one command: the receipt, the terminal
// result and the delivery hints for the gateway loop.
type Outcome = layer.Outcome

// FaultInjection schedules one deterministic protocol fault by accepted
// command ordinal.
type FaultInjection = layer.FaultInjection

// Plant is the physical process a device actuates: the effect a command has
// on it is the ground truth, never the acknowledgement.
type Plant = layer.Plant

// SafeStopper is a Plant that can also be driven to its declared safe state.
type SafeStopper = layer.SafeStopper

// PlantCommand is one admitted command handed to the plant.
type PlantCommand = layer.PlantCommand

// PlantEffect is what the plant physically did.
type PlantEffect = layer.PlantEffect

// Capabilities is the device's data-defined capability catalog: the targets,
// operations and bounds it admits, with its canonical digest. Its methods are
// read-only lookups.
type Capabilities = layer.Capabilities

// WireFaults is a deterministic, scripted transport-fault plan applied to the
// device's outbound frames, indexed by emission order across a connection.
type WireFaults = uds.WireFaults

// Errors a plant returns and a transport reports.
var (
	// ErrPlantInterlocked identifies a plant refusal caused by an independent
	// safety system.
	ErrPlantInterlocked = layer.ErrPlantInterlocked
	// ErrPlantUnavailable identifies a plant that could not apply the command.
	ErrPlantUnavailable = layer.ErrPlantUnavailable
	// ErrInjectedDisconnect identifies a deterministic disconnect fault.
	ErrInjectedDisconnect = uds.ErrInjectedDisconnect
)

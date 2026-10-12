package device

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/uds"
)

type Config = layer.Config

type Outcome = layer.Outcome

type FaultInjection = layer.FaultInjection

type Plant = layer.Plant

type SafeStopper = layer.SafeStopper

type PlantCommand = layer.PlantCommand

type PlantEffect = layer.PlantEffect

type Capabilities = layer.Capabilities

type WireFaults = uds.WireFaults

var (
	ErrPlantInterlocked = layer.ErrPlantInterlocked

	ErrPlantUnavailable = layer.ErrPlantUnavailable

	ErrInjectedDisconnect = uds.ErrInjectedDisconnect
)

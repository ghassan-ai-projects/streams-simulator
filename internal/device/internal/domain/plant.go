package domain

import "errors"

var (
	ErrPlantInterlocked = errors.New("device plant interlocked")

	ErrPlantUnavailable = errors.New("device plant unavailable")
)

type Plant interface {
	Apply(PlantCommand) (PlantEffect, error)
}

type SafeStopper interface {
	SafeStop(target string, atMicros int64) (PlantEffect, error)
}

type PlantCommand struct {
	Target    string
	Operation string
	Params    map[string]float64
	CommandID string
	AtMicros  int64
}

type PlantEffect struct {
	Value     float64
	Energized bool
}

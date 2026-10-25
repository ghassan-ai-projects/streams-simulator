package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

const (
	FaultWorld               = "world"
	FaultMeasurement         = "measurement"
	FaultWorldAndMeasurement = "world_and_measurement"
)

func FaultCategories(spec *model.DomainSpec) map[string]map[string]string {
	categories := map[string]map[string]string{}
	for _, kind := range spec.Kinds {
		byFault := map[string]string{}
		for i := range kind.Faults {
			byFault[kind.Faults[i].ID] = faultCategory(&kind.Faults[i])
		}
		categories[kind.Name] = byFault
	}
	return categories
}

func faultCategory(fault *model.Fault) string {
	switch {
	case fault.Observation == nil:
		return FaultWorld
	case len(fault.Affects) == 0:
		return FaultMeasurement
	}
	return FaultWorldAndMeasurement
}

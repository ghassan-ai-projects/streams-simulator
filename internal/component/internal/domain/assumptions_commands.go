package domain

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func commandAssumptions(kind *model.Kind, overlay *twinOverlay) []CommandAssumption {
	commands := make([]CommandAssumption, 0, len(kind.Effectors))
	for i := range kind.Effectors {
		commands = append(commands, commandAssumption(kind.Name, &kind.Effectors[i], overlay))
	}
	return commands
}

func commandAssumption(kind string, effector *model.Effector, overlay *twinOverlay) CommandAssumption {
	command := declaredCommand(effector)
	if fact, found := overlay.command(kind, effector.Name); found {
		command.Response, command.RecordedIn, command.EqualsBinarySuggestion = fact.response, RecordedTwinChoices, fact.equals
		command.Drives = append(command.Drives, fact.drives...)
	}
	command.Statement = statementOf(command)
	return command
}

func declaredCommand(effector *model.Effector) CommandAssumption {
	command := CommandAssumption{Name: effector.Name, RiskClass: effector.RiskClass, SafeStop: effector.SafeStop,
		Arguments: argumentsOf(effector.ArgsSchema), Response: ResponseDeclared, RecordedIn: RecordedDomainFile, Sets: setsOf(effector.Effect),
		Drives: []DriveValue{}, Acknowledgement: acknowledgementOf(effector.Ack),
		Effect: EffectValue{DeadTimeS: effector.Effect.DeadTimeS, TimeConstantS: effector.Effect.TimeConstantS, StateChanges: len(effector.Effect.StateDeltas)}}
	if effector.Lease != nil {
		command.Lease = &LeaseValue{Argument: effector.Lease.Arg, Unit: effector.Lease.Unit, Fallback: effector.Lease.Fallback}
	}
	return command
}

func argumentsOf(schema map[string]any) []ArgumentValue {
	properties, _ := schema["properties"].(map[string]any)
	arguments := make([]ArgumentValue, 0, len(properties))
	for _, name := range slices.Sorted(maps.Keys(properties)) {
		property, _ := properties[name].(map[string]any)
		kind, _ := property["type"].(string)
		argument := ArgumentValue{Name: name, Type: kind}
		argument.Minimum, argument.Maximum = numberOf(property["minimum"]), numberOf(property["maximum"])
		arguments = append(arguments, argument)
	}
	return arguments
}

func numberOf(value any) *float64 {
	number, ok := jsonvalue.Number(value)
	if !ok {
		return nil
	}
	return &number
}

func setsOf(effect model.Effect) []SetValue {
	sets := []SetValue{}
	for _, delta := range effect.StateDeltas {
		if argument := argumentOf(delta); argument != "" {
			sets = append(sets, SetValue{Argument: argument, State: delta.State})
		}
	}
	return sets
}

func argumentOf(delta model.StateDelta) string {
	if delta.AssignFromArg != "" {
		return delta.AssignFromArg
	}
	return delta.FromArg
}

func acknowledgementOf(ack model.Ack) AcknowledgementValue {
	value := AcknowledgementValue{LatencyMean: ack.LatencyMS.Mean, LatencySigma: ack.LatencyMS.Sigma, FailureModes: []FailureModeValue{}}
	for _, mode := range ack.FailureModes {
		value.FailureModes = append(value.FailureModes, FailureModeValue{Mode: mode.Mode, Probability: mode.Probability})
	}
	return value
}

func statementOf(command CommandAssumption) string {
	switch command.Response {
	case ResponseFixedSafeStop:
		return "Returns every commanded reading to its minimum at once; the acknowledgement does not wait for a device."
	case ResponseNone:
		return "Is accepted and acknowledged; it changes nothing in the twin."
	case ResponseDirect:
		return "Sets " + argumentList(command) + " to the commanded value at once."
	case ResponseFirstOrder:
		return "Sets " + argumentList(command) + ", then " + driveList(command.Drives) + "."
	}
	return fmt.Sprintf("Applies %d state change(s) after a dead time of %g s and a time constant of %g s.",
		command.Effect.StateChanges, command.Effect.DeadTimeS, command.Effect.TimeConstantS)
}

func argumentList(command CommandAssumption) string {
	names := make([]string, len(command.Sets))
	for i, set := range command.Sets {
		names[i] = set.Argument
	}
	return strings.Join(names, ", ")
}

func driveList(drives []DriveValue) string {
	parts := make([]string, len(drives))
	for i, drive := range drives {
		parts[i] = fmt.Sprintf("%s.%s follows with gain %g and a time constant of %g s (first-order lag)", drive.Kind, drive.Metric, drive.Gain, drive.TimeConstantS)
	}
	return strings.Join(parts, "; ")
}

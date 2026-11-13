package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrUnknownOutcomeMode = errors.New("run: unknown outcome mode")

func OutcomeCommands(log []model.Command, mode model.OutcomeMode) ([]model.Command, error) {
	if mode == model.OutcomeAsRun {
		return log, nil
	}
	if _, policy := mode.Policy(); mode != model.OutcomeDoNothing && !policy {
		return nil, fmt.Errorf("%w %q: want %s, %s or policy:<name>", ErrUnknownOutcomeMode, mode, model.OutcomeAsRun, model.OutcomeDoNothing)
	}
	kept := make([]model.Command, 0, len(log))
	for _, command := range log {
		if command.Op != model.OpEffectorInvoke {
			kept = append(kept, command)
		}
	}
	return kept, nil
}

package files

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func commandAt(atNS int64) model.Command {
	return model.Command{AtNS: atNS, Op: model.OpClockAdvance, Args: map[string]any{"to_ns": atNS}}
}

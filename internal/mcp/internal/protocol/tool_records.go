package protocol

import (
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func decodeToolRecord(raw map[string]any, out any) error {
	data, err := json.Marshal(raw)
	if err != nil {
		return director.ToolErrorf(director.CodeInvalidArgs, "%v", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return director.ToolErrorf(director.CodeInvalidArgs, "%v", err)
	}
	return nil
}

func consumerVerdictArgument(args map[string]any) (*model.Verdict, error) {
	raw, ok := args["verdict"].(map[string]any)
	if !ok {
		return nil, nil
	}
	verdict := &model.Verdict{}
	if err := decodeToolRecord(raw, verdict); err != nil {
		return nil, err
	}
	return verdict, nil
}

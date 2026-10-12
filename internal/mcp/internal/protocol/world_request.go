package protocol

import "github.com/ghassan-ai-projects/streams-simulator/internal/director"

func worldRequest(args map[string]any) director.WorldRequest {
	return director.WorldRequest{
		Domain: stringArg(args, "domain"), Adapter: stringArg(args, "adapter"),
		Seed: numberArg(args, "seed", 1), Sink: stringArg(args, "sink"),
		SinkTarget: stringArg(args, "sink_target"), TimeMode: stringArg(args, "time_mode"),
		StartTimeNS: worldStartTime(args), Entities: worldEntityArguments(args),
		ScenarioProfile: stringArg(args, "scenario_profile"), Label: stringArg(args, "label"),
		Explore: boolArg(args, "explore"),
	}
}

func worldStartTime(args map[string]any) *int64 {
	if _, present := args["start_time"]; !present {
		return nil
	}
	start := numberArg(args, "start_time", 0)
	return &start
}

func worldEntityArguments(args map[string]any) []string {
	var entities []string
	if items, ok := args["entities"].([]any); ok {
		for _, item := range items {
			if entity, ok := item.(string); ok {
				entities = append(entities, entity)
			}
		}
	}
	return entities
}

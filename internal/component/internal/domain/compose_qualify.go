package domain

func qualifyFaults(kind map[string]any, name string) {
	for _, fault := range members(kind, "faults") {
		fault["id"] = name + "__" + stringOf(fault["id"])
		qualifyMember(mapOf(mapOf(fault["observability"])["detector"]), name, detectorChannels...)
		qualifyMember(mapOf(fault["observation"]), name, "channel")
		if effector, named := fault["expected_effector"].(string); named {
			fault["expected_effector"] = name + "__" + effector
		}
	}
}

func qualifyMember(owner map[string]any, name string, members ...string) {
	for _, member := range members {
		if channel, named := owner[member].(string); named {
			owner[member] = name + "." + channel
		}
	}
}

func qualifyEffectors(kind map[string]any, name string) {
	for _, effector := range members(kind, "effectors") {
		effector["name"] = name + "__" + stringOf(effector["name"])
		channels, _ := effector["confirmation_channels"].([]any)
		for i, channel := range channels {
			channels[i] = name + "." + stringOf(channel)
		}
	}
}

func stripDescriptions(kind map[string]any) {
	for _, section := range []string{"state", "dynamics", "flows", "faults", "effectors", "boundaries"} {
		for _, item := range members(kind, section) {
			delete(item, "description")
		}
	}
}

func mapOf(value any) map[string]any {
	out, _ := value.(map[string]any)
	return out
}

func stringOf(value any) string {
	out, _ := value.(string)
	return out
}

package domain

var detectorChannels = []string{"channel", "channel_a", "channel_b"}

type pointReference struct {
	name string
	at   string
	what string
}

func pointReferences(kind map[string]any) []pointReference {
	var references []pointReference
	for i, fault := range members(kind, "faults") {
		at := pointer("kind", "faults", indexPart(i))
		references = append(references, faultReferences(fault, at)...)
	}
	for i, effector := range members(kind, "effectors") {
		at := pointer("kind", "effectors", indexPart(i))
		references = append(references, confirmationReferences(effector, at)...)
	}
	return references
}

func faultReferences(fault map[string]any, at string) []pointReference {
	references := detectorReferences(fault, at)
	observation := mapOf(fault["observation"])
	if name, named := observation["channel"].(string); named {
		references = append(references, pointReference{name, at + "/observation/channel", "the observation of fault " + stringOf(fault["id"])})
	}
	return references
}

func detectorReferences(fault map[string]any, at string) []pointReference {
	var references []pointReference
	detector := mapOf(mapOf(fault["observability"])["detector"])
	for _, member := range detectorChannels {
		if name, named := detector[member].(string); named {
			references = append(references, pointReference{name, at + "/observability/detector/" + member, "the detector of fault " + stringOf(fault["id"])})
		}
	}
	return references
}

func confirmationReferences(effector map[string]any, at string) []pointReference {
	var references []pointReference
	name, _ := effector["name"].(string)
	names, _ := effector["confirmation_channels"].([]any)
	for i, raw := range names {
		channel, _ := raw.(string)
		references = append(references, pointReference{channel, at + "/confirmation_channels/" + indexPart(i), "a confirmation of effector " + name})
	}
	return references
}

func checkReferences(kind map[string]any, index *kindIndex, issues *issueList) {
	for _, reference := range pointReferences(kind) {
		if !index.points[reference.name] {
			issues.addf(reference.at, "%s names %q, which is not a measurement point", reference.what, reference.name)
		}
	}
	for i, fault := range index.faults {
		if name, named := fault["expected_effector"].(string); named && !index.effectors[name] {
			issues.addf(pointer("kind", "faults", indexPart(i), "expected_effector"), "expected effector %q is not an effector of the kind", name)
		}
	}
}

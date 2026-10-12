package app

import (
	"encoding/json"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const distinctiveLength = 16

func hiddenFacts(spec *model.DomainSpec) []string {
	public := publicText(spec)
	facts := distinctive(spec.Description, spec.Stresses)
	for i := range spec.Kinds {
		kind := &spec.Kinds[i]
		for _, state := range kind.State {
			if quoted := `"` + state.Name + `"`; !strings.Contains(public, quoted) {
				facts = append(facts, quoted)
			}
			facts = append(facts, distinctive(state.Description)...)
		}
		facts = append(facts, kindDescriptions(kind)...)
	}
	return facts
}

func publicText(spec *model.DomainSpec) string {
	public := []any{"virtual", "running"}
	for i := range spec.Kinds {
		kind := &spec.Kinds[i]
		public = append(public, kind.Name)
		for _, channel := range kind.Channels {
			public = append(public, channel.Name, channel.Unit, channel.EnumValues)
		}
		for _, effector := range kind.Effectors {
			public = append(public, effector.Name, effector.ArgsSchema)
		}
	}
	raw, _ := json.Marshal(public)
	return string(raw)
}

func kindDescriptions(kind *model.Kind) []string {
	var facts []string
	for _, channel := range kind.Channels {
		facts = append(facts, distinctive(channel.Description)...)
	}
	for _, fault := range kind.Faults {
		facts = append(facts, distinctive(fault.Description)...)
	}
	for _, effector := range kind.Effectors {
		facts = append(facts, distinctive(effector.Description)...)
		facts = append(facts, annotations(effector.ArgsSchema)...)
	}
	return facts
}

func annotations(schema any) []string {
	var facts []string
	switch typed := schema.(type) {
	case map[string]any:
		for key, value := range typed {
			if text, isText := value.(string); isText && (key == "description" || key == "title" || key == "$comment") {
				facts = append(facts, distinctive(text)...)
			}
			facts = append(facts, annotations(value)...)
		}
	case []any:
		for _, item := range typed {
			facts = append(facts, annotations(item)...)
		}
	}
	return facts
}

func distinctive(texts ...string) []string {
	var facts []string
	for _, text := range texts {
		if len(text) >= distinctiveLength {
			facts = append(facts, text)
		}
	}
	return facts
}

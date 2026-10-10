package app

import "encoding/json"

type toolInput struct {
	Properties map[string]any
	Required   []string
	OneOf      []any
	AllOf      []any
}

// toolSchema returns the same public input contract consumed by the handlers.
func toolSchema(name string) json.RawMessage {
	factory, known := toolInputFactories[name]
	if !known {
		panic("mcp: schema is missing for tool " + name)
	}
	return marshalToolInput(factory())
}

var toolInputFactories = map[string]func() toolInput{
	"sim.catalog.list":      catalogListInput,
	"sim.catalog.describe":  catalogDescribeInput,
	"sim.catalog.coverage":  emptyToolInput,
	"sim.adapter.list":      emptyToolInput,
	"sim.world.create":      worldCreateInput,
	"sim.world.describe":    worldIDInput,
	"sim.world.destroy":     worldIDInput,
	"sim.clock.state":       worldIDInput,
	"sim.clock.advance":     clockAdvanceInput,
	"sim.fault.inject":      faultInjectInput,
	"sim.fault.clear":       faultClearInput,
	"sim.fault.list":        worldIDInput,
	"sim.perturb.apply":     perturbApplyInput,
	"sim.perturb.clear":     perturbClearInput,
	"sim.env.inject":        envInjectInput,
	"sim.run.begin":         runBeginInput,
	"sim.run.end":           worldIDInput,
	"sim.truth.seal":        truthSealInput,
	"sim.run.verify":        runVerifyInput,
	"sim.truth.reveal":      truthRevealInput,
	"sim.truth.seal_status": runIDInput,
	"sim.score":             runIDInput,
	"sim.scenario.audit":    scenarioAuditInput,
	"sim.entity.retire":     entityRetireInput,
	"sim.nameplate.read":    tokenInput,
	"sim.effector.list":     tokenInput,
	"sim.effector.invoke":   effectorInvokeInput,
	"sim.consumer.report":   consumerReportInput,
}

func marshalToolInput(input toolInput) json.RawMessage {
	doc := toolInputDocument(input)
	raw, err := json.Marshal(doc)
	if err != nil {
		panic("mcp: marshal tool schema: " + err.Error())
	}
	return raw
}

func toolInputDocument(input toolInput) map[string]any {
	required := input.Required
	if required == nil {
		required = []string{}
	}
	doc := map[string]any{"type": "object", "properties": input.Properties, "required": required, "additionalProperties": false}
	if len(input.OneOf) > 0 {
		doc["oneOf"] = input.OneOf
	}
	if len(input.AllOf) > 0 {
		doc["allOf"] = input.AllOf
	}
	return doc
}

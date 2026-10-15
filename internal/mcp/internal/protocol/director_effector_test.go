package protocol

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestDirectorListsAndInvokesAWorldsEffectorsWithoutTheCapabilityToken(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	world := newWorld(t, d)
	cs, _ := connect(t, NewDirectorServer(d))
	listed := mustCall(t, cs, "sim.world.effectors", map[string]any{"world_id": world.id})
	effectors, _ := listed["effectors"].([]any)
	if len(effectors) == 0 || listed["world_id"] != world.id {
		t.Fatalf("effectors = %v", listed)
	}
	first, _ := effectors[0].(map[string]any)
	schema, _ := first["args_schema"].(map[string]any)
	ack := mustCall(t, cs, "sim.world.invoke", map[string]any{
		"world_id": world.id, "entity_id": world.entity, "effector": first["name"],
		"args": testsupport.ArgsForSchema(schema, world.entity),
	})
	if ack["accepted"] != true || ack["command_id"] != "director-1" || ack["world_id"] == "" {
		t.Fatalf("ack = %v", ack)
	}
}

func TestDirectorEffectorToolsRefuseUnknownWorldsAndEffectors(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	world := newWorld(t, d)
	cs, _ := connect(t, NewDirectorServer(d))
	if _, refusal := toolCall(t, cs, "sim.world.effectors", map[string]any{"world_id": "w-none"}); !strings.HasPrefix(refusal, "world_not_found: ") {
		t.Fatalf("unknown world refusal = %q", refusal)
	}
	_, refusal := toolCall(t, cs, "sim.world.invoke", map[string]any{"world_id": world.id, "entity_id": world.entity, "effector": "no_such"})
	if !strings.HasPrefix(refusal, "unknown_effector: ") {
		t.Fatalf("unknown effector refusal = %q", refusal)
	}
}

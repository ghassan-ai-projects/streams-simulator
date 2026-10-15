package app

import (
	"errors"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestWorldEffectorsAreTheNameplateEffectorsOfTheWorld(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	listed, err := d.WorldEffectors(worldID)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := listed["effectors"].([]rules.EffectorInfo)
	if want := d.Worlds[worldID].Nameplate.Effectors; len(got) == 0 || len(got) != len(want) || got[0].Name != want[0].Name {
		t.Fatalf("effectors = %+v, want %+v", got, want)
	}
}

func TestInvokeWorldEffectorNumbersTheCommandsItNamesItself(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	entity := d.Worlds[worldID].Run.World.EntityIDs()[0]
	effector := d.Worlds[worldID].Nameplate.Effectors[0]
	args := testsupport.ArgsForSchema(effector.ArgsSchema, entity)
	first, err := d.InvokeWorldEffector(worldID, effector.Name, entity, "", args)
	if err != nil || first.CommandID != "director-1" || !first.Accepted {
		t.Fatalf("first = %+v, %v", first, err)
	}
	second, err := d.InvokeWorldEffector(worldID, effector.Name, entity, "mine", args)
	if err != nil || second.CommandID != "mine" {
		t.Fatalf("second = %+v, %v", second, err)
	}
	third, err := d.InvokeWorldEffector(worldID, effector.Name, entity, "", args)
	if err != nil || third.CommandID != "director-2" {
		t.Fatalf("third = %+v, %v", third, err)
	}
}

func TestWorldEffectorUseCasesRefuseAnUnknownWorld(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	var tool *rules.ToolError
	if _, err := d.WorldEffectors("w-none"); !errors.As(err, &tool) || tool.Code != rules.CodeWorldNotFound {
		t.Fatalf("list: %v", err)
	}
	if _, err := d.InvokeWorldEffector("w-none", "e", "x", "", nil); !errors.As(err, &tool) || tool.Code != rules.CodeWorldNotFound {
		t.Fatalf("invoke: %v", err)
	}
}

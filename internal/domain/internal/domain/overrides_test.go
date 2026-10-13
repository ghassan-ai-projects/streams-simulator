package domain

import (
	"testing"
)

func TestAnOverrideBuildsTheEffectiveKindOfItsSetAndLeavesTheKindAlone(t *testing.T) {
	t.Parallel()
	c := loadStructured(t)
	plain, named := c.InstanceKind(0), c.InstanceKind(1)
	if plain != c.Kind("tank") || named == plain || named.Name != "tank" {
		t.Fatal("a set without an override shares its kind; a set with one has its own effective kind")
	}
	if c.Kind("tank").StateDecl("temp") != nil || c.Kind("tank").ChannelDecl("tank.temp") != nil || len(c.Kind("tank").Effectors) != 1 {
		t.Fatal("the override changed the declared kind")
	}
	if named.StateDecl("temp") == nil || named.DynamicsOf("temp") == nil || named.ChannelDecl("tank.temp") == nil || named.FaultDecl("tank_overheat") == nil || named.EffectorDecl("chill") == nil {
		t.Fatal("the effective kind lacks an added declaration")
	}
	if level := named.StateDecl("level_true"); *level.Min != 0.5 || *level.Max != 4 || c.Kind("tank").StateDecl("level_true").Max != nil {
		t.Fatalf("narrowed level bounds = [%v, %v]", *level.Min, *level.Max)
	}
}

func TestOverrideDeclarationsResolveDomainWide(t *testing.T) {
	t.Parallel()
	c := loadStructured(t)
	if c.Fault("tank_overheat") == nil || c.FaultKind("tank_overheat") != c.InstanceKind(1) || c.Channel("tank.temp") == nil || c.Effector("chill") == nil {
		t.Fatal("an added declaration did not resolve by its domain-wide name")
	}
	if !c.HasChannel("tank.temp") || !c.HasFault("tank_overheat") || !c.HasEffector("chill") || c.FaultCount() != 6 || c.EffectorCount() != 3 {
		t.Fatalf("symbols: faults %d effectors %d", c.FaultCount(), c.EffectorCount())
	}
	if names := c.KindChannelNames("tank"); len(names) != 2 || names[1] != "tank.temp" {
		t.Fatalf("tank channel names = %v", names)
	}
	if channels := c.DeclaredChannels(); len(channels) != 4 || len(c.DeclaredEffectors()) != 3 {
		t.Fatalf("declared channels %d effectors %d", len(channels), len(c.DeclaredEffectors()))
	}
}

func TestNarrowedBoundsApplyToTheEntityNotTheKind(t *testing.T) {
	t.Parallel()
	c := loadStructured(t)
	if c.EntityChannel("P-slow", "pump.flow").Range.Max != 12 || c.EntityChannel("pump-1", "pump.flow").Range.Max != 20 {
		t.Fatal("the channel range narrowed for the wrong entities")
	}
	if c.EntityChannel("T-main", "tank.temp") == nil || c.EntityChannel("nobody", "tank.level") == nil {
		t.Fatal("entity channel lookup lost a declaration")
	}
	speed := func(set int) float64 {
		schema := c.InstanceKind(set).EffectorDecl("boost").ArgsSchema
		return at(schema, "properties", "speed", "maximum").(float64)
	}
	if speed(3) != 50 || speed(2) != 100 {
		t.Fatalf("boost speed maximum = %v for P-slow and %v for the kind", speed(3), speed(2))
	}
}

func TestOverridesThatAddAreCheckedAtTheirPointers(t *testing.T) {
	t.Parallel()
	runViolations(t, []violationCase{
		{"an added channel redeclares a kind channel", func(d map[string]any) {
			put(d, "tank.level", "instances", 1, "override", "channels", 0, "name")
		}, "/instances/1/override/channels/0/name"},
		{"an added fault redeclares a kind fault", func(d map[string]any) {
			put(d, "tank_leak", "instances", 1, "override", "faults", 0, "id")
		}, "/instances/1/override/faults/0/id"},
		{"an added effector redeclares a kind effector", func(d map[string]any) {
			put(d, "refill", "instances", 1, "override", "effectors", 0, "name")
		}, "/instances/1/override/effectors/0/name"},
		{"an added state redeclares a kind state", func(d map[string]any) {
			put(d, "level_true", "instances", 1, "override", "state", 0, "name")
		}, "/instances/1/override/state/0/name"},
		{"added dynamics retype a kind state", func(d map[string]any) {
			put(d, "level_true", "instances", 1, "override", "dynamics", 0, "target")
		}, "/instances/1/override/dynamics/0/target"},
		{"an added channel observes an undeclared state", func(d map[string]any) {
			put(d, "nothing", "instances", 1, "override", "channels", 0, "observes")
		}, "/instances/1/override/channels/0/observes"},
		{"an added fault names a channel of another kind", func(d map[string]any) {
			put(d, "pump.flow", "instances", 1, "override", "faults", 0, "observability", "detector", "channel")
		}, "/instances/1/override/faults/0/observability/detector/channel"},
		{"added dynamics form a cycle", func(d map[string]any) {
			put(d, []any{map[string]any{"state": "temp"}}, "instances", 1, "override", "dynamics", 0, "f1", "inputs")
		}, "/instances/1/override/dynamics/0/target"},
		{"an effector delta targets an undeclared state", func(d map[string]any) {
			put(d, "nothing", "instances", 1, "override", "effectors", 0, "effect", "state_deltas", 0, "state")
		}, "/instances/1/override/effectors/0/effect/state_deltas/0/state"},
	})
}

func TestOverridesCannotRemoveOrRetype(t *testing.T) {
	t.Parallel()
	runViolations(t, []violationCase{
		{"a remove member", func(d map[string]any) {
			put(d, []any{"tank.level"}, "instances", 1, "override", "remove")
		}, "/instances/1/override/remove"},
		{"a unit change in a narrowing", func(d map[string]any) {
			put(d, "K", "instances", 1, "override", "narrow", "state", 0, "unit")
		}, "/instances/1/override/narrow/state/0/unit"},
		{"a value type change in a channel narrowing", func(d map[string]any) {
			put(d, "string", "instances", 3, "override", "narrow", "channels", 0, "value_type")
		}, "/instances/3/override/narrow/channels/0/value_type"},
	})
}

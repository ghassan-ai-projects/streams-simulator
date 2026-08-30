package deviceworld

import (
	"os"
	"testing"
)

func TestLoadBindingsFromData(t *testing.T) {
	data, err := os.ReadFile("testdata/thermal.bindings.json")
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := LoadBindings(data, "entity-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 1 {
		t.Fatalf("loaded %d bindings, want one", len(bindings))
	}
	for target, binding := range bindings {
		if binding.effector == "" || binding.valueState == "" {
			t.Fatalf("binding %q is incomplete: %+v", target, binding)
		}
		args, err := binding.args(map[string]float64{})
		if err != nil {
			t.Fatalf("build binding arguments: %v", err)
		}
		if args["reefer_id"] != "entity-01" || args["setpoint_c"] != float64(-20) {
			t.Fatalf("binding arguments = %#v, want runtime entity and catalog value", args)
		}
	}
}

func TestLoadBindingsFailsClosed(t *testing.T) {
	cases := map[string]string{
		"unknown field":     `{"bindings":{"target":{"effector":"effect","unexpected":true}}}`,
		"trailing json":     `{"bindings":{"target":{"effector":"effect"}}} {}`,
		"unknown source":    `{"bindings":{"target":{"effector":"effect","arguments":{"arg":{"source":"unknown"}}}}}`,
		"missing parameter": `{"bindings":{"target":{"effector":"effect","arguments":{"arg":{"source":"parameter"}}}}}`,
		"entity with value": `{"bindings":{"target":{"effector":"effect","arguments":{"arg":{"source":"entity","value":"x"}}}}}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadBindings([]byte(body), "entity-01"); err == nil {
				t.Fatal("invalid binding catalog was accepted")
			}
		})
	}
}

func TestLoadBindingsRequiresEntitySourceValue(t *testing.T) {
	data, err := os.ReadFile("testdata/thermal.bindings.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadBindings(data, ""); err == nil {
		t.Fatal("entity-sourced binding must require a runtime entity")
	}
}

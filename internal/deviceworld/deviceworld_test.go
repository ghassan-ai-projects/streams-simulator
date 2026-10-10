package deviceworld_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func coldChain(t *testing.T) (*world.World, string) {
	t.Helper()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "cold-chain-transit.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.New(spec, 1, "w-dw", model.DefaultStartTimeNS, world.Options{EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return w, w.EntityIDs()[0]
}

func thermalBindings(t *testing.T, entity string) map[string]deviceworld.Binding {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "thermal.bindings.json"))
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := deviceworld.LoadBindings(data, entity)
	if err != nil {
		t.Fatal(err)
	}
	return bindings
}

func TestNewRefusesAMissingWorld(t *testing.T) {
	t.Parallel()
	if _, err := deviceworld.New(nil, nil); !errors.Is(err, deviceworld.ErrNoWorld) {
		t.Fatalf("err = %v, want ErrNoWorld", err)
	}
}

func TestLoadBindingsRefusesAnUnknownField(t *testing.T) {
	t.Parallel()
	if _, err := deviceworld.LoadBindings([]byte(`{"no_such_field": 1}`), "e-1"); err == nil {
		t.Fatal("a catalog with an undeclared field must be refused")
	}
}

func TestBindingsValidateAgainstTheirWorldAndNamedTargetsMustExist(t *testing.T) {
	t.Parallel()
	w, entity := coldChain(t)
	bindings := thermalBindings(t, entity)
	var targets []string
	for target := range bindings {
		targets = append(targets, target)
	}
	if err := deviceworld.ValidateBindings(w, bindings, targets, nil); err != nil {
		t.Fatal(err)
	}
	err := deviceworld.ValidateBindings(w, bindings, []string{"no-such-target"}, nil)
	if err == nil || !strings.Contains(err.Error(), "no-such-target") {
		t.Fatalf("missing required target: %v", err)
	}
}

func TestPlantRefusesAnUnmappedTarget(t *testing.T) {
	t.Parallel()
	w, entity := coldChain(t)
	plant, err := deviceworld.New(w, thermalBindings(t, entity))
	if err != nil {
		t.Fatal(err)
	}
	_, err = plant.Apply(device.PlantCommand{Target: "no-such-target", Operation: "set", CommandID: "c-1"})
	if !errors.Is(err, device.ErrPlantUnavailable) {
		t.Fatalf("err = %v, want ErrPlantUnavailable", err)
	}
}

package mcp

import (
	"context"
	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
	"strings"
	"testing"
)

// TestOperatorEndpointPerShippedDomain: every shipped domain can be driven
// through the operator endpoint — nameplate, effector list, an invocation
// with arguments derived from the declared schema, quiescence and a verdict.
func TestOperatorEndpointPerShippedDomain(t *testing.T) {
	specs, err := domain.LoadAll("../../domains")
	if err != nil {
		t.Fatal(err)
	}
	adap, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	d := NewDirector(context.Background(), domain.NewCatalog(specs), map[string]*model.Adapter{"native-jsonl": adap}, t.TempDir())
	endpoint := startOperatorEndpoint(t, d)

	for _, spec := range specs {
		id := spec.Spec.ID
		t.Run(id, func(t *testing.T) {
			created, err := d.CreateWorld(map[string]any{
				"domain": id, "seed": float64(7), "adapter": "native-jsonl",
				"sink": model.SinkInproc, "time_mode": model.TimeStepped,
			})
			if err != nil {
				t.Fatalf("world.create: %v", err)
			}
			worldID := created["world_id"].(string)
			w := d.Worlds[worldID]
			op, err := refconsumer.NewMCPOperator(endpoint, w.Token, w.Run.ID)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = op.Close() }()

			np, err := op.Nameplate()
			if err != nil {
				t.Fatal(err)
			}
			if len(np.Entities) == 0 || len(np.Channels) == 0 {
				t.Fatalf("nameplate incomplete for %s: %+v", id, np)
			}
			effs, err := op.ListEffectors()
			if err != nil {
				t.Fatal(err)
			}
			entity := np.Entities[0].ID
			if len(effs) > 0 {
				res, err := op.InvokeEffector(effs[0].Name, entity, "e2e-"+id, argsForSchema(effs[0].Schema, entity), 0)
				if err != nil {
					t.Fatalf("invoke %s on %s: %v", effs[0].Name, id, err)
				}
				if !res.Accepted {
					t.Fatalf("invoke %s on %s not accepted: %+v", effs[0].Name, id, res)
				}
			}
			if err := op.ReportQuiesced(w.Run.World.Clock()); err != nil {
				t.Fatalf("quiescence report: %v", err)
			}
			if err := op.SubmitVerdict(&model.Verdict{
				SchemaVersion: "0.1", RunID: w.Run.ID,
				Consumer: model.ConsumerInfo{Name: "e2e", Version: "0.0.1"},
			}); err != nil {
				t.Fatalf("verdict submission: %v", err)
			}
		})
	}
}

// argsForSchema synthesizes arguments for an effector from its declared
// argument schema: required *_id strings bind to the entity, numbers to 1.
func argsForSchema(schema map[string]any, entity string) map[string]any {
	out := map[string]any{}
	if schema == nil {
		return out
	}
	props, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]any)
	for _, r := range required {
		name, _ := r.(string)
		if name == "" {
			continue
		}
		p, _ := props[name].(map[string]any)
		switch p["type"] {
		case "string":
			if strings.HasSuffix(name, "_id") {
				out[name] = entity
			} else {
				out[name] = "x"
			}
		case "number", "integer":
			out[name] = 1
		case "boolean":
			out[name] = true
		default:
			out[name] = ""
		}
	}
	return out
}

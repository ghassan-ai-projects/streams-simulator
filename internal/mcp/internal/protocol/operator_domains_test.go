package protocol

import (
	"context"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// TestOperatorEndpointPerShippedDomain: every shipped domain can be driven
// through the operator endpoint — nameplate, effector list, an invocation
// with arguments derived from the declared schema, quiescence and a verdict.
func TestOperatorEndpointPerShippedDomain(t *testing.T) {
	t.Parallel()
	specs, err := domain.LoadAll(testsupport.DomainsDir())
	if err != nil {
		t.Fatal(err)
	}
	adap, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	d := director.New(context.Background(), domain.NewCatalog(specs), map[string]*model.Adapter{"native-jsonl": adap}, t.TempDir())
	endpoint := startOperatorEndpoint(t, d)

	for _, spec := range specs {
		id := spec.Spec.ID
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			created, err := d.CreateWorld(director.WorldRequest{
				Domain: id, Seed: 7, Adapter: "native-jsonl",
				Sink: model.SinkInproc, TimeMode: model.TimeStepped,
			})
			if err != nil {
				t.Fatalf("world.create: %v", err)
			}
			token, runID := created["token"].(string), created["run_id"].(string)
			op, err := refconsumer.NewMCPOperator(endpoint, token, runID)
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
			if err := op.ReportQuiesced(model.DefaultStartTimeNS); err != nil {
				t.Fatalf("quiescence report: %v", err)
			}
			if err := op.SubmitVerdict(&model.Verdict{
				SchemaVersion: "0.1", RunID: runID,
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

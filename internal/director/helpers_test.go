package director_test

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// newTestDirector builds a director with the aquaculture-pond domain and the
// native-jsonl adapter.
func newTestDirector(t *testing.T) *director.Director {
	t.Helper()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	cat := domain.NewCatalog([]*domain.Compiled{spec})
	return director.New(t.Context(), cat, map[string]*model.Adapter{"native-jsonl": native}, t.TempDir())
}

// argsForSchema synthesizes arguments for an effector from its declared
// argument schema: required *_id strings bind to the entity, numbers to 1.
func argsForSchema(schema map[string]any, entity string) map[string]any {
	out := map[string]any{}
	props, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]any)
	for _, r := range required {
		name, _ := r.(string)
		p, _ := props[name].(map[string]any)
		switch p["type"] {
		case "string":
			out[name] = "x"
			if strings.HasSuffix(name, "_id") {
				out[name] = entity
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

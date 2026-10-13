package director_test

import (
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

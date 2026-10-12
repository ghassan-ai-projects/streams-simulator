package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestLookupFindsAKnownAdapterAndRefusesAnUnknownID(t *testing.T) {
	t.Parallel()
	native := &model.Adapter{ID: "native"}
	adapters := map[string]*model.Adapter{"native": native}
	if got, err := Lookup(adapters, "native"); err != nil || got != native {
		t.Fatalf("Lookup(native) = %v, %v", got, err)
	}
	if _, err := Lookup(adapters, "no-such"); !errors.Is(err, ErrUnknownAdapter) || err.Error() != `unknown adapter "no-such"` {
		t.Fatalf("Lookup(no-such) = %v, want ErrUnknownAdapter naming the id", err)
	}
}

package perturb

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func TestComposedPerturbationStreamRegression(t *testing.T) {
	t.Parallel()
	layer := New("w", 71, testSpec(t))
	for _, tc := range []struct {
		name   string
		params map[string]any
	}{
		{DuplicateBurst, map[string]any{"rate": 0.4}},
		{IDReuse, map[string]any{"rate": 0.3}},
		{DelayTail, map[string]any{"mean_s": 30.0, "sigma_s": 10.0}},
		{ClockSkew, map[string]any{"offset_s": 17.0}},
		{Drop, map[string]any{"rate": 0.2}},
	} {
		if _, err := layer.Apply(tc.name, tc.params, 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	var records []Delivered
	for seq := int64(0); seq < 100; seq++ {
		records = append(records, layer.Process(ev(seq, "num", float64(seq)), 1e9+seq*1e9)...)
	}
	// Private delivery records are characterized here, not exposed as a wire contract.
	raw, err := json.Marshal(records) //nolint:musttag
	if err != nil {
		t.Fatal(err)
	}
	const want = "sha256:d34d9a91693b25eb5699be5b627c4733c67584680428ae9c87e4168b8c7dd0a5"
	if got := canonical.DigestBytes(raw); got != want {
		t.Fatalf("composed stream changed: %s", got)
	}
}

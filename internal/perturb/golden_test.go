package perturb

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// goldenParams makes every perturbation fire often enough that a changed RNG
// draw order, application order or record shape shows in the stream digest.
var goldenParams = map[string]map[string]any{
	Drop:           {"rate": 0.5},
	DuplicateBurst: {"rate": 0.5},
	IDReuse:        {"rate": 0.5},
	OutOfEnum:      {"rate": 0.5},
	OutOfRange:     {"rate": 0.5, "magnitude": 10.0},
	Reorder:        {"max_displacement": int64(3)},
	DelayTail:      {"mean_s": 30.0, "sigma_s": 60.0},
	ClockSkew:      {"offset_s": 5.0, "sign": "negative"},
	Oversize:       {"bytes": int64(64)},
	Storm:          {"multiplier": int64(3)},
}

// goldenRecord is Delivered with explicit JSON names (the Go field names, so
// the recorded digests are the plain encoding of Delivered).
type goldenRecord struct {
	DeliveryID uint64         `json:"DeliveryID"`
	Event      model.SimEvent `json:"Event"`
	Reason     string         `json:"Reason"`
	Delivered  bool           `json:"Delivered"`
	Malformed  bool           `json:"Malformed"`
}

// goldenStream runs a mixed 60-event stream through a layer, flushing at
// every 15th event and at the end, and returns the digest of everything the
// observer would have received, in order.
func goldenStream(t *testing.T, names []string) string {
	t.Helper()
	l := New("w", 42, testSpec(t))
	for _, name := range names {
		var until int64
		if name == GrossBackfill {
			until = 1000000000 + 30*1000000000 // withhold the first 30 s
		}
		if _, err := l.Apply(name, goldenParams[name], 0, until); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
	var out []Delivered
	for i := int64(0); i < 60; i++ {
		event := ev(i, "num", float64(i)+0.5)
		event.Unit = "u"
		if i%2 == 1 {
			event = ev(i, "txt", "a")
		}
		at := 1000000000 + i*1000000000
		event.ObservedTime = model.FormatTime(at + 500000000) // received after it happened
		out = append(out, l.Process(event, at)...)
		if i%15 == 14 {
			out = append(out, l.Flush(at)...)
		}
	}
	out = append(out, l.Flush(1000000000+60*1000000000)...)
	var records []goldenRecord // stays nil (encoded as null) for an empty stream
	for _, d := range out {
		records = append(records, goldenRecord(d))
	}
	raw, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func TestEveryPerturbationStreamIsPinned(t *testing.T) {
	t.Parallel()
	for _, name := range Names {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got, want := goldenStream(t, []string{name}), goldenDigests[name]; got != want {
				t.Fatalf("%s stream digest = %s, want %s", name, got, want)
			}
		})
	}
}

func TestAllPerturbationsTogetherAreOrderedAndPinned(t *testing.T) {
	t.Parallel()
	if got := goldenStream(t, Names); got != goldenAllDigest {
		t.Fatalf("combined stream digest = %s, want %s", got, goldenAllDigest)
	}
}

func TestGoldenDigestsCoverTheCatalog(t *testing.T) {
	t.Parallel()
	if len(goldenDigests) != len(Names) {
		t.Fatalf("%d pinned digests for %d perturbations", len(goldenDigests), len(Names))
	}
}

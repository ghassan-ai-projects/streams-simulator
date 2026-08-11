package randutil

import (
	"math"
	"testing"
)

func TestSplitMix64Deterministic(t *testing.T) {
	a := NewSplitMix64(42)
	b := NewSplitMix64(42)
	for i := 0; i < 100; i++ {
		if a.Next() != b.Next() {
			t.Fatal("same seed diverged")
		}
	}
}

func TestSubstreamStabilityAndIndependence(t *testing.T) {
	s1 := Substream(7, "w/entity/chan/noise")
	first1, second1 := s1.Next(), s1.Next()
	// Same seed+name → same stream.
	s2 := Substream(7, "w/entity/chan/noise")
	if s2.Next() != first1 || s2.Next() != second1 {
		t.Fatal("substream not reproducible")
	}
	// Different names → different streams.
	s3 := Substream(7, "w/entity/chan/delay")
	if s3.Next() == first1 {
		t.Fatal("different substreams must differ")
	}
	// Interleaved creation and drawing of other substreams must not change an
	// existing stream: a fresh instance of the same name replays the same
	// sequence regardless of what else happened in between.
	_ = s3.Next()
	_ = Substream(7, "w/other/chan/noise").Next()
	_ = s3.Next()
	s5 := Substream(7, "w/entity/chan/noise")
	if s5.Next() != first1 || s5.Next() != second1 {
		t.Fatal("interleaved substream creation changed an existing stream")
	}
}

func TestFloat64Range(t *testing.T) {
	r := NewSplitMix64(1)
	for i := 0; i < 10000; i++ {
		f := r.Float64()
		if f < 0 || f >= 1 {
			t.Fatalf("Float64 out of range: %v", f)
		}
	}
}

func TestNormDistribution(t *testing.T) {
	r := NewSplitMix64(9)
	var sum, sumsq float64
	n := 20000
	for i := 0; i < n; i++ {
		v := r.Norm()
		sum += v
		sumsq += v * v
	}
	mean := sum / float64(n)
	variance := sumsq/float64(n) - mean*mean
	if math.Abs(mean) > 0.05 {
		t.Fatalf("mean drifted: %v", mean)
	}
	if math.Abs(variance-1) > 0.05 {
		t.Fatalf("variance drifted: %v", variance)
	}
}

func TestPicker(t *testing.T) {
	r := NewSplitMix64(3)
	p := NewPicker(r, map[string]float64{"a": 0.7, "b": 0.3})
	counts := map[string]int{}
	for i := 0; i < 10000; i++ {
		counts[p.Pick()]++
	}
	if counts["a"] < 6000 || counts["a"] > 8000 {
		t.Fatalf("picker distribution off: %v", counts)
	}
	// Empty picker returns "".
	empty := NewPicker(r, nil)
	if empty.Pick() != "" {
		t.Fatal("empty picker must return empty string")
	}
}

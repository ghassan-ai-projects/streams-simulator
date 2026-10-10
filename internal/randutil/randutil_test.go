package randutil

import (
	"math"
	"testing"
)

func TestSplitMix64Deterministic(t *testing.T) {
	t.Parallel()
	a := NewSplitMix64(42)
	b := NewSplitMix64(42)
	for i := 0; i < 100; i++ {
		if a.Next() != b.Next() {
			t.Fatal("same seed diverged")
		}
	}
}

func TestSubstreamStabilityAndIndependence(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	r := NewSplitMix64(1)
	for i := 0; i < 10000; i++ {
		f := r.Float64()
		if f < 0 || f >= 1 {
			t.Fatalf("Float64 out of range: %v", f)
		}
	}
}

func TestNormDistribution(t *testing.T) {
	t.Parallel()
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

func TestIntnCoversTheRangeAndRejectsEmptyRange(t *testing.T) {
	t.Parallel()
	r := NewSplitMix64(11)
	seen := map[int]bool{}
	for i := 0; i < 1000; i++ {
		v := r.Intn(7)
		if v < 0 || v >= 7 {
			t.Fatalf("Intn(7) = %d", v)
		}
		seen[v] = true
	}
	if len(seen) != 7 {
		t.Fatalf("Intn(7) produced %d distinct values, want 7", len(seen))
	}
	if r.Intn(1) != 0 {
		t.Fatal("Intn(1) must be 0")
	}
	for _, n := range []int{0, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("Intn(%d) must panic", n)
				}
			}()
			r.Intn(n)
		}()
	}
}

func TestExpMeanAndZeroMean(t *testing.T) {
	t.Parallel()
	r := NewSplitMix64(5)
	if r.Exp(0) != 0 || r.Exp(-3) != 0 {
		t.Fatal("non-positive mean must return 0")
	}
	sum := 0.0
	const n = 20000
	for i := 0; i < n; i++ {
		v := r.Exp(2)
		if v < 0 {
			t.Fatalf("Exp returned %v", v)
		}
		sum += v
	}
	if mean := sum / n; math.Abs(mean-2) > 0.1 {
		t.Fatalf("Exp mean drifted: %v", mean)
	}
}

func TestLognormalIsPositiveWithTheRequestedMedianAndSpread(t *testing.T) {
	t.Parallel()
	r := NewSplitMix64(9)
	belowMedian, belowOneSigma := 0, 0
	const n = 20000
	for i := 0; i < n; i++ {
		v := r.Lognormal(1, 0.5)
		if v <= 0 {
			t.Fatalf("Lognormal returned %v", v)
		}
		if v < math.E {
			belowMedian++
		}
		if v < math.Exp(1.5) {
			belowOneSigma++
		}
	}
	if frac := float64(belowMedian) / n; math.Abs(frac-0.5) > 0.02 {
		t.Fatalf("fraction below the median e^mu = %v", frac)
	}
	if frac := float64(belowOneSigma) / n; math.Abs(frac-0.8413) > 0.02 {
		t.Fatalf("fraction below e^(mu+sigma) = %v, want about 0.841", frac)
	}
}

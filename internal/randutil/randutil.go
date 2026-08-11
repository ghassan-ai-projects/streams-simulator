// Package randutil provides the simulator's deterministic PRNG tree.
//
// A single global generator is the classic determinism bug: output depends on
// the order callers happen to draw, so adding one entity shifts every other
// entity's noise. Instead every consumer draws from a named substream:
//
//	substream(name) = splitmix64( seed XOR fnv1a64(name) )
//	name = "<world_id>/<entity_id>/<channel>/<purpose>"
//
// Substreams are independent of each other, created lazily and never shared,
// so enabling duplicate injection (perturb) never changes sensor noise
// (noise). The purpose component keeps the injection surfaces separated.
package randutil

import (
	"hash/fnv"
	"math"
)

// SplitMix64 is a 64-bit deterministic PRNG (the splitmix64 construction).
// It is fast, seedable from a single uint64, and has good output
// distribution for simulation noise. Not cryptographically secure.
type SplitMix64 struct {
	state uint64
}

// NewSplitMix64 returns a PRNG with the given seed.
func NewSplitMix64(seed uint64) *SplitMix64 {
	return &SplitMix64{state: seed}
}

// Next returns the next 64-bit value.
func (s *SplitMix64) Next() uint64 {
	s.state += 0x9e3779b97f4a7c15
	z := s.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// Float64 returns a uniform value in [0, 1) with 53 bits of precision.
func (s *SplitMix64) Float64() float64 {
	return float64(s.Next()>>11) * (1.0 / (1 << 53))
}

// Intn returns a uniform value in [0, n). Panics if n <= 0.
func (s *SplitMix64) Intn(n int) int {
	if n <= 0 {
		panic("randutil: Intn called with n <= 0")
	}
	// #nosec G115 -- n is a positive int, the modulus is < n, and Intn is only called with small n.
	return int(s.Next() % uint64(n))
}

// Norm returns a standard normal sample via Box-Muller.
func (s *SplitMix64) Norm() float64 {
	u1 := s.Float64()
	if u1 < 1e-300 {
		u1 = 1e-300
	}
	u2 := s.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

// Exp returns an exponential sample with the given mean (>= 0).
// A zero mean always returns 0.
func (s *SplitMix64) Exp(mean float64) float64 {
	if mean <= 0 {
		return 0
	}
	u := s.Float64()
	if u < 1e-300 {
		u = 1e-300
	}
	return -mean * math.Log(u)
}

// Lognormal returns a lognormal sample with the given mean and sigma
// (parameters of the underlying normal distribution).
func (s *SplitMix64) Lognormal(mean, sigma float64) float64 {
	return math.Exp(mean + sigma*s.Norm())
}

// Fnv1a64 is the FNV-1a 64-bit hash of name. Deterministic across runs and
// platforms.
func Fnv1a64(name string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return h.Sum64()
}

// Substream returns a fresh PRNG for the named substream under seed.
func Substream(seed uint64, name string) *SplitMix64 {
	return NewSplitMix64(seed ^ Fnv1a64(name))
}

// Picker draws from a weighted discrete distribution.
type Picker struct {
	names   []string
	weights []float64
	total   float64
	rng     *SplitMix64
}

// NewPicker builds a weighted picker over the given name/weight pairs.
// Names with zero weight are skipped; an empty picker returns "" forever.
func NewPicker(rng *SplitMix64, weights map[string]float64) *Picker {
	p := &Picker{rng: rng}
	// determinism-safe: the name/weight pairs are insertion-ordered by the
	// sort below before any pick is made.
	for name, w := range weights {
		if w <= 0 {
			continue
		}
		p.names = append(p.names, name)
		p.weights = append(p.weights, w)
		p.total += w
	}
	// Deterministic order so the picker itself is stable.
	for i := 1; i < len(p.names); i++ {
		for j := i; j > 0 && p.names[j] < p.names[j-1]; j-- {
			p.names[j], p.names[j-1] = p.names[j-1], p.names[j]
			p.weights[j], p.weights[j-1] = p.weights[j-1], p.weights[j]
		}
	}
	return p
}

// Pick returns a name drawn from the distribution, or "" if empty.
func (p *Picker) Pick() string {
	if len(p.names) == 0 || p.total <= 0 {
		return ""
	}
	r := p.rng.Float64() * p.total
	for i, w := range p.weights {
		if r < w {
			return p.names[i]
		}
		r -= w
	}
	return p.names[len(p.names)-1]
}

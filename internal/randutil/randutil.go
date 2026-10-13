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

type SplitMix64 struct {
	state uint64
}

func NewSplitMix64(seed uint64) *SplitMix64 {
	return &SplitMix64{state: seed}
}

func (s *SplitMix64) Next() uint64 {
	s.state += 0x9e3779b97f4a7c15
	z := s.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (s *SplitMix64) Float64() float64 {
	return float64(s.Next()>>11) * (1.0 / (1 << 53))
}

func (s *SplitMix64) Intn(n int) int {
	if n <= 0 {
		panic("randutil: Intn called with n <= 0")
	}
	// #nosec G115 -- n is a positive int, the modulus is < n, and Intn is only called with small n.
	return int(s.Next() % uint64(n))
}

func (s *SplitMix64) Norm() float64 {
	u1 := s.Float64()
	if u1 < 1e-300 {
		u1 = 1e-300
	}
	u2 := s.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

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

func (s *SplitMix64) Lognormal(mean, sigma float64) float64 {
	return math.Exp(mean + sigma*s.Norm())
}

func Fnv1a64(name string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return h.Sum64()
}

func Substream(seed uint64, name string) *SplitMix64 {
	return NewSplitMix64(seed ^ Fnv1a64(name))
}

package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"math"
)

func sampleLinkDelay(delay *model.LinkDelay, rng *randutil.SplitMix64) float64 {
	switch delay.Model {
	case "constant":
		return delay.MeanS
	case "exponential":
		return exponentialLinkDelay(delay, rng)
	case "lognormal":
		return lognormalLinkDelay(delay, rng)
	case "store_and_forward":
		return delay.MeanS + rng.Exp(delay.MeanS)
	}
	return 0
}

func exponentialLinkDelay(delay *model.LinkDelay, rng *randutil.SplitMix64) float64 {
	value := rng.Exp(delay.MeanS)
	if delay.MaxS > 0 && value > delay.MaxS {
		value = delay.MaxS
	}
	return value
}

func lognormalLinkDelay(delay *model.LinkDelay, rng *randutil.SplitMix64) float64 {
	sigma := delay.SigmaS
	if sigma <= 0 {
		sigma = delay.MeanS / 2
	}
	mu := math.Log(delay.MeanS) - sigma*sigma/2
	return rng.Lognormal(mu, sigma)
}

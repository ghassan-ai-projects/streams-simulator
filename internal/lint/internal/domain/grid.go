package domain

import "fmt"

const (
	secondNS   int64 = 1_000_000_000
	sampleStep int64 = 60 * secondNS
	hourNS     int64 = 3600 * secondNS
)

type grid struct {
	startNS int64
	endNS   int64
	timesNS []int64
}

func newGrid(startNS int64, hours float64) (*grid, error) {
	if !(hours > 0 && hours <= maxHours) {
		return nil, fmt.Errorf("%w, got %v", ErrInvalidHours, hours)
	}
	g := &grid{startNS: startNS, endNS: startNS + int64(hours*float64(hourNS))}
	for t := startNS; t < g.endNS; t += sampleStep {
		g.timesNS = append(g.timesNS, t)
	}
	g.timesNS = append(g.timesNS, g.endNS)
	return g, nil
}

func (g *grid) atS(i int) float64 {
	return float64(g.timesNS[i]-g.startNS) / float64(secondNS)
}

func (g *grid) hoursBetween(from, to int) float64 {
	return float64(g.timesNS[to]-g.timesNS[from]) / float64(hourNS)
}

func (g *grid) lastTenthStart() int {
	from := g.endNS - (g.endNS-g.startNS)/10
	for i, t := range g.timesNS {
		if t >= from {
			return i
		}
	}
	return len(g.timesNS) - 1
}

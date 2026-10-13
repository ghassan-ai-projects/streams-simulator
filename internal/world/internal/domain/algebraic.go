package domain

import (
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (w *World) algebraicValue(ent *Entity, dyn *model.Dynamics, at int64) float64 {
	f1 := dyn.F1
	var value float64
	switch f1.Form {
	case model.FormSum:
		value = w.inputValue(ent, dyn, at)
	case model.FormProduct:
		value = w.weightedProduct(ent, f1, at)
	default:
		var axes [2]float64
		value = curveValue(f1.Table, w.weightedInputs(axes[:len(f1.Inputs)], ent, f1, at))
	}
	return clamp(value, f1.Clamp)
}

func (w *World) weightedInputs(values []float64, ent *Entity, f1 *model.F1Dyn, at int64) []float64 {
	for i, in := range f1.Inputs {
		values[i] = in.Coef * w.inputState(ent, in, at)
	}
	return values
}

func (w *World) weightedProduct(ent *Entity, f1 *model.F1Dyn, at int64) float64 {
	result := 1.0
	for _, in := range f1.Inputs {
		result *= in.Coef * w.inputState(ent, in, at)
	}
	return result
}

func curveValue(table *model.Table, inputs []float64) float64 {
	i, fx := segment(table.X, inputs[0])
	if len(inputs) == 1 {
		return lerp(table.Values[i], table.Values[i+1], fx)
	}
	j, fy := segment(table.Y, inputs[1])
	low := lerp(table.Grid[i][j], table.Grid[i][j+1], fy)
	high := lerp(table.Grid[i+1][j], table.Grid[i+1][j+1], fy)
	return lerp(low, high, fx)
}

func segment(xs []float64, x float64) (int, float64) {
	x = min(max(x, xs[0]), xs[len(xs)-1])
	i := min(max(sort.SearchFloat64s(xs, x)-1, 0), len(xs)-2)
	return i, (x - xs[i]) / (xs[i+1] - xs[i])
}

func lerp(a, b, f float64) float64 {
	return a + f*(b-a)
}

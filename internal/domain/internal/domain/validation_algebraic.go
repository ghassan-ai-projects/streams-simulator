package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const maxProductFactors = 4

func (checks domainChecks) f1Factors(f1 *model.F1Dyn) error {
	if n := len(f1.Inputs); n < 2 || n > maxProductFactors {
		return checks.bad("/inputs", "a product multiplies 2 to %d inputs, not %d", maxProductFactors, n)
	}
	return nil
}

func (checks domainChecks) f1Terms(f1 *model.F1Dyn) error {
	if len(f1.Inputs) == 0 {
		return checks.bad("/inputs", "a sum needs at least one input")
	}
	return nil
}

func (checks domainChecks) f1Table(f1 *model.F1Dyn) error {
	table := f1.Table
	if err := checks.increasing("/table/x", table.X); err != nil {
		return err
	}
	switch len(f1.Inputs) {
	case 1:
		return checks.curveValues(table)
	case 2:
		return checks.curveGrid(table)
	}
	return checks.bad("/inputs", "a curve reads one input (values) or two inputs (y and grid), not %d", len(f1.Inputs))
}

func (checks domainChecks) curveValues(table *model.Table) error {
	if len(table.Values) != len(table.X) || table.Y != nil || table.Grid != nil {
		return checks.bad("/table/values", "a one-input curve has one value per x (%d) and no y or grid", len(table.X))
	}
	return nil
}

func (checks domainChecks) curveGrid(table *model.Table) error {
	if err := checks.increasing("/table/y", table.Y); err != nil {
		return err
	}
	if table.Values != nil || len(table.Grid) != len(table.X) {
		return checks.bad("/table/grid", "a two-input curve has one grid row per x (%d) and no values", len(table.X))
	}
	for _, row := range table.Grid {
		if len(row) != len(table.Y) {
			return checks.bad("/table/grid", "every grid row has one value per y (%d)", len(table.Y))
		}
	}
	return nil
}

func (checks domainChecks) increasing(member string, xs []float64) error {
	if len(xs) < 2 {
		return checks.bad(member, "a curve axis needs at least two points")
	}
	for i := 1; i < len(xs); i++ {
		if xs[i] <= xs[i-1] {
			return checks.bad(member, "a curve axis must increase strictly: %v then %v", xs[i-1], xs[i])
		}
	}
	return nil
}

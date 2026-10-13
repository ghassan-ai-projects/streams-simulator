package model

const (
	FormCurve   = "curve"
	FormProduct = "product"
	FormSum     = "sum"
)

type Table struct {
	X      []float64   `json:"x"`
	Values []float64   `json:"values,omitempty"`
	Y      []float64   `json:"y,omitempty"`
	Grid   [][]float64 `json:"grid,omitempty"`
}

func (d *Dynamics) Algebraic() bool {
	if d.F1 == nil {
		return false
	}
	switch d.F1.Form {
	case FormCurve, FormProduct, FormSum:
		return true
	}
	return false
}

func (d *Dynamics) Integrated() bool {
	return d.Tier == "F1" && d.F1 != nil && !d.Algebraic()
}

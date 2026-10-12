package model

const FormStore = "store"

type Flow struct {
	Name        string   `json:"name"`
	From        string   `json:"from"`
	To          string   `json:"to,omitempty"`
	Conductance float64  `json:"conductance"`
	Scale       []string `json:"scale,omitempty"`
}

func (k *Kind) IsStore(state string) bool {
	dyn := k.DynamicsOf(state)
	return dyn != nil && dyn.F1 != nil && dyn.F1.Form == FormStore
}

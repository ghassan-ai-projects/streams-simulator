package model

const (
	ObservationStuck   = "stuck"
	ObservationFreeze  = "freeze"
	ObservationDropout = "dropout"
	ObservationScale   = "scale"
	ObservationNoise   = "noise"
	QualityGood        = "good"
)

type ObservationFault struct {
	Channel string   `json:"channel"`
	Form    string   `json:"form"`
	Value   *float64 `json:"value,omitempty"`
	Factor  *float64 `json:"factor,omitempty"`
	Quality string   `json:"quality,omitempty"`
}

func (o *ObservationFault) Reported() string {
	if o.Quality == QualityGood {
		return ""
	}
	return o.Quality
}

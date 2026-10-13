package perturb

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"
)

func (l *Layer) Apply(name string, params map[string]any, fromNS, untilNS int64) (string, error) {
	return l.layer.Apply(name, params, fromNS, untilNS)
}

func (l *Layer) Clear(id string) error {
	return l.layer.Clear(id)
}

func (l *Layer) Process(ev model.SimEvent, atNS int64) []Delivered {
	return l.layer.Process(ev, atNS)
}

func (l *Layer) Flush(atNS int64) []Delivered {
	return l.layer.Flush(atNS)
}

func CheckParams(name string, params map[string]any) error {
	return layer.CheckParams(name, params)
}

func CheckWindow(name string, fromNS, untilNS int64) error {
	return layer.CheckWindow(name, fromNS, untilNS)
}

func Names() []string {
	return layer.CatalogNames()
}

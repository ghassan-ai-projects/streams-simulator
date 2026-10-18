package component

import (
	"encoding/json"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/files"
)

func NewLibrary(customerDirs []string) (*Library, error) {
	inner, err := files.Load(customerDirs)
	if err != nil {
		return nil, err
	}
	return &Library{library: inner}, nil
}

func Compose(lib *Library, composition []byte, source string) (*Composed, error) {
	if lib == nil || lib.library == nil {
		return nil, ErrNoLibrary
	}
	return layer.ComposeDomain(lib.library, composition, source)
}

func Twin(lib *Library, kinds map[string]map[string]any, choices []byte, source string) (*Composed, error) {
	if lib == nil || lib.library == nil {
		return nil, ErrNoLibrary
	}
	return layer.ScaffoldTwin(lib.library, kinds, choices, source)
}

func Suggest(lib *Library, kinds map[string]map[string]any) (map[string]any, error) {
	if lib == nil || lib.library == nil {
		return nil, ErrNoLibrary
	}
	return layer.SuggestTwinChoices(lib.library, kinds)
}

func TwinRequirements(kind map[string]any) []Issue {
	return layer.TwinRequirements(kind)
}

func (l *Library) List() []Summary {
	if l == nil || l.library == nil {
		return nil
	}
	return l.library.List()
}

func (l *Library) ListFor(unit string, low, high *float64) ([]Summary, error) {
	if l == nil || l.library == nil {
		return nil, nil
	}
	return l.library.ListForReading(unit, low, high)
}

func (l *Library) Describe(ref string) (*Detail, error) {
	if l == nil || l.library == nil {
		return nil, ErrNoLibrary
	}
	return l.library.Describe(ref)
}

func (l *Library) Counts() (equipment, instruments int) {
	if l == nil || l.library == nil {
		return 0, 0
	}
	return l.library.Counts()
}

func CustomerDirectories(flags []string, environment string) []string {
	return files.CustomerDirectories(flags, environment)
}

func Examples() ([]ExampleSummary, error) {
	return files.ListExamples()
}

func Example(id string) (json.RawMessage, error) {
	return files.FindExample(id)
}

func NewComposition(base []byte, plant NewPlant, source string) (map[string]any, error) {
	return layer.NewComposition(base, plant, source)
}

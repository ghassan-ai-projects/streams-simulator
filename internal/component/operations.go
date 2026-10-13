package component

import (
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

func (l *Library) List() []Summary {
	if l == nil || l.library == nil {
		return nil
	}
	return l.library.List()
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
	return layer.CustomerDirectories(flags, environment)
}

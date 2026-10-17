package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
)

func CheckImplements(spec *model.DomainSpec, kindsDirs []string) error {
	sources, err := files.Directories(kindsDirs)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	kindFiles, _, err := domain.ParseKindFiles(sources)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return wrapped(domain.CheckImplements(spec, kindFiles))
}

func ValidateKindFiles(paths []string) ([]domain.KindSummary, error) {
	sources, err := files.Paths(paths)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	_, summaries, err := domain.ParseKindFiles(sources)
	return summaries, wrapped(err)
}

func wrapped(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w", err)
}

package app

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
)

func CheckImplements(spec *model.DomainSpec, kindsDirs []string) ([]model.ImplementsChecked, error) {
	if !spec.HasImplements() {
		return []model.ImplementsChecked{}, nil
	}
	kindFiles, err := readKindFiles(nil, kindsDirs)
	if err != nil {
		return nil, err
	}
	return domain.CheckImplements(spec, kindFiles)
}

func Scaffold[T any](paths, dirs []string, source string, scaffold func(map[string]map[string]any) (T, *model.DomainSpec, error)) (T, error) {
	var none T
	kindFiles, err := readKindFiles(paths, dirs)
	if err != nil {
		return none, err
	}
	built, spec, err := scaffold(kindFiles)
	if err != nil {
		return none, err
	}
	if _, err := domain.CheckImplements(spec, kindFiles); err != nil {
		return none, fmt.Errorf("plantwire: check implements: %w", model.WithSource(err, source))
	}
	return built, nil
}

func ValidateKindFiles(paths []string) ([]domain.KindSummary, error) {
	sources, err := files.Paths(paths)
	if err != nil {
		return nil, err
	}
	_, summaries, err := domain.ParseKindFiles(sources)
	return summaries, err
}

func readKindFiles(paths, dirs []string) (domain.KindFiles, error) {
	byPath, err := files.Paths(paths)
	if err != nil {
		return nil, err
	}
	byDirectory, err := files.Directories(dirs)
	if err != nil {
		return nil, err
	}
	kindFiles, _, err := domain.ParseKindFiles(slices.Concat(byPath, byDirectory))
	return kindFiles, err
}

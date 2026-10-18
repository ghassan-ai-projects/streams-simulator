package app

import (
	"fmt"
	"io"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
)

func CheckImplements(spec *model.DomainSpec, kindsDirs []string) ([]model.ImplementsChecked, error) {
	if !spec.HasImplements() {
		return []model.ImplementsChecked{}, nil
	}
	kindFiles, err := readImplemented(spec, kindsDirs)
	if err != nil {
		return nil, err
	}
	return domain.CheckImplements(spec, kindFiles)
}

func readImplemented(spec *model.DomainSpec, dirs []string) (domain.KindFiles, error) {
	sources, err := files.Scan(dirs)
	if err != nil {
		return nil, err
	}
	kindFiles, _, err := domain.ParseKindFiles(domain.SelectKindSources(sources, domain.ImplementedKinds(spec)))
	return kindFiles, err
}

func WithKindFiles[T any](paths, dirs []string, use func(map[string]map[string]any) (T, error)) (T, error) {
	kindFiles, err := readKindFiles(paths, dirs)
	if err != nil {
		var none T
		return none, err
	}
	return use(kindFiles)
}

func Scaffold[T any](paths, dirs []string, source string, scaffold func(map[string]map[string]any) (T, *model.DomainSpec, error)) (T, error) {
	return WithKindFiles(paths, dirs, func(kindFiles map[string]map[string]any) (T, error) {
		var none T
		built, spec, err := scaffold(kindFiles)
		if err != nil {
			return none, err
		}
		if _, err := domain.CheckImplements(spec, kindFiles); err != nil {
			return none, fmt.Errorf("plantwire: check implements: %w", model.WithSource(err, source))
		}
		return built, nil
	})
}

func ValidateKindFilesWith(paths []string, rules domain.TwinRules) ([]domain.KindSummary, error) {
	sources, err := files.Paths(paths)
	if err != nil {
		return nil, err
	}
	return validateSources(sources, rules)
}

func ValidateKindStream(reader io.Reader, rules domain.TwinRules) (domain.KindSummary, error) {
	source, err := files.Stream(reader)
	if err != nil {
		return domain.KindSummary{}, err
	}
	summaries, err := validateSources([]domain.KindSource{source}, rules)
	if err != nil {
		return domain.KindSummary{}, err
	}
	return summaries[0], nil
}

func validateSources(sources []domain.KindSource, rules domain.TwinRules) ([]domain.KindSummary, error) {
	kindFiles, summaries, err := domain.ParseKindFiles(sources)
	if err != nil {
		return nil, err
	}
	if err := domain.CheckTwinRules(kindFiles, summaries, rules); err != nil {
		return nil, err
	}
	return summaries, nil
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

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
	checked, _, err := loadInterfaces(spec, kindsDirs)
	return checked, err
}

func loadInterfaces(spec *model.DomainSpec, kindsDirs []string) ([]model.ImplementsChecked, domain.Interfaces, error) {
	if !spec.HasImplements() {
		return []model.ImplementsChecked{}, domain.Interfaces{}, nil
	}
	kindFiles, err := readImplemented(spec, kindsDirs)
	if err != nil {
		return nil, nil, err
	}
	boards, err := readBoards(spec, kindsDirs)
	if err != nil {
		return nil, nil, err
	}
	return domain.CheckInterfaces(spec, kindFiles, boards)
}

func boardDigests(checked []model.ImplementsChecked) map[string]string {
	digests := map[string]string{}
	for _, entry := range checked {
		if entry.BoardType != "" {
			digests[entry.BoardType] = entry.BoardDigest
		}
	}
	if len(digests) == 0 {
		return nil
	}
	return digests
}

func readBoards(spec *model.DomainSpec, dirs []string) (domain.BoardFiles, error) {
	wanted := domain.BoardTypes(spec)
	if len(wanted) == 0 {
		return domain.BoardFiles{}, nil
	}
	sources, err := files.ScanBoards(dirs)
	if err != nil {
		return nil, err
	}
	boards, _, err := domain.ParseBoardFiles(domain.SelectBoardSources(sources, wanted))
	return boards, err
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

func ValidateKindFilesWith(paths []string, rules domain.TwinRules) (domain.DeviceFiles, error) {
	sources, err := files.Paths(paths)
	if err != nil {
		return domain.DeviceFiles{}, err
	}
	kindSources, boardSources := domain.SplitDeviceSources(sources)
	_, boards, err := domain.ParseBoardFiles(boardSources)
	if err != nil {
		return domain.DeviceFiles{}, err
	}
	kinds, err := validateSources(kindSources, rules)
	return domain.DeviceFiles{Kinds: kinds, Boards: boards}, err
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

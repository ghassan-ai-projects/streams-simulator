package app

import (
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
)

func CaptureFile(path string, kindsDirs []string) (domain.Capture, error) {
	source, err := files.Capturable(path)
	if err != nil {
		return domain.Capture{}, err
	}
	return capture(source, kindsDirs)
}

func CaptureStream(reader io.Reader, kindsDirs []string) (domain.Capture, error) {
	source, err := files.Stream(reader)
	if err != nil {
		return domain.Capture{}, err
	}
	return capture(source, kindsDirs)
}

func capture(source domain.KindSource, kindsDirs []string) (domain.Capture, error) {
	captured, err := domain.ParseCapture(source.Raw)
	if err != nil {
		return domain.Capture{}, fmt.Errorf("%s: %w", source.Path, model.WithSource(err, source.Path))
	}
	return compareWithWorkspace(captured, kindsDirs)
}

func compareWithWorkspace(captured domain.Capture, kindsDirs []string) (domain.Capture, error) {
	if len(kindsDirs) == 0 {
		return captured, nil
	}
	sources, err := files.Scan(kindsDirs)
	if err != nil {
		return domain.Capture{}, err
	}
	kindFiles, summaries, err := domain.ParseKindFiles(domain.SelectKindSources(sources, domain.KindNames(captured)))
	if err != nil {
		return domain.Capture{}, err
	}
	return domain.CompareWithFiles(captured, kindFiles, summaries)
}

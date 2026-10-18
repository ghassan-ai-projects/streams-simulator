package app

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
)

func CaptureFile(path string, kindsDirs []string) (domain.Capture, error) {
	file, err := files.OpenCapturable(path)
	if err != nil {
		return domain.Capture{}, err
	}
	defer func() { _ = file.Close() }()
	return capture(file, path, kindsDirs)
}

func CaptureStream(reader io.Reader, kindsDirs []string) (domain.Capture, error) {
	source, err := files.Stream(reader)
	if err != nil {
		return domain.Capture{}, err
	}
	return capture(bytes.NewReader(source.Raw), source.Path, kindsDirs)
}

func capture(reader io.Reader, source string, kindsDirs []string) (domain.Capture, error) {
	captured, err := domain.ParseCapture(reader)
	if err != nil {
		return domain.Capture{}, fmt.Errorf("%s: %w", source, model.WithSource(err, source))
	}
	return compareWithWorkspace(captured, kindsDirs)
}

func compareWithWorkspace(captured domain.Capture, kindsDirs []string) (domain.Capture, error) {
	if len(kindsDirs) == 0 {
		return captured, nil
	}
	sources, err := files.ScanPresent(kindsDirs)
	if err != nil {
		return domain.Capture{}, err
	}
	workspace := domain.LoadWorkspaceKinds(domain.SelectFolded(sources, domain.KindNames(captured)))
	return domain.CompareWithFiles(captured, workspace, kindsDirs[0])
}

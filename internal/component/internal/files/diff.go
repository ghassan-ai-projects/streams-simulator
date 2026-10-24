package files

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

const runFileName = "run.json"

func Diff(a, b string) (*domain.Diff, error) {
	left, err := ReadDiffInput(a)
	if err != nil {
		return nil, err
	}
	right, err := ReadDiffInput(b)
	if err != nil {
		return nil, err
	}
	return domain.DescribeDiff(left, right)
}

func ReadDiffInput(path string) (domain.DiffInput, error) {
	raw, err := kernel.ReadRegular(path, maxRunFile)
	if errors.Is(err, kernel.ErrNotRegular) {
		return readRunFolder(path)
	}
	if err != nil {
		return domain.DiffInput{}, fmt.Errorf("component: read %s: %w", path, err)
	}
	return domain.DiffInput{Raw: raw, Label: path}, nil
}

func readRunFolder(folder string) (domain.DiffInput, error) {
	raw, err := kernel.ReadRegular(filepath.Join(folder, runFileName), maxRunFile)
	if err != nil {
		return domain.DiffInput{}, fmt.Errorf("component: read %s: %w", filepath.Join(folder, runFileName), err)
	}
	return domain.DiffInput{Raw: raw, Label: folder, Run: true}, nil
}

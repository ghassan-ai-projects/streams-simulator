package files

import (
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func InstallExampleKinds(id, kindsDir string) (domain.KindFilesInstalled, error) {
	example, err := FindExample(id)
	if err != nil {
		return domain.KindFilesInstalled{}, err
	}
	present, err := presentKindFiles(kindsDir, example.KindFiles)
	if err != nil {
		return domain.KindFilesInstalled{}, err
	}
	installed, writes, err := domain.PlanKindFileInstall(id, example.KindFiles, present)
	if err != nil {
		return domain.KindFilesInstalled{}, err
	}
	return installed, writeKindFiles(kindsDir, writes)
}

func presentKindFiles(kindsDir string, files []domain.KindFile) (map[string][]byte, error) {
	present := map[string][]byte{}
	for _, file := range files {
		raw, found, err := kernel.ReadIfPresent(filepath.Join(kindsDir, file.Name))
		if err != nil {
			return nil, fmt.Errorf("component: read %s: %w", file.Name, err)
		}
		if found {
			present[file.Name] = raw
		}
	}
	return present, nil
}

func writeKindFiles(kindsDir string, files []domain.KindFile) error {
	if len(files) == 0 {
		return nil
	}
	if err := kernel.PrivateDir(kindsDir); err != nil {
		return fmt.Errorf("component: kinds folder: %w", err)
	}
	for _, file := range files {
		if err := kernel.WriteAtomic(filepath.Join(kindsDir, file.Name), file.Document); err != nil {
			return fmt.Errorf("component: write %s: %w", file.Name, err)
		}
	}
	return nil
}

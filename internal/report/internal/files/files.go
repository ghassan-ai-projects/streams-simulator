// Package files is the file-system edge of the report module: the reads of a run folder.
package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
)

func Read(path string) (domain.Document, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return domain.Document{}, nil
	}
	if err != nil {
		return domain.Document{}, fmt.Errorf("report: %w", err)
	}
	return domain.Document{Raw: raw, Present: true}, nil
}

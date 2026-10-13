// Package files is the file-system edge of the report module: the reads of a run folder.
package files

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
)

func Read(path string) (domain.Document, error) {
	raw, present, err := kernel.ReadIfPresent(path)
	if err != nil {
		return domain.Document{}, fmt.Errorf("report: %w", err)
	}
	return domain.Document{Raw: raw, Present: present}, nil
}

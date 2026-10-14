package domain

import (
	"errors"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func ValidateArtifactDocumentForVerdict(raw []byte, path string) error {
	err := ValidateArtifactDocument(raw, path)
	var invalid *model.ValidationError
	if errors.As(err, &invalid) && onlyMissingEvidenceDigests(invalid.Issues) {
		return nil
	}
	return err
}

func onlyMissingEvidenceDigests(issues []model.Issue) bool {
	absent := map[string]bool{}
	for _, digest := range new(model.RunArtifact).EvidenceDigests() {
		absent["/"+digest.Field] = true
	}
	for _, issue := range issues {
		if !absent[issue.Pointer] || !strings.HasPrefix(issue.Message, "missing required property") {
			return false
		}
	}
	return len(issues) > 0
}

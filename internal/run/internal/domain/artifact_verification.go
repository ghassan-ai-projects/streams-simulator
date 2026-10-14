package domain

import (
	"errors"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func ValidateArtifactDocumentForVerification(raw []byte, path string) error {
	err := ValidateArtifactDocument(raw, path)
	var invalid *model.ValidationError
	if errors.As(err, &invalid) && onlyUnusableEvidenceDigests(invalid.Issues) {
		return nil
	}
	return err
}

func onlyUnusableEvidenceDigests(issues []model.Issue) bool {
	fields := map[string]bool{}
	for _, digest := range new(model.RunArtifact).EvidenceDigests() {
		fields["/"+digest.Field] = true
	}
	for _, issue := range issues {
		if !fields[issue.Pointer] || !unusableDigestMessage(issue.Message) {
			return false
		}
	}
	return len(issues) > 0
}

func unusableDigestMessage(message string) bool {
	return strings.HasPrefix(message, "missing required property") || strings.HasPrefix(message, "string does not match pattern")
}

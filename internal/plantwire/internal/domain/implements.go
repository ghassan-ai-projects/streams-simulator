package domain

import (
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	ErrKindFileMissing = errors.New("plantwire: no such kind file")
	ErrKindDrift       = errors.New("plantwire: kind differs from its kind file")
)

func CheckImplements(spec *model.DomainSpec, files KindFiles) error {
	var problems []error
	var issues []model.Issue
	for i := range spec.Kinds {
		if err := implementation(&spec.Kinds[i], files); err != nil {
			problems = append(problems, err)
			issues = append(issues, model.LocatedIssue(fmt.Sprintf("/kinds/%d/implements", i), err.Error()))
		}
	}
	return violation(spec.ID, problems, issues)
}

func violation(id string, problems []error, issues []model.Issue) error {
	if len(problems) == 0 {
		return nil
	}
	messages := make([]string, len(issues))
	for i, issue := range issues {
		messages[i] = issue.Message
	}
	summary := fmt.Sprintf("domain: %s does not implement its kind files: %s", id, strings.Join(messages, "; "))
	return &model.ValidationError{Contract: model.ContractDomainSpec, Summary: summary, Issues: issues, Cause: errors.Join(problems...)}
}

func implementation(kind *model.Kind, files KindFiles) error {
	if kind.Implements == "" {
		return nil
	}
	file, found := files[kind.Implements]
	if !found {
		return fmt.Errorf("kind %q implements %q, but no kind file of that name was loaded (%d loaded): %w", kind.Name, kind.Implements, len(files), ErrKindFileMissing)
	}
	return compareWithFile(kind, file)
}

func compareWithFile(kind *model.Kind, file map[string]any) error {
	want, err := canonicalTree(withoutDigest(file))
	if err != nil {
		return err
	}
	got, err := canonicalTree(KindTemplates([]*model.Kind{kind}, safeStopsOf(file))[0])
	if err != nil {
		return err
	}
	return drift(kind, want, got)
}

func drift(kind *model.Kind, want, got any) error {
	path, differs := firstDifference(want, got, "")
	if !differs {
		return nil
	}
	return fmt.Errorf("kind %q differs from its kind file %q at %s: %w", kind.Name, kind.Implements, path, ErrKindDrift)
}

func withoutDigest(file map[string]any) map[string]any {
	body := maps.Clone(file)
	delete(body, "kind_digest")
	return body
}

func safeStopsOf(file map[string]any) []string {
	var names []string
	commands, _ := file["commands"].([]any)
	for _, command := range commands {
		entry, _ := command.(map[string]any)
		if stop, _ := entry["safe_stop"].(bool); stop {
			names = append(names, nameOf(entry))
		}
	}
	return names
}

func canonicalTree(value any) (any, error) {
	raw, err := canonical.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("plantwire: canonical form of a kind: %w", err)
	}
	var tree any
	if err := decode(raw, &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

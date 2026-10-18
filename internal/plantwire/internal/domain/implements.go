package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	ErrKindFileMissing = errors.New("plantwire: no such kind file")
	ErrKindDrift       = errors.New("plantwire: kind differs from its kind file")
)

type findings struct {
	checked  []model.ImplementsChecked
	problems []error
	issues   []model.Issue
}

func CheckImplements(spec *model.DomainSpec, files KindFiles) ([]model.ImplementsChecked, error) {
	found := findings{checked: []model.ImplementsChecked{}}
	for i := range spec.Kinds {
		found.examine(i, &spec.Kinds[i], files)
	}
	return found.outcome(spec.ID)
}

func (f *findings) examine(index int, kind *model.Kind, files KindFiles) {
	if kind.Implements == "" {
		return
	}
	digest, err := implementation(kind, files)
	if err != nil {
		f.problems = append(f.problems, err)
		f.issues = append(f.issues, model.LocatedIssue(fmt.Sprintf("/kinds/%d/implements", index), err.Error()))
		return
	}
	f.checked = append(f.checked, model.ImplementsChecked{Kind: kind.Implements, KindDigest: digest})
}

func (f *findings) outcome(id string) ([]model.ImplementsChecked, error) {
	if len(f.problems) == 0 {
		return f.checked, nil
	}
	messages := make([]string, len(f.issues))
	for i, issue := range f.issues {
		messages[i] = issue.Message
	}
	summary := fmt.Sprintf("domain: %s does not implement its kind files: %s", id, strings.Join(messages, "; "))
	return nil, &model.ValidationError{Contract: model.ContractDomainSpec, Summary: summary, Issues: f.issues, Cause: errors.Join(f.problems...)}
}

func implementation(kind *model.Kind, files KindFiles) (string, error) {
	file, found := files[kind.Implements]
	if !found {
		return "", fmt.Errorf("kind %q implements %q, but no kind file of that name was loaded (%d loaded): %w", kind.Name, kind.Implements, len(files), ErrKindFileMissing)
	}
	if err := compareWithFile(kind, file); err != nil {
		return "", err
	}
	return KindDigest(file)
}

func compareWithFile(kind *model.Kind, file map[string]any) error {
	want, err := canonicalTree(file)
	if err != nil {
		return err
	}
	got, err := canonicalTree(KindTemplates([]*model.Kind{kind})[0])
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

func canonicalTree(value any) (any, error) {
	raw, err := canonical.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotCanonical, err)
	}
	var tree any
	if err := decode(raw, &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

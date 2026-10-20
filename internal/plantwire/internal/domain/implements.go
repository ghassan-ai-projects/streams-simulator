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
	ErrBoardMissing    = errors.New("plantwire: no such board mapping")
)

type findings struct {
	checked  []model.ImplementsChecked
	problems []error
	issues   []model.Issue
	boards   BoardFiles
	faces    Interfaces

	overridden map[string]bool
}

func overriddenKinds(spec *model.DomainSpec) map[string]bool {
	kinds := map[string]bool{}
	for _, set := range spec.Instances {
		if set.Override != nil {
			kinds[set.Kind] = true
		}
	}
	return kinds
}

func CheckImplements(spec *model.DomainSpec, files KindFiles) ([]model.ImplementsChecked, error) {
	checked, _, err := CheckInterfaces(spec, files, nil)
	return checked, err
}

func CheckInterfaces(spec *model.DomainSpec, files KindFiles, boards BoardFiles) ([]model.ImplementsChecked, Interfaces, error) {
	found := findings{checked: []model.ImplementsChecked{}, boards: boards, faces: Interfaces{}, overridden: overriddenKinds(spec)}
	for i := range spec.Kinds {
		found.examine(i, &spec.Kinds[i], files)
	}
	checked, err := found.outcome(spec.ID)
	if err != nil {
		return nil, nil, err
	}
	return checked, found.faces, nil
}

func (f *findings) examine(index int, kind *model.Kind, files KindFiles) {
	if kind.Implements == "" {
		return
	}
	digest, err := f.implementation(kind, files)
	if err != nil {
		f.problems = append(f.problems, err)
		f.issues = append(f.issues, model.RuleIssue(fmt.Sprintf("/kinds/%d/implements", index), implementsKeyword(err), "implements", err.Error()))
		return
	}
	f.checked = append(f.checked, model.ImplementsChecked{Kind: kind.Implements, KindDigest: digest, BoardType: kind.BoardType,
		BoardDigest: f.boards[kind.BoardType].Digest})
}

func implementsKeyword(err error) string {
	switch {
	case errors.Is(err, ErrKindFileMissing), errors.Is(err, ErrBoardMissing):
		return "implements_missing"
	case errors.Is(err, ErrKindDrift):
		return "implements_drift"
	}
	return "implements_failed"
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

func (f *findings) implementation(kind *model.Kind, files KindFiles) (string, error) {
	file, found := files[kind.Implements]
	if !found {
		return "", fmt.Errorf("kind %q implements %q, but no kind file of that name was loaded (%d loaded): %w", kind.Name, kind.Implements, len(files), ErrKindFileMissing)
	}
	projected, err := f.projected(kind, file)
	if err != nil {
		return "", err
	}
	if err := compareWithFile(projected, file); err != nil {
		return "", err
	}
	return KindDigest(file)
}

func (f *findings) projected(kind *model.Kind, file map[string]any) (*model.Kind, error) {
	if kind.BoardType == "" {
		return kind, nil
	}
	mapped, err := f.mappedKind(kind)
	if err != nil {
		return nil, err
	}
	projected, err := ProjectKind(kind, mapped, file)
	if err == nil {
		f.faces[kind.Name] = Interface{Kind: projected, Mapped: mapped}
	}
	return projected, err
}

func (f *findings) mappedKind(kind *model.Kind) (MappedKind, error) {
	board, loaded := f.boards[kind.BoardType]
	mapped, maps := board.Kinds[kind.Implements]
	switch {
	case f.overridden[kind.Name]:
		return MappedKind{}, overriddenMapping(kind)
	case !loaded:
		return MappedKind{}, fmt.Errorf("kind %q implements %q through the board mapping %q, but no board mapping of that type was loaded (%d loaded): %w",
			kind.Name, kind.Implements, kind.BoardType, len(f.boards), ErrBoardMissing)
	case !maps:
		return MappedKind{}, fmt.Errorf("kind %q implements %q through the board mapping %q, which does not map that kind: %w",
			kind.Name, kind.Implements, kind.BoardType, ErrKindDrift)
	}
	return mapped, nil
}

func overriddenMapping(kind *model.Kind) error {
	return fmt.Errorf("kind %q implements %q through the board mapping %q, and an instance set overrides it: the wire would announce the base kind: %w",
		kind.Name, kind.Implements, kind.BoardType, ErrKindDrift)
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

package domain

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	CompareNew        = "new"
	CompareSame       = "same"
	CompareDiffers    = "differs"
	CompareUnreadable = "unreadable"
)

type KindComparison struct {
	Status       string `json:"status"`
	Target       string `json:"target"`
	File         string `json:"file,omitempty"`
	FileDigest   string `json:"file_digest,omitempty"`
	DiffersAt    string `json:"differs_at,omitempty"`
	CollidesWith string `json:"collides_with,omitempty"`
}

func KindNames(capture Capture) []string {
	names := make([]string, len(capture.Kinds))
	for i, kind := range capture.Kinds {
		names[i] = kind.Kind
	}
	return names
}

func CompareWithFiles(capture Capture, workspace []WorkspaceKind, home string) (Capture, error) {
	compared := slices.Clone(capture.Kinds)
	for i := range compared {
		comparison, issues, err := compareKind(compared[i], workspace, home)
		if err != nil {
			return Capture{}, err
		}
		compared[i].Comparison = &comparison
		compared[i].Issues = append(slices.Clone(compared[i].Issues), issues...)
	}
	capture.Kinds = compared
	return capture, nil
}

func compareKind(kind CapturedKind, workspace []WorkspaceKind, home string) (KindComparison, []model.Issue, error) {
	exact, collisions := matching(kind.Kind, workspace)
	comparison, issues, err := decide(kind, exact, collisions, home)
	return comparison, append(collisionIssues(kind.Kind, collisions), issues...), err
}

func matching(name string, workspace []WorkspaceKind) (exact, collisions []WorkspaceKind) {
	for _, file := range workspace {
		switch {
		case file.Name == name:
			exact = append(exact, file)
		case foldKind(file.Name) == foldKind(name):
			collisions = append(collisions, file)
		}
	}
	return exact, collisions
}

func decide(kind CapturedKind, exact, collisions []WorkspaceKind, home string) (KindComparison, []model.Issue, error) {
	switch {
	case len(exact) > 1:
		return unreadable(exact[0], fmt.Errorf("%w: %q is defined by %s and %s", ErrKindFileDuplicate, clip(kind.Kind), exact[0].Path, exact[1].Path))
	case len(exact) == 1 && exact[0].Err != nil:
		return unreadable(exact[0], exact[0].Err)
	case len(exact) == 1:
		return compareFile(kind, exact[0])
	case len(collisions) > 0:
		return KindComparison{Status: CompareDiffers, Target: collisions[0].Path, File: collisions[0].Path, FileDigest: collisions[0].Digest,
			DiffersAt: "/kind", CollidesWith: collisions[0].Path}, nil, nil
	}
	return KindComparison{Status: CompareNew, Target: filepath.Join(home, KindFileStem(kind.Kind)+KindFileSuffix)}, nil, nil
}

func compareFile(kind CapturedKind, file WorkspaceKind) (KindComparison, []model.Issue, error) {
	comparison := KindComparison{Status: CompareSame, Target: file.Path, File: file.Path, FileDigest: file.Digest}
	if file.Digest == kind.KindDigest {
		return comparison, nil, nil
	}
	comparison.Status = CompareDiffers
	var err error
	comparison.DiffersAt, err = differenceAt(file.Entry, kind.Draft)
	return comparison, nil, err
}

func unreadable(file WorkspaceKind, cause error) (KindComparison, []model.Issue, error) {
	message := fmt.Sprintf("the kind file %s cannot be compared: %v", file.Path, cause)
	return KindComparison{Status: CompareUnreadable, Target: file.Path, File: file.Path},
		[]model.Issue{model.UnlocatedRuleIssue("kind_file_unreadable", message)}, nil
}

func collisionIssues(name string, collisions []WorkspaceKind) []model.Issue {
	var issues []model.Issue
	for _, file := range collisions {
		message := fmt.Sprintf("kind %q collides with the kind file %s, which declares %q: names that differ only in case name one kind", clip(name), file.Path, clip(file.Name))
		issues = append(issues, model.RuleIssue("/kind", "kind_name_collision", "kind", message))
	}
	return issues
}

func differenceAt(file, draft map[string]any) (string, error) {
	want, err := canonicalTree(file)
	if err != nil {
		return "", err
	}
	got, err := canonicalTree(draft)
	if err != nil {
		return "", err
	}
	at, _ := firstDifference(want, got, "")
	return at, nil
}

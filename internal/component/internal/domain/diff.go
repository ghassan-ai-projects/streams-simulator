package domain

import (
	"encoding/json"
	"errors"
	"fmt"

	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func DescribeDiff(a, b DiffInput) (*Diff, error) {
	left, err := sideOf(a)
	if err != nil {
		return nil, err
	}
	right, err := sideOf(b)
	if err != nil {
		return nil, err
	}
	return validatedDiff(compare(left, right))
}

func validatedDiff(diff *Diff) (*Diff, error) {
	raw, err := json.Marshal(diff)
	if err != nil {
		return nil, fmt.Errorf("component: encode domain diff: %w", err)
	}
	if err := model.ValidateDomainDiff(raw); err != nil {
		return nil, fmt.Errorf("component: %w", err)
	}
	return diff, nil
}

func compare(a, b *diffSide) *Diff {
	var tally UnchangedSummary
	diff := &Diff{SchemaVersion: diffVersion, A: a.header(), B: b.header(), Identity: identityOf(a, b)}
	diff.Kinds = append([]KindDiff{}, diffKinds(a, b, &tally)...)
	diff.DomainChanges = append([]Change{}, diffDomain(a, b, &tally)...)
	diff.Summary, diff.Unchanged = summarize(diff), tally
	return diff
}

func identityOf(a, b *diffSide) DiffIdentity {
	same := DiffIdentity{SameID: a.spec.ID == b.spec.ID, SameVersion: a.spec.Version == b.spec.Version, SameDigest: a.compiled.Digest == b.compiled.Digest}
	same.SameIDAndVersionDifferentDigest = same.SameID && same.SameVersion && !same.SameDigest
	return same
}

func summarize(diff *Diff) DiffSummary {
	summary := DiffSummary{Changes: len(diff.DomainChanges)}
	for _, kind := range diff.Kinds {
		summary.Changes += len(kind.Changes)
		switch kind.Status {
		case StatusAdded:
			summary.KindsAdded++
		case StatusRemoved:
			summary.KindsRemoved++
		default:
			summary.KindsChanged++
		}
	}
	return summary
}

type diffSide struct {
	compiled    *specs.Compiled
	spec        *model.DomainSpec
	assumptions *Assumptions
	source      DiffSource
}

func (s *diffSide) header() DiffSide {
	return DiffSide{ID: s.spec.ID, Version: s.spec.Version, Title: s.spec.Title, Digest: s.compiled.Digest, Source: s.source}
}

var (
	ErrRunNoDomainSpec = errors.New("run.json holds no domain_spec")
	ErrRunDomainDigest = errors.New("the domain_spec of run.json does not hash to the digest the run recorded")
)

func sideOf(input DiffInput) (*diffSide, error) {
	raw, recorded, kind := input.Raw, "", SourceFile
	if input.Run {
		var err error
		if raw, recorded, err = runDomain(input); err != nil {
			return nil, err
		}
		kind = SourceRun
	}
	return compileSide(raw, recorded, DiffSource{Kind: kind, Label: input.Label})
}

func compileSide(raw []byte, recorded string, source DiffSource) (*diffSide, error) {
	compiled, err := specs.Parse(raw, source.Label)
	if err != nil {
		return nil, fmt.Errorf("component: %w", err)
	}
	if recorded != "" && compiled.Digest != recorded {
		return nil, runViolation(source.Label, ErrRunDomainDigest)
	}
	return &diffSide{compiled: compiled, spec: compiled.Spec, assumptions: assemble(compiled, nil), source: source}, nil
}

func runDomain(input DiffInput) ([]byte, string, error) {
	if err := model.ValidateRunArtifact(input.Raw); err != nil {
		return nil, "", fmt.Errorf("component: %s: %w", input.Label, err)
	}
	var artifact model.RunArtifact
	if err := json.Unmarshal(input.Raw, &artifact); err != nil || len(artifact.DomainSpec) == 0 {
		return nil, "", runViolation(input.Label, ErrRunNoDomainSpec)
	}
	return artifact.DomainSpec, artifact.Domain.Digest, nil
}

func runViolation(label string, cause error) error {
	message := fmt.Sprintf("component: %s: %s", label, cause)
	return &model.ValidationError{Contract: model.ContractRunArtifact, Source: label, Summary: message,
		Issues: []model.Issue{model.UnlocatedIssue(cause.Error())}, Cause: cause}
}

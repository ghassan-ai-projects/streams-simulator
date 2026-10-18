package domain

import (
	"errors"
	"fmt"
	"maps"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrBirthRefused = errors.New("plantwire: the birth is refused")

type Capture struct {
	Environment string         `json:"environment"`
	BDSeq       uint64         `json:"bd_seq"`
	Kinds       []CapturedKind `json:"kinds"`
}

type CapturedKind struct {
	Kind       string          `json:"kind"`
	KindDigest string          `json:"kind_digest"`
	Draft      map[string]any  `json:"draft"`
	Issues     []model.Issue   `json:"issues,omitempty"`
	Comparison *KindComparison `json:"comparison,omitempty"`
}

func ParseCapture(raw []byte) (Capture, error) {
	birth, err := FindBirth(raw)
	if err != nil {
		return Capture{}, err
	}
	message, err := decodeBirth(birth)
	if err != nil {
		return Capture{}, err
	}
	if err := claimRefusal(message); err != nil {
		return Capture{}, err
	}
	return capturedFrom(message), nil
}

func decodeBirth(birth []byte) (Message, error) {
	record, err := birthRecord(birth)
	if err != nil {
		return Message{}, err
	}
	if errs := schemas()[recordBirth].Validate(record); len(errs) > 0 {
		return Message{}, birthSchemaViolation(errs)
	}
	var message Message
	if err := decode(birth, &message); err != nil {
		return Message{}, birthViolation("NBIRTH does not decode: "+err.Error(), "json", err)
	}
	return message, nil
}

func birthRecord(birth []byte) (map[string]any, error) {
	if len(birth) > MaxBirthBytes {
		return nil, birthViolation(fmt.Sprintf("NBIRTH of %d bytes is larger than %d", len(birth), MaxBirthBytes), "birth_too_large", ErrTooLarge)
	}
	var record map[string]any
	err := decode(birth, &record)
	if err == nil {
		err = refuseDuplicateMembers(birth)
	}
	if err != nil {
		return nil, birthViolation("NBIRTH is not a JSON record: "+err.Error(), "json", err)
	}
	return record, nil
}

func birthSchemaViolation(errs []jsonschema.Error) error {
	return &model.ValidationError{Contract: model.ContractPlantWireBirth, Summary: "plantwire: NBIRTH fails its schema: " + errs[0].Error(),
		Issues: sortedByPointer(model.SchemaIssues(errs)), Cause: errs[0]}
}

func claimRefusal(m Message) error {
	issues := claimIssues(m)
	if len(issues) == 0 {
		return nil
	}
	return &model.ValidationError{Contract: model.ContractPlantWireBirth, Summary: "plantwire: the birth is refused: " + issues[0].Message,
		Issues: issues, Cause: ErrBirthRefused}
}

func claimIssues(m Message) []model.Issue {
	var issues []model.Issue
	if ProtocolInvalid(m) {
		message := fmt.Sprintf("protocol_min %d is above protocol_max %d", m.ProtocolMin, m.ProtocolMax)
		issues = append(issues, model.RuleIssue("/protocol_min", "protocol_invalid", "protocol_min", message))
	}
	if ContractUnsupported(m) {
		issues = append(issues, contractIssue(m))
	}
	return append(issues, kindClaimIssues(m.Kinds)...)
}

func contractIssue(m Message) model.Issue {
	member := "contract_digest"
	if m.ContractVersion != ContractVersion {
		member = "contract_version"
	}
	message := fmt.Sprintf("the node speaks contract %s (%s), this binary %s (%s)", m.ContractVersion, m.ContractDigest, ContractVersion, ContractDigest())
	return model.RuleIssue("/"+member, "contract_unsupported", member, message)
}

func kindClaimIssues(kinds []any) []model.Issue {
	var issues []model.Issue
	seen := map[string]bool{}
	for i, kind := range kinds {
		entry, _ := kind.(map[string]any)
		issues = append(issues, digestIssues(i, entry)...)
		issues = append(issues, duplicateIssues(i, entry, seen)...)
	}
	return issues
}

func digestIssues(index int, entry map[string]any) []model.Issue {
	if !KindMisdigested(entry) {
		return nil
	}
	declared, _ := entry[kindDigestMember].(string)
	computed, err := KindDigest(entry)
	if err != nil {
		computed = "nothing (" + err.Error() + ")"
	}
	message := fmt.Sprintf("kind %q announces kind_digest %q, its content digests to %s: a birth is a claim and this one does not hold", kindNameOf(entry), declared, computed)
	return []model.Issue{model.RuleIssue(fmt.Sprintf("/kinds/%d/%s", index, kindDigestMember), "kind_digest_mismatch", kindDigestMember, message)}
}

func duplicateIssues(index int, entry map[string]any, seen map[string]bool) []model.Issue {
	name := kindNameOf(entry)
	if !seen[name] {
		seen[name] = true
		return nil
	}
	return []model.Issue{model.RuleIssue(fmt.Sprintf("/kinds/%d/kind", index), "kind_duplicate", "kind", fmt.Sprintf("kind %q is announced twice", name))}
}

func capturedFrom(m Message) Capture {
	kinds := make([]CapturedKind, len(m.Kinds))
	for i, kind := range m.Kinds {
		entry, _ := kind.(map[string]any)
		kinds[i] = capturedKind(entry)
	}
	return Capture{Environment: m.Environment, BDSeq: m.BDSeq, Kinds: kinds}
}

func capturedKind(entry map[string]any) CapturedKind {
	draft := maps.Clone(entry)
	delete(draft, kindDigestMember)
	digest, _ := KindDigest(entry)
	return CapturedKind{Kind: kindNameOf(entry), KindDigest: digest, Draft: draft, Issues: draftIssues(draft)}
}

func draftIssues(draft map[string]any) []model.Issue {
	err := CheckKind(draft)
	var invalid *model.ValidationError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &invalid):
		return invalid.Issues
	}
	return []model.Issue{model.UnlocatedRuleIssue("kind_draft", err.Error())}
}

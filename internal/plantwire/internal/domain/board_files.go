package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const BoardFileSuffix = ".board.json"

var (
	ErrBoardTypeDuplicate = errors.New("plantwire: two board mappings of one board type")
	ErrBoardRule          = errors.New("plantwire: a board mapping breaks a rule its schema cannot say")
)

type BoardMapping struct {
	BoardType string                `json:"board_type"`
	Kinds     map[string]MappedKind `json:"kinds"`
	Digest    string                `json:"-"`
}

type MappedKind struct {
	Commands map[string]MappedCommand `json:"commands"`
	Metrics  map[string]MappedMetric  `json:"metrics"`
	RawEnums map[string][]string      `json:"raw_enums,omitempty"`
}

type MappedCommand struct {
	Raw   string                    `json:"raw"`
	Args  map[string]MappedArg      `json:"args,omitempty"`
	By    string                    `json:"by,omitempty"`
	Table map[string]map[string]any `json:"table,omitempty"`
	Lease *MappedLease              `json:"lease,omitempty"`
}

type MappedArg struct {
	Value any         `json:"value,omitempty"`
	From  string      `json:"from,omitempty"`
	Scale json.Number `json:"scale,omitempty"`
}

type MappedMetric struct {
	Raw   string            `json:"raw"`
	Scale json.Number       `json:"scale,omitempty"`
	Table map[string]string `json:"table,omitempty"`
}

type MappedLease struct {
	Arg string `json:"arg"`
	MS  int64  `json:"ms"`
}

type BoardFiles map[string]BoardMapping

type DeviceFiles struct {
	Kinds  []KindSummary  `json:"kinds"`
	Boards []BoardSummary `json:"boards,omitempty"`
}

func SplitDeviceSources(sources []KindSource) (kinds, boards []KindSource) {
	for _, source := range sources {
		if strings.HasSuffix(source.Path, BoardFileSuffix) {
			boards = append(boards, source)
		} else {
			kinds = append(kinds, source)
		}
	}
	return kinds, boards
}

type BoardSummary struct {
	Path        string   `json:"path"`
	BoardType   string   `json:"board_type"`
	BoardDigest string   `json:"board_digest"`
	Kinds       []string `json:"kinds"`
}

func ParseBoardFiles(sources []KindSource) (BoardFiles, []BoardSummary, error) {
	index := boardIndex{files: BoardFiles{}, origin: map[string]string{}}
	summaries := make([]BoardSummary, 0, len(sources))
	for _, source := range sources {
		summary, err := index.parse(source)
		if err != nil {
			return nil, nil, err
		}
		summaries = append(summaries, summary)
	}
	return index.files, summaries, nil
}

type boardIndex struct {
	files  BoardFiles
	origin map[string]string
}

func (i boardIndex) parse(source KindSource) (BoardSummary, error) {
	mapping, err := ParseBoardMapping(source)
	if err != nil {
		return BoardSummary{}, err
	}
	kinds := slices.Sorted(maps.Keys(mapping.Kinds))
	if err := i.addBoard(source.Path, mapping); err != nil {
		return BoardSummary{}, err
	}
	return BoardSummary{Path: source.Path, BoardType: mapping.BoardType, BoardDigest: mapping.Digest, Kinds: kinds}, nil
}

func (i boardIndex) addBoard(path string, mapping BoardMapping) error {
	if first, taken := i.origin[mapping.BoardType]; taken {
		return fmt.Errorf("%w: %q is mapped by %s and %s", ErrBoardTypeDuplicate, mapping.BoardType, first, path)
	}
	i.files[mapping.BoardType], i.origin[mapping.BoardType] = mapping, path
	return nil
}

func ParseBoardMapping(source KindSource) (BoardMapping, error) {
	if source.Err != nil {
		return BoardMapping{}, source.Err
	}
	document, err := decodeKind(source.Raw)
	if err != nil {
		return BoardMapping{}, model.NotJSONError(model.ContractBoardFile, source.Path, "plantwire: board mapping "+source.Path, err)
	}
	mapping, err := checkBoard(source.Raw, document)
	if err != nil {
		return BoardMapping{}, fmt.Errorf("%s: %w", source.Path, model.WithSource(err, source.Path))
	}
	return mapping, nil
}

func checkBoard(raw []byte, document map[string]any) (BoardMapping, error) {
	if errs := schemas()[boardFileSchema].Validate(document); len(errs) > 0 {
		return BoardMapping{}, boardInvalid("fails its schema", model.SchemaIssues(errs), errs[0])
	}
	var mapping BoardMapping
	if err := decode(raw, &mapping); err != nil {
		return BoardMapping{}, model.NotJSONError(model.ContractBoardFile, "", "plantwire: board mapping", err)
	}
	if issues := boardRuleIssues(mapping); len(issues) > 0 {
		return BoardMapping{}, boardInvalid("breaks a rule", issues, ErrBoardRule)
	}
	return withBoardDigest(mapping, document)
}

func withBoardDigest(mapping BoardMapping, document map[string]any) (BoardMapping, error) {
	digest, err := canonical.Digest(document)
	if err != nil {
		cause := fmt.Errorf("%w: %w", ErrNotCanonical, err)
		return BoardMapping{}, boardInvalid("has no canonical form", []model.Issue{model.UnlocatedRuleIssue("canonical", err.Error())}, cause)
	}
	mapping.Digest = digest
	return mapping, nil
}

func boardInvalid(what string, issues []model.Issue, cause error) error {
	issues = sortedByPointer(issues)
	summary := fmt.Sprintf("plantwire: board mapping %s: %s: %s", what, issues[0].Pointer, issues[0].Message)
	return &model.ValidationError{Contract: model.ContractBoardFile, Summary: summary, Issues: issues, Cause: cause}
}

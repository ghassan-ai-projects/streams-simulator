package domain

import (
	"maps"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	KeywordBoardNameInvalid      = "board_name_invalid"
	KeywordBoardRawWordUnknown   = "board_raw_word_unknown"
	KeywordBoardArgumentConflict = "board_argument_conflict"
	KeywordBoardTableUneven      = "board_table_uneven"
	KeywordBoardScaleMismatch    = "board_scale_mismatch"
)

var pointerEscapes = strings.NewReplacer("~", "~0", "/", "~1")

func boardRuleIssues(mapping BoardMapping) []model.Issue {
	var issues []model.Issue
	for _, name := range sortedKeys(mapping.Kinds) {
		at := "/kinds/" + pointerEscapes.Replace(name)
		issues = append(issues, nameIssues(at, name)...)
		issues = append(issues, mappedKindIssues(at, mapping.Kinds[name])...)
	}
	return issues
}

func mappedKindIssues(at string, kind MappedKind) []model.Issue {
	var issues []model.Issue
	for _, name := range sortedKeys(kind.RawEnums) {
		issues = append(issues, nameIssues(at+"/raw_enums/"+pointerEscapes.Replace(name), name)...)
	}
	issues = append(issues, mappedCommandsIssues(at, kind)...)
	for _, name := range sortedKeys(kind.Metrics) {
		pointer := at + "/metrics/" + pointerEscapes.Replace(name)
		issues = append(issues, nameIssues(pointer, name)...)
		issues = append(issues, mappedMetricIssues(pointer, name, kind)...)
	}
	return issues
}

func mappedCommandsIssues(at string, kind MappedKind) []model.Issue {
	var issues []model.Issue
	for _, name := range sortedKeys(kind.Commands) {
		pointer := at + "/commands/" + pointerEscapes.Replace(name)
		issues = append(issues, nameIssues(pointer, name)...)
		issues = append(issues, mappedCommandIssues(pointer, kind.Commands[name], kind.RawEnums)...)
	}
	return issues
}

func mappedCommandIssues(at string, command MappedCommand, rawEnums map[string][]string) []model.Issue {
	var issues []model.Issue
	for _, name := range sortedKeys(command.Args) {
		pointer := at + "/args/" + pointerEscapes.Replace(name)
		issues = append(issues, nameIssues(pointer, name)...)
		issues = append(issues, rawWordIssues(pointer+"/value", rawEnums[name], command.Args[name].Value)...)
	}
	issues = append(issues, tableIssues(at, command, rawEnums)...)
	return append(issues, leaseIssues(at, command)...)
}

func tableIssues(at string, command MappedCommand, rawEnums map[string][]string) []model.Issue {
	words := sortedKeys(command.Table)
	if len(words) == 0 {
		return nil
	}
	first := sortedKeys(command.Table[words[0]])
	var issues []model.Issue
	for _, word := range words {
		pointer := at + "/table/" + pointerEscapes.Replace(word)
		issues = append(issues, nameIssues(pointer, word)...)
		issues = append(issues, unevenRowIssues(pointer, word, command.Table[word], first)...)
		issues = append(issues, rowCellIssues(pointer, command, command.Table[word], rawEnums)...)
	}
	return issues
}

func unevenRowIssues(at, word string, row map[string]any, first []string) []model.Issue {
	if slices.Equal(sortedKeys(row), first) {
		return nil
	}
	message := "every table row sets the same raw arguments, " + strings.Join(first, ", ")
	return []model.Issue{model.RuleIssue(at, KeywordBoardTableUneven, word, message)}
}

func rowCellIssues(at string, command MappedCommand, row map[string]any, rawEnums map[string][]string) []model.Issue {
	var issues []model.Issue
	for _, arg := range sortedKeys(row) {
		pointer := at + "/" + pointerEscapes.Replace(arg)
		issues = append(issues, nameIssues(pointer, arg)...)
		issues = append(issues, rawWordIssues(pointer, rawEnums[arg], row[arg])...)
		if _, set := command.Args[arg]; set {
			issues = append(issues, argumentConflict(pointer, arg))
		}
	}
	return issues
}

func leaseIssues(at string, command MappedCommand) []model.Issue {
	if command.Lease == nil {
		return nil
	}
	arg := command.Lease.Arg
	_, inArgs := command.Args[arg]
	inTable := slices.ContainsFunc(slices.Collect(maps.Values(command.Table)), func(row map[string]any) bool {
		_, set := row[arg]
		return set
	})
	if !inArgs && !inTable {
		return nil
	}
	return []model.Issue{argumentConflict(at+"/lease/arg", arg)}
}

func argumentConflict(at, arg string) model.Issue {
	return model.RuleIssue(at, KeywordBoardArgumentConflict, arg, "the raw argument "+arg+" is set more than once by args, a table row and the lease")
}

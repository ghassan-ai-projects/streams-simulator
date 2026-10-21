package contract

import (
	"path/filepath"
	"slices"
	"testing"
)

var identityMembers = map[string][]string{
	"device":     {"device"},
	"command_id": {"command_id"},
	"boot_id":    {"boot_id", "expected_boot_id"},
}

func TestEveryExpectationNamesADeviceCommandAndBootItsExchangeCarries(t *testing.T) {
	t.Parallel()
	for _, file := range fixtures(t, "exchanges", "*.jsonl") {
		_, steps := exchangeSteps(t, file)
		carried := carriedIdentities(steps)
		for _, step := range steps {
			for field, value := range expectedIdentities(step) {
				if !slices.Contains(carried[field], value) {
					t.Errorf("%s: an expectation names %s %v, which no record of the exchange carries", filepath.Base(file), field, value)
				}
			}
		}
	}
}

func carriedIdentities(steps []any) map[string][]any {
	carried := map[string][]any{}
	for _, step := range steps {
		record := member(step, "record")
		for field, members := range identityMembers {
			for _, name := range members {
				if value := member(record, name); value != nil {
					carried[field] = append(carried[field], value)
				}
			}
		}
	}
	return carried
}

func expectedIdentities(step any) map[string]any {
	expected := map[string]any{}
	fields, _ := member(step, "expect", "fields").(map[string]any)
	for field := range identityMembers {
		if value, named := fields[field]; named {
			expected[field] = value
		}
	}
	return expected
}

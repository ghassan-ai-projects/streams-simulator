package app

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
)

func TestAdapterVerdictFailsOnlyForADeclaredCheckOrForNothingToCheck(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		res     adapter.VerifyResult
		wantErr string
	}{
		{"nothing declared", adapter.VerifyResult{Adapter: "a"}, "declares no conformance schema or golden"},
		{"schema only, ok", adapter.VerifyResult{Adapter: "a", SchemaChecked: true, SchemaOK: true}, ""},
		{"golden only, ok", adapter.VerifyResult{Adapter: "a", GoldenChecked: true, GoldenMatch: true}, ""},
		{"schema failed", adapter.VerifyResult{Adapter: "a", SchemaChecked: true, FirstDivergence: "record 0 fails"}, "adapter verify FAILED: record 0 fails"},
		{"golden failed", adapter.VerifyResult{Adapter: "a", SchemaChecked: true, SchemaOK: true, GoldenChecked: true, FirstDivergence: "lengths differ"}, "adapter verify FAILED: lengths differ"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := adapterVerdict(&tc.res)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("err = %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("err = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

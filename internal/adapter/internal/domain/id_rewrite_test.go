package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestHashSuffixKeepsTheLimitAndStaysHexWhateverTheLimit(t *testing.T) {
	t.Parallel()
	long := "site-a/line-12/asset-0000000000000000000000000001"
	other := "site-a/line-12/asset-0000000000000000000000000002"
	for _, limit := range []int{1, 4, 15, 16, 24, len(long) - 1} {
		got, err := rewriteOverflow(long, long, &model.IDRewrite{MaxLength: limit, OnViolation: "hash_suffix"})
		if err != nil {
			t.Fatalf("limit %d: %v", limit, err)
		}
		if len(got) != limit {
			t.Fatalf("limit %d: id %q has length %d", limit, got, len(got))
		}
		suffix := got[len(got)-min(16, limit):]
		if strings.Trim(suffix, "0123456789abcdef") != "" || strings.Contains(got, ":") {
			t.Fatalf("limit %d: suffix %q is not plain hex in %q", limit, suffix, got)
		}
	}
	a, _ := rewriteOverflow(long, long, &model.IDRewrite{MaxLength: 24, OnViolation: "hash_suffix"})
	b, _ := rewriteOverflow(other, other, &model.IDRewrite{MaxLength: 24, OnViolation: "hash_suffix"})
	if a == b {
		t.Fatalf("distinct ids must keep distinct hashed forms: %q", a)
	}
}

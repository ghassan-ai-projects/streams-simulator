package capability

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNewTokenIsPrefixedFixedWidthAndUnique(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for range 16 {
		token, err := NewToken()
		if err != nil {
			t.Fatalf("new token: %v", err)
		}
		raw, ok := strings.CutPrefix(token, "t-")
		if !ok {
			t.Fatalf("token %q lacks the t- prefix", token)
		}
		decoded, err := base64.RawURLEncoding.DecodeString(raw)
		if err != nil || len(decoded) != 32 {
			t.Fatalf("token body must be 32 raw-url-base64 bytes: len=%d err=%v", len(decoded), err)
		}
		if seen[token] {
			t.Fatalf("token %q repeated", token)
		}
		seen[token] = true
	}
}

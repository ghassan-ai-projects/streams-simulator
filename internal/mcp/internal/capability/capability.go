// Package capability is the entropy edge of the MCP module: it mints the
// opaque capability tokens that name one world to the operator role.
package capability

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// NewToken returns a fresh unguessable capability token.
func NewToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read capability token randomness: %w", err)
	}
	return "t-" + base64.RawURLEncoding.EncodeToString(buf), nil
}

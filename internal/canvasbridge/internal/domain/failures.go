package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
)

type Failure struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Issues  json.RawMessage `json:"issues,omitempty"`
}

func (f *Failure) Error() string {
	return f.Code + ": " + f.Message
}

func ToolFailure(text string) *Failure {
	code, message, found := strings.Cut(text, ": ")
	if !found || strings.ContainsAny(code, " \n") || code == "" {
		return &Failure{Code: "director_refused", Message: text}
	}
	return &Failure{Code: code, Message: message}
}

func CommandFailure(stdout []byte, exitCode int, stderr string) *Failure {
	var document struct {
		Error *Failure `json:"error"`
	}
	if json.Unmarshal(stdout, &document) == nil && document.Error != nil && document.Error.Code != "" {
		return document.Error
	}
	message := strings.TrimSpace(stderr)
	if message == "" {
		message = "streamsim exited with status " + strconv.Itoa(exitCode)
	}
	return &Failure{Code: "command_failed", Message: message}
}

func Version(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:16])
}

type RunOpening struct {
	Type string `json:"type"`
	Mode string `json:"mode"`
}

func ExploreRun(firstJournalLine []byte) bool {
	var opening RunOpening
	return json.Unmarshal(firstJournalLine, &opening) == nil && opening.Type == "run.opened" && opening.Mode == "explore"
}

type ConflictError struct {
	Path string
	Etag string
}

func (c *ConflictError) Error() string {
	return ErrConflict.Error() + ": " + c.Path
}

func (c *ConflictError) Is(target error) bool {
	return target == ErrConflict
}

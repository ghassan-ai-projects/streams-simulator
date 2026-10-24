package domain

import (
	"errors"
	"regexp"
	"testing"
)

func TestAFailureReadsAsItsCodeAndMessage(t *testing.T) {
	t.Parallel()
	failure := &Failure{Code: "world_unknown", Message: "no such world"}
	var asError error = failure
	if asError.Error() != "world_unknown: no such world" {
		t.Fatalf("Error() = %q", asError.Error())
	}
	var target *Failure
	if !errors.As(asError, &target) || target != failure {
		t.Fatal("a Failure is found again with errors.As")
	}
}

func TestAToolFailureTextIsSplitIntoCodeAndMessageOrKeptAsARefusal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, text, code, message string
	}{
		{"a coded text", "world_unknown: no such world", "world_unknown", "no such world"},
		{"only the first separator splits", "bad_input: expected: a number", "bad_input", "expected: a number"},
		{"no separator", "the director refused", "director_refused", "the director refused"},
		{"a code with a space", "not a code: message", "director_refused", "not a code: message"},
		{"a code with a newline", "line\nbreak: message", "director_refused", "line\nbreak: message"},
		{"an empty code", ": message", "director_refused", ": message"},
		{"an empty text", "", "director_refused", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ToolFailure(tc.text); got.Code != tc.code || got.Message != tc.message {
				t.Fatalf("ToolFailure(%q) = %+v, want %s / %q", tc.text, got, tc.code, tc.message)
			}
		})
	}
}

func TestACommandFailureIsTheErrorDocumentOfTheCommandOrItsStderrOrItsExitStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		stdout   string
		exit     int
		stderr   string
		code     string
		message  string
		wantSame bool
	}{
		{"the error document wins", `{"error":{"code":"invalid_domain","message":"bad","issues":[1]}}`, 2, "ignored", "invalid_domain", "bad", true},
		{"a document with an empty code is not one", `{"error":{"code":"","message":"bad"}}`, 1, " boom \n", "command_failed", "boom", false},
		{"a document without an error", `{"valid":false}`, 1, "boom", "command_failed", "boom", false},
		{"output that is not JSON", `not json`, 3, "", "command_failed", "streamsim exited with status 3", false},
		{"no output at all", ``, 1, "  \n", "command_failed", "streamsim exited with status 1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := CommandFailure([]byte(tc.stdout), tc.exit, tc.stderr)
			if got.Code != tc.code || got.Message != tc.message {
				t.Fatalf("CommandFailure = %+v, want %s / %q", got, tc.code, tc.message)
			}
			if tc.wantSame && string(got.Issues) != "[1]" {
				t.Fatalf("issues = %s, want the document's issues kept", got.Issues)
			}
		})
	}
}

func TestAVersionIsAStableDigestOfTheContent(t *testing.T) {
	t.Parallel()
	hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)
	first, again, other := Version([]byte("{}")), Version([]byte("{}")), Version([]byte("{} "))
	if !hex32.MatchString(first) || first != again || first == other || !hex32.MatchString(Version(nil)) {
		t.Fatalf("versions = %q %q %q, want 32 hex digits, equal for equal content, different otherwise", first, again, other)
	}
}

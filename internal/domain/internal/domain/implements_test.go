package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpWithImplements(t *testing.T, value any) []byte {
	t.Helper()
	return documentOf(t, testsupport.PumpBoardDomain(), func(doc map[string]any) { kindAt(doc, 0)["implements"] = value })
}

func TestAKindNamesTheKindFileItImplementsAndFormatKeepsIt(t *testing.T) {
	t.Parallel()
	c, err := Parse(pumpWithImplements(t, "pump"), "implements.json")
	if err != nil || c.Spec.Kinds[0].Implements != "pump" {
		t.Fatalf("Parse = %v, %v, want the kind to implement pump", c, err)
	}
	formatted, err := Format(c)
	if err != nil {
		t.Fatal(err)
	}
	text := string(formatted)
	if strings.Index(text, `"name": "pump"`) >= strings.Index(text, `"implements": "pump"`) ||
		strings.Index(text, `"implements": "pump"`) >= strings.Index(text, `"state"`) {
		t.Fatalf("Format puts implements after the kind's name and before its state, got:\n%s", text)
	}
	again, err := Parse(formatted, "formatted.json")
	if err != nil || again.Digest != c.Digest {
		t.Fatalf("Parse(Format) = %v, %v, want the digest %s kept", again, err, c.Digest)
	}
}

func TestAnImplementsThatIsNotAKindFileNameIsASchemaViolationAtTheKind(t *testing.T) {
	t.Parallel()
	for name, value := range map[string]any{"empty": "", "a space": "pump board", "a number": 3, "too long": strings.Repeat("k", 129)} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(pumpWithImplements(t, value), "implements.json")
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || len(invalid.Issues) == 0 || invalid.Issues[0].Pointer != "/kinds/0/implements" {
				t.Fatalf("Parse = %v, want a schema issue at /kinds/0/implements", err)
			}
		})
	}
}

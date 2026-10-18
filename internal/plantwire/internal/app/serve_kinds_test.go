package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func implementingPumpRun(t *testing.T) *run.Run {
	t.Helper()
	r := runOf(t, testsupport.PumpBoardDomain())
	r.Domain().Spec.Kinds[0].Implements = "pump"
	return r
}

func TestServeChecksTheKindFilesBeforeItDials(t *testing.T) {
	t.Parallel()
	matching := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	drifted := kindsDirectoryWith(t, map[string]string{"pump.kind.json": strings.Replace(pumpKindContent(t), `"%"`, `"pct"`, 1)})
	cases := []struct {
		name string
		dirs []string
		want error
	}{
		{"no kind file for a domain that implements one", nil, wire.ErrKindFileMissing},
		{"a kind file the domain drifted from", []string{drifted}, wire.ErrKindDrift},
		{"a kind file that matches, which leaves the dial to fail", []string{matching}, fs.ErrNotExist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			options := sliceOptions(1)
			options.KindsDirs = tc.dirs
			err := Serve(t.Context(), implementingPumpRun(t), "/nonexistent/host.sock", time.Second, options)
			var invalid *model.ValidationError
			if !errors.Is(err, tc.want) {
				t.Fatalf("Serve = %v, want %v", err, tc.want)
			}
			if !errors.Is(tc.want, fs.ErrNotExist) && (!errors.As(err, &invalid) || invalid.Source != "pump-board") {
				t.Fatalf("Serve = %v, want a validation error whose source is the domain", err)
			}
		})
	}
}

func TestANodeWhoseKindHasNoCanonicalFormIsRefusedAtItsBirth(t *testing.T) {
	t.Parallel()
	r := runOf(t, testsupport.PumpBoardDomain())
	r.Domain().Spec.Kinds[0].Effectors[0].ArgsSchema = map[string]any{"type": "object", "maximum": json.Number("1e999")}
	err := NewSession(r, &scriptedHost{t: t}, sliceOptions(1)).Run(t.Context())
	if !errors.Is(err, wire.ErrNotCanonical) {
		t.Fatalf("Run = %v, want ErrNotCanonical from the node birth", err)
	}
}

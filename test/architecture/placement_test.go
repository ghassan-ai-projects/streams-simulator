package architecture

import (
	"slices"
	"strings"
	"testing"
)

// Go packages live under cmd/ and internal/, plus the repository root and the
// build-tagged tools package. A package anywhere else needs the owner's
// approval (AGENTS.md, MUST 4).

func TestGoPackagesLiveUnderCmdAndInternal(t *testing.T) {
	t.Parallel()
	var directories []string
	for _, file := range productionFiles(t) {
		if !slices.Contains(directories, file.pkgDir) {
			directories = append(directories, file.pkgDir)
		}
	}
	for _, violation := range placementViolations(directories, nil) {
		t.Error(violation)
	}
}

func TestPlacementRejectsPackagesOutsideCmdAndInternal(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		directories []string
		exceptions  []string
		want        int
	}{
		"the root, tools, cmd and internal":  {[]string{".", "tools", "cmd/streamsim", "internal/run", "internal/run/internal/app"}, nil, 0},
		"a top-level package":                {[]string{"pkg/util"}, nil, 1},
		"a package beside the repository":    {[]string{"scripts"}, nil, 1},
		"a listed exception":                 {[]string{"components"}, []string{"components"}, 0},
		"an exception does not cover others": {[]string{"components", "extras"}, []string{"components"}, 1},
		"a prefix is not a directory":        {[]string{"internalx/run", "cmdline"}, nil, 2},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := placementViolations(tc.directories, tc.exceptions); len(got) != tc.want {
				t.Fatalf("placementViolations(%v) = %v, want %d", tc.directories, got, tc.want)
			}
		})
	}
}

func placementViolations(directories, exceptions []string) []string {
	var violations []string
	for _, directory := range directories {
		if !placedCorrectly(directory) && !slices.Contains(exceptions, directory) {
			violations = append(violations, directory+": Go packages live under cmd/ or internal/ (or are the root or tools); a new location needs the owner's approval")
		}
	}
	return violations
}

func placedCorrectly(directory string) bool {
	return directory == "." || directory == "tools" || strings.HasPrefix(directory, "cmd/") || strings.HasPrefix(directory, "internal/")
}
